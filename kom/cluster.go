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
	config, err := clientcmd.BuildConfigFromFlags("", path)
	if err != nil {
		return nil, fmt.Errorf("RegisterByPath Error %s %v", path, err)
	}
	return c.RegisterByConfig(config)
}

// RegisterByString 通过kubeconfig文件的string 内容进行注册
func (c *ClusterInstances) RegisterByString(str string) (*Kubectl, error) {
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

	var cluster *ClusterInst
	// 检查是否已存在
	if value, exists := clusterInstances.clusters.Load(id); exists {
		cluster = value.(*ClusterInst)
		if cluster.Kubectl != nil {
			return cluster.Kubectl, nil
		}
	}

	// 检查是否需要 exec 认证处理 (AWS EKS 或其他 exec 模式)
	if cluster != nil && cluster.IsEKS {
		err := cluster.AWSAuthProvider.SetEKSExecProvider(config.ExecProvider)
		if err != nil {
			return nil, err
		}
		// 获取初始 token
		ctx := context.Background()
		token, _, err := cluster.AWSAuthProvider.GetToken(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get initial AWS token: %w", err)
		}
		config.BearerToken = token

	}
	// key 不存在，进行初始化
	k := initKubectl(config, id)
	// 正常情况下，走到这里，cluster都是空
	if cluster == nil {
		cluster = &ClusterInst{
			ID:      id,
			Kubectl: k,
			Config:  config,
		}
	} else {
		// 注册EKS集群时，在之前已经占用了一个空的cluster实例，
		// 存储了如下的信息，但是没有kubectl、config
		// IsEKS:              true,
		// AWSAuthProvider:    authProvider,
		// tokenRefreshCancel: tokenCancel,
		cluster.Kubectl = k
		cluster.Config = config
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
	if err != nil {
		return nil, fmt.Errorf("failed to create cache: %w", err)
	}
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

// RegisterAWSCluster 注册EKS集群
func (c *ClusterInstances) RegisterAWSCluster(config aws.EKSAuthConfig) (*Kubectl, error) {

	// 生成集群ID
	clusterID := fmt.Sprintf("%s-%s", config.Region, config.ClusterName)

	var cluster *ClusterInst
	// 检查是否已存在
	if value, exists := clusterInstances.clusters.Load(clusterID); exists {
		cluster = value.(*ClusterInst)
		if cluster.Kubectl != nil {
			return cluster.Kubectl, nil
		}
	}

	var kubeconfigContent string
	var err error

	if cluster == nil {
		// 生成kubeconfig
		generator := aws.NewKubeconfigGenerator()
		kubeconfigContent, err = generator.GenerateFromAWS(config)
		if err != nil {
			return nil, fmt.Errorf("failed to generate kubeconfig for EKS cluster %s: %w", clusterID, err)
		}
		klog.V(2).Infof("Generated kubeconfig for EKS cluster: %s", clusterID)

		// 手动设置 EKS 配置
		tokenManager, err := aws.NewTokenManager(&config)
		if err != nil {
			return nil, fmt.Errorf("failed to create AWS token manager: %w", err)
		}

		authProvider := NewAWSAuthProvider()
		// 设置内部状态
		authProvider.SetEKSConfig(&config)
		authProvider.SetTokenManager(tokenManager)
		config.TokenCache = &aws.TokenCache{}
		tokenCtx, tokenCancel := context.WithCancel(context.Background())
		// 启动自动刷新
		authProvider.StartAutoRefresh(tokenCtx)

		// 启动 token 刷新循环
		go func() {
			ticker := time.NewTicker(5 * time.Minute) // 每5分钟检查一次
			defer ticker.Stop()

			for {
				select {
				case <-tokenCtx.Done():
					klog.V(2).Infof("Stopping token refresh for cluster %s", clusterID)
					return
				case <-ticker.C:
					// 检查 token 是否需要刷新
					if !authProvider.IsTokenValid() {
						klog.V(3).Infof("Token expired for cluster %s, refreshing...", clusterID)
						if token, _, err := authProvider.GetToken(tokenCtx); err != nil {
							klog.Errorf("Failed to refresh token for cluster %s: %v", clusterID, err)
						} else {
							// 更新 rest.Config 中的 BearerToken
							cluster.Config.BearerToken = token
							klog.V(2).Infof("Successfully refreshed token for cluster %s", clusterID)
						}
					}
				}
			}
		}()
		cluster = &ClusterInst{
			ID:                 clusterID,
			IsEKS:              true,
			AWSAuthProvider:    authProvider,
			tokenRefreshCancel: tokenCancel,
		}
	}
	klog.V(8).Infof("kubeconfigContent=\n%s", kubeconfigContent)

	clusterInstances.clusters.Store(clusterID, cluster)

	// 使用kubeconfig注册集群
	kubectl, err := c.RegisterByStringWithID(kubeconfigContent, clusterID)
	if err != nil {
		return nil, fmt.Errorf("failed to register EKS cluster %s: %w", clusterID, err)
	}

	klog.V(1).Infof("Successfully registered EKS cluster: %s", clusterID)

	return kubectl, nil
}

// RefreshEKSCredentials 刷新EKS集群凭证
func (c *ClusterInstances) RefreshEKSCredentials(clusterID string) error {

	// 如果集群实例存在且为EKS集群，触发token刷新
	if cluster := c.GetClusterById(clusterID); cluster != nil && cluster.IsEKS {
		if cluster.AWSAuthProvider != nil {
			cluster.AWSAuthProvider.TriggerRefresh()
		}

	}

	return nil
}
