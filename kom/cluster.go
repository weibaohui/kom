package kom

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/dgraph-io/ristretto/v2"
	openapi_v2 "github.com/google/gnostic-models/openapiv2"
	"github.com/weibaohui/kom/kom/aws"
	"github.com/weibaohui/kom/kom/describe"
	"github.com/weibaohui/kom/kom/doc"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/version"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
)

var clusterInstances *ClusterInstances

// ClusterInstances 集群实例管理器
type ClusterInstances struct {
	clusters             sync.Map                          // map[string]*ClusterInst
	callbackRegisterFunc func(cluster *ClusterInst) func() // 用来注册回调参数的回调方法
}

// ClusterInst 单一集群实例
type ClusterInst struct {
	ID                 string                       // 集群ID
	Kubectl            *Kubectl                     // kom
	Client             *kubernetes.Clientset        // kubernetes 客户端
	Config             *rest.Config                 // rest config
	DynamicClient      *dynamic.DynamicClient       // 动态客户端
	apiResources       []*metav1.APIResource        // 当前k8s已注册资源
	crdList            []*unstructured.Unstructured // 当前k8s已注册资源 //TODO 定时更新或者Watch更新
	callbacks          *callbacks                   // 回调
	docs               *doc.Docs                    // 文档
	serverVersion      *version.Info                // 服务器版本
	describerMap       map[schema.GroupKind]describe.ResourceDescriber
	Cache              *ristretto.Cache[string, any]
	openAPISchema      *openapi_v2.Document // openapi
	watchCRDCancelFunc context.CancelFunc   // CRD取消方法，用于断开连接的时候停止

	// AWS EKS 特定字段
	AWSAuthProvider    *aws.AuthProvider  // AWS 认证提供者
	IsEKS              bool               // 是否为 EKS 集群
	tokenRefreshCancel context.CancelFunc // token 刷新取消函数
}

// Clusters 集群实例管理器
func Clusters() *ClusterInstances {
	return clusterInstances
}

// 初始化
func init() {
	clusterInstances = &ClusterInstances{}
}

// DefaultCluster 获取默认集群，简化调用方式
func DefaultCluster() *Kubectl {
	return Clusters().DefaultCluster().Kubectl
}

// Cluster 获取集群
func Cluster(id string) *Kubectl {
	var cluster *ClusterInst
	if id == "" {
		cluster = Clusters().DefaultCluster()
	} else {
		cluster = Clusters().GetClusterById(id)
	}
	if cluster == nil {
		return nil
	}
	return cluster.Kubectl
}

// RegisterInCluster 注册InCluster集群
func (c *ClusterInstances) RegisterInCluster() (*Kubectl, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("InCluster Error %v", err)
	}
	return c.RegisterByConfigWithID(config, "InCluster")
}

// SetRegisterCallbackFunc 设置回调注册函数
func (c *ClusterInstances) SetRegisterCallbackFunc(callback func(cluster *ClusterInst) func()) {
	c.callbackRegisterFunc = callback
}

// RegisterByPath 通过kubeconfig文件路径注册集群
func (c *ClusterInstances) RegisterByPath(path string) (*Kubectl, error) {
	// 先读取文件内容以检测是否为 EKS 配置
	content, err := clientcmd.LoadFromFile(path)
	if err != nil {
		return nil, fmt.Errorf("RegisterByPath Error loading file %s: %v", path, err)
	}

	// 将配置转换为字节数组进行 EKS 检测
	kubeconfigBytes, err := clientcmd.Write(*content)
	if err != nil {
		return nil, fmt.Errorf("RegisterByPath Error serializing kubeconfig %s: %v", path, err)
	}

	// 自动检测是否为 EKS 配置
	authProvider := aws.NewAuthProvider()
	if authProvider.IsEKSConfig(kubeconfigBytes) {
		klog.V(2).Infof("Detected EKS configuration for path %s, using EKS registration method", path)
		return c.RegisterEKSByString(string(kubeconfigBytes))
	}

	// 使用标准注册方法
	config, err := clientcmd.BuildConfigFromFlags("", path)
	if err != nil {
		return nil, fmt.Errorf("RegisterByPath Error %s %v", path, err)
	}
	return c.RegisterByConfig(config)
}

// RegisterByString 通过kubeconfig文件的string 内容进行注册
func (c *ClusterInstances) RegisterByString(str string) (*Kubectl, error) {
	// 自动检测是否为 EKS 配置
	authProvider := aws.NewAuthProvider()
	if authProvider.IsEKSConfig([]byte(str)) {
		klog.V(2).Infof("Detected EKS configuration, using EKS registration method")
		return c.RegisterEKSByString(str)
	}

	// 使用标准注册方法
	config, err := clientcmd.Load([]byte(str))
	if err != nil {
		return nil, fmt.Errorf("RegisterByString Error,content=:\n%s\n,err:%v", str, err)
	}

	clientConfig := clientcmd.NewDefaultClientConfig(*config, &clientcmd.ConfigOverrides{})
	restConfig, err := clientConfig.ClientConfig()
	if err != nil {
		return nil, err
	}
	return c.RegisterByConfig(restConfig)
}

// RegisterByStringWithID 通过kubeconfig文件的string 内容进行注册
func (c *ClusterInstances) RegisterByStringWithID(str string, id string) (*Kubectl, error) {
	// 自动检测是否为 EKS 配置
	authProvider := aws.NewAuthProvider()
	if authProvider.IsEKSConfig([]byte(str)) {
		klog.V(2).Infof("Detected EKS configuration for ID %s, using EKS registration method", id)
		return c.RegisterEKSByStringWithID(str, id)
	}

	// 使用标准注册方法
	config, err := clientcmd.Load([]byte(str))
	if err != nil {
		return nil, fmt.Errorf("RegisterByStringWithID Error content=\n%s\n,id:%s,err:%v", str, id, err)
	}

	clientConfig := clientcmd.NewDefaultClientConfig(*config, &clientcmd.ConfigOverrides{})
	restConfig, err := clientConfig.ClientConfig()
	if err != nil {
		return nil, err
	}
	return c.RegisterByConfigWithID(restConfig, id)
}

// RegisterByPathWithID 通过kubeconfig文件路径注册集群
func (c *ClusterInstances) RegisterByPathWithID(path string, id string) (*Kubectl, error) {
	// 先读取文件内容以检测是否为 EKS 配置
	content, err := clientcmd.LoadFromFile(path)
	if err != nil {
		return nil, fmt.Errorf("RegisterByPathWithID Error loading file path:%s,id:%s,err:%v", path, id, err)
	}

	// 将配置转换为字节数组进行 EKS 检测
	kubeconfigBytes, err := clientcmd.Write(*content)
	if err != nil {
		return nil, fmt.Errorf("RegisterByPathWithID Error serializing kubeconfig path:%s,id:%s,err:%v", path, id, err)
	}

	// 自动检测是否为 EKS 配置
	authProvider := aws.NewAuthProvider()
	if authProvider.IsEKSConfig(kubeconfigBytes) {
		klog.V(2).Infof("Detected EKS configuration for path %s, ID %s, using EKS registration method", path, id)
		return c.RegisterEKSByStringWithID(string(kubeconfigBytes), id)
	}

	// 使用标准注册方法
	config, err := clientcmd.BuildConfigFromFlags("", path)
	if err != nil {
		return nil, fmt.Errorf("RegisterByPathWithID Error path:%s,id:%s,err:%v", path, id, err)
	}
	return c.RegisterByConfigWithID(config, id)
}

// RegisterByConfig 注册集群
func (c *ClusterInstances) RegisterByConfig(config *rest.Config) (*Kubectl, error) {
	if config == nil {
		return nil, fmt.Errorf("config is nil")
	}
	host := config.Host

	return c.RegisterByConfigWithID(config, host)
}

// RegisterByConfigWithID 注册集群
func (c *ClusterInstances) RegisterByConfigWithID(config *rest.Config, id string) (*Kubectl, error) {
	if config == nil {
		return nil, fmt.Errorf("config is nil")
	}
	config.QPS = 200
	config.Burst = 2000
	if value, exists := clusterInstances.clusters.Load(id); exists {
		cluster := value.(*ClusterInst)
		return cluster.Kubectl, nil
	} else {
		// key 不存在，进行初始化
		k := initKubectl(config, id)
		cluster := &ClusterInst{
			ID:      id,
			Kubectl: k,
			Config:  config,
		}
		clusterInstances.clusters.Store(id, cluster)

		client, err := kubernetes.NewForConfig(config)
		if err != nil {
			return nil, fmt.Errorf("RegisterByConfigWithID Error %s %v", id, err)
		}
		dynamicClient, err := dynamic.NewForConfig(config)
		if err != nil {
			return nil, fmt.Errorf("RegisterByConfigWithID Error %s %v", id, err)
		}
		cluster.Client = client               // kubernetes 客户端
		cluster.DynamicClient = dynamicClient // 动态客户端
		// 缓存
		cluster.crdList = k.initializeCRDList(time.Minute * 10) // CRD列表,10分钟缓存
		cluster.callbacks = k.initializeCallbacks()             // 回调
		cluster.serverVersion = k.initializeServerVersion()     // 服务器版本
		cluster.openAPISchema = k.getOpenAPISchema()
		cluster.docs = doc.InitTrees(k.getOpenAPISchema()) // 文档
		cluster.describerMap = k.initializeDescriberMap()  // 初始化描述器
		if c.callbackRegisterFunc != nil {                 // 注册回调方法
			c.callbackRegisterFunc(cluster)
		}

		cache, err := ristretto.NewCache(&ristretto.Config[string, any]{
			NumCounters: 1e7,     // number of keys to track frequency of (10M).
			MaxCost:     1 << 30, // maximum cost of cache (1GB).
			BufferItems: 64,      // number of keys per Get buffer.
		})
		cluster.Cache = cache

		// 启动CRD监控，有更新的时候，更新APIResources
		ctx, cf := context.WithCancel(context.Background())
		cluster.watchCRDCancelFunc = cf
		err = k.WatchCRDAndRefreshDiscovery(ctx)

		if err != nil {
			return nil, err
		}
		return k, nil
	}
}

// GetClusterById 根据集群ID获取集群实例
func (c *ClusterInstances) GetClusterById(id string) *ClusterInst {
	if value, exists := c.clusters.Load(id); exists {
		return value.(*ClusterInst)
	}
	return nil
}

// RemoveClusterById 删除集群
func (c *ClusterInstances) RemoveClusterById(id string) {
	if value, exists := c.clusters.Load(id); exists {
		cluster := value.(*ClusterInst)

		// 如果是 EKS 集群，停止 token 刷新
		if cluster.IsEKS {
			if cluster.tokenRefreshCancel != nil {
				cluster.tokenRefreshCancel()
			}
			if cluster.AWSAuthProvider != nil {
				cluster.AWSAuthProvider.Stop()
			}
		}

		// 释放 ristretto.Cache 资源
		if cluster.Cache != nil {
			cluster.Cache.Close()
			cluster.Cache = nil
		}
		// 释放其他成员（如有需要，可扩展）
		cluster.Client = nil
		cluster.DynamicClient = nil
		cluster.apiResources = nil
		cluster.crdList = nil
		cluster.callbacks = nil
		cluster.docs = nil
		cluster.serverVersion = nil
		cluster.describerMap = nil
		cluster.openAPISchema = nil
		if cluster.watchCRDCancelFunc != nil {
			cluster.watchCRDCancelFunc()
		}

	}
	c.clusters.Delete(id)
	go runtime.GC()
}

// AllClusters 返回所有集群实例
func (c *ClusterInstances) AllClusters() map[string]*ClusterInst {
	result := make(map[string]*ClusterInst)
	c.clusters.Range(func(key, value interface{}) bool {
		result[key.(string)] = value.(*ClusterInst)
		return true
	})
	return result
}

// DefaultCluster 返回一个默认的 ClusterInst 实例。
// 当 clusters 列表为空时，返回 nil。
// 首先尝试返回 ID 为 "InCluster" 的实例，如果不存在，
// 则尝试返回 ID 为 "default" 的实例。
// 如果上述两个实例都不存在，则返回 clusters 列表中的任意一个实例。
func (c *ClusterInstances) DefaultCluster() *ClusterInst {
	// 尝试获取 ID 为 "InCluster" 的集群实例
	id := "InCluster"
	if value, exists := c.clusters.Load(id); exists {
		return value.(*ClusterInst)
	}

	// 尝试获取 ID 为 "default" 的集群实例
	id = "default"
	if value, exists := c.clusters.Load(id); exists {
		return value.(*ClusterInst)
	}

	// 如果上述两个实例都不存在，遍历 clusters 列表，返回任意一个实例
	var result *ClusterInst
	c.clusters.Range(func(key, value interface{}) bool {
		result = value.(*ClusterInst)
		return false // 返回第一个找到的实例
	})

	return result
}

// Show 显示所有集群信息
func (c *ClusterInstances) Show() {
	klog.Infof("Show Clusters\n")
	c.clusters.Range(func(key, value interface{}) bool {
		k := key.(string)
		v := value.(*ClusterInst)
		if v.serverVersion == nil {
			klog.Infof("%s=nil\n", k)
		} else {
			klog.Infof("%s[%s,%s]=%s\n", k, v.serverVersion.Platform, v.serverVersion.GitVersion, v.Config.Host)
		}
		return true
	})
}

// AWS EKS 相关注册方法

// RegisterEKSByString 通过 kubeconfig 字符串注册 EKS 集群
func (c *ClusterInstances) RegisterEKSByString(str string) (*Kubectl, error) {
	// 检测是否为 EKS 配置
	authProvider := aws.NewAuthProvider()
	if !authProvider.IsEKSConfig([]byte(str)) {
		return nil, fmt.Errorf("not an EKS kubeconfig")
	}

	// 解析配置并获取集群名称作为 ID
	if err := authProvider.InitializeFromKubeconfig([]byte(str)); err != nil {
		return nil, fmt.Errorf("failed to initialize EKS auth provider: %w", err)
	}

	clusterName, _, _ := authProvider.GetClusterInfo()
	id := fmt.Sprintf("eks-%s", clusterName)

	return c.RegisterEKSByStringWithID(str, id)
}

// RegisterEKSByStringWithID 通过 kubeconfig 字符串注册 EKS 集群 (带 ID)
func (c *ClusterInstances) RegisterEKSByStringWithID(str string, id string) (*Kubectl, error) {
	// 检查是否已存在
	if value, exists := c.clusters.Load(id); exists {
		cluster := value.(*ClusterInst)
		return cluster.Kubectl, nil
	}

	// 创建 AWS 认证提供者
	authProvider := aws.NewAuthProvider()
	if err := authProvider.InitializeFromKubeconfig([]byte(str)); err != nil {
		return nil, fmt.Errorf("failed to initialize EKS auth provider: %w", err)
	}

	// 获取初始 token
	ctx := context.Background()
	token, _, err := authProvider.GetToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get initial AWS token: %w", err)
	}

	// 解析标准 kubeconfig 结构
	config, err := clientcmd.Load([]byte(str))
	if err != nil {
		return nil, fmt.Errorf("failed to load kubeconfig: %w", err)
	}

	// 创建 rest.Config，但使用 Bearer token 认证
	clientConfig := clientcmd.NewDefaultClientConfig(*config, &clientcmd.ConfigOverrides{})
	restConfig, err := clientConfig.ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to create rest config: %w", err)
	}

	// 清除原有认证信息，使用 Bearer token
	restConfig.Username = ""
	restConfig.Password = ""
	restConfig.CertFile = ""
	restConfig.KeyFile = ""
	restConfig.CertData = nil
	restConfig.KeyData = nil
	restConfig.BearerToken = token
	restConfig.BearerTokenFile = ""

	// 设置性能参数
	restConfig.QPS = 200
	restConfig.Burst = 2000

	// 创建集群实例
	k := initKubectl(restConfig, id)
	cluster := &ClusterInst{
		ID:              id,
		Kubectl:         k,
		Config:          restConfig,
		AWSAuthProvider: authProvider,
		IsEKS:           true,
	}
	c.clusters.Store(id, cluster)

	// 创建 Kubernetes 客户端
	client, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes client: %w", err)
	}
	dynamicClient, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create dynamic client: %w", err)
	}

	cluster.Client = client
	cluster.DynamicClient = dynamicClient

	// 初始化其他组件
	cluster.crdList = k.initializeCRDList(time.Minute * 10)
	cluster.callbacks = k.initializeCallbacks()
	cluster.serverVersion = k.initializeServerVersion()
	cluster.openAPISchema = k.getOpenAPISchema()
	cluster.docs = doc.InitTrees(k.getOpenAPISchema())
	cluster.describerMap = k.initializeDescriberMap()

	// 注册回调
	if c.callbackRegisterFunc != nil {
		c.callbackRegisterFunc(cluster)
	}

	// 创建缓存
	cache, err := ristretto.NewCache(&ristretto.Config[string, any]{
		NumCounters: 1e7,
		MaxCost:     1 << 30,
		BufferItems: 64,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create cache: %w", err)
	}
	cluster.Cache = cache

	// 启动 CRD 监控
	crdCtx, crdCancel := context.WithCancel(context.Background())
	cluster.watchCRDCancelFunc = crdCancel
	if err := k.WatchCRDAndRefreshDiscovery(crdCtx); err != nil {
		return nil, fmt.Errorf("failed to start CRD watch: %w", err)
	}

	// 启动 token 自动刷新
	tokenCtx, tokenCancel := context.WithCancel(context.Background())
	cluster.tokenRefreshCancel = tokenCancel
	authProvider.StartAutoRefresh(tokenCtx)

	// 设置 token 刷新回调，更新 rest.Config 中的 BearerToken
	go c.startTokenRefreshForCluster(tokenCtx, cluster)

	klog.V(2).Infof("Successfully registered EKS cluster: %s", id)

	return k, nil
}

// RegisterEKSByPath 通过 kubeconfig 文件路径注册 EKS 集群
func (c *ClusterInstances) RegisterEKSByPath(path string) (*Kubectl, error) {
	// 读取文件内容
	content, err := clientcmd.LoadFromFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to load kubeconfig from file %s: %w", path, err)
	}

	// 将配置转换为字节数组
	kubeconfigBytes, err := clientcmd.Write(*content)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize kubeconfig: %w", err)
	}

	return c.RegisterEKSByString(string(kubeconfigBytes))
}

// RegisterEKSByPathWithID 通过 kubeconfig 文件路径注册 EKS 集群 (带 ID)
func (c *ClusterInstances) RegisterEKSByPathWithID(path string, id string) (*Kubectl, error) {
	// 读取文件内容
	content, err := clientcmd.LoadFromFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to load kubeconfig from file %s: %w", path, err)
	}

	// 将配置转换为字节数组
	kubeconfigBytes, err := clientcmd.Write(*content)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize kubeconfig: %w", err)
	}

	return c.RegisterEKSByStringWithID(string(kubeconfigBytes), id)
}

// startTokenRefreshForCluster 为 EKS 集群启动 token 刷新循环
func (c *ClusterInstances) startTokenRefreshForCluster(ctx context.Context, cluster *ClusterInst) {
	if !cluster.IsEKS || cluster.AWSAuthProvider == nil {
		return
	}

	ticker := time.NewTicker(5 * time.Minute) // 每5分钟检查一次
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			klog.V(2).Infof("Stopping token refresh for cluster %s", cluster.ID)
			return
		case <-ticker.C:
			// 检查 token 是否需要刷新
			if !cluster.AWSAuthProvider.IsTokenValid() {
				klog.V(3).Infof("Token expired for cluster %s, refreshing...", cluster.ID)
				if token, _, err := cluster.AWSAuthProvider.GetToken(ctx); err != nil {
					klog.Errorf("Failed to refresh token for cluster %s: %v", cluster.ID, err)
				} else {
					// 更新 rest.Config 中的 BearerToken
					cluster.Config.BearerToken = token
					klog.V(2).Infof("Successfully refreshed token for cluster %s", cluster.ID)
				}
			}
		}
	}
}

// IsEKSCluster 检查集群是否为 EKS 集群
func (ci *ClusterInst) IsEKSCluster() bool {
	return ci.IsEKS
}

// GetAWSAuthProvider 获取 AWS 认证提供者
func (ci *ClusterInst) GetAWSAuthProvider() *aws.AuthProvider {
	return ci.AWSAuthProvider
}

// StopTokenRefresh 停止 token 自动刷新
func (ci *ClusterInst) StopTokenRefresh() {
	if ci.tokenRefreshCancel != nil {
		ci.tokenRefreshCancel()
		ci.tokenRefreshCancel = nil
	}
}

// GetServerVersion 获取服务器版本信息
func (ci *ClusterInst) GetServerVersion() *version.Info {
	return ci.serverVersion
}

// NewAWSAuthProvider 创建新的 AWS 认证提供者实例
func NewAWSAuthProvider() *aws.AuthProvider {
	return aws.NewAuthProvider()
}
