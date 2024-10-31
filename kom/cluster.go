package kom

import (
	"fmt"

	"github.com/weibaohui/kom/kom/doc"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
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
	clusters             map[string]*clusterInst
	callbackRegisterFunc func(clusters *ClusterInstances) func() // 用来注册回调参数的回调方法
}

// 集群实例
type clusterInst struct {
	ID            string                       // 集群ID
	Kubectl       *Kubectl                     // kom
	Client        *kubernetes.Clientset        // kubernetes 客户端
	Config        *rest.Config                 // rest config
	DynamicClient *dynamic.DynamicClient       // 动态客户端
	apiResources  []*metav1.APIResource        // 当前k8s已注册资源
	crdList       []*unstructured.Unstructured // 当前k8s已注册资源
	callbacks     *callbacks                   // 回调
	docs          *doc.Docs                    // 文档
	serverVersion *version.Info                // 服务器版本
}

// Clusters 集群实例管理器
func Clusters() *ClusterInstances {
	return clusterInstances
}

// 初始化
func init() {
	clusterInstances = &ClusterInstances{
		clusters: make(map[string]*clusterInst),
	}
}

// DefaultCluster 获取默认集群，简化调用方式
func DefaultCluster() *Kubectl {
	return Clusters().DefaultCluster().Kubectl
}

// Cluster 获取集群
func Cluster(id string) *Kubectl {
	return Clusters().GetClusterById(id).Kubectl
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
func (c *ClusterInstances) SetRegisterCallbackFunc(callback func(clusters *ClusterInstances) func()) {
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

	cluster, exists := clusterInstances.clusters[id]
	if exists {
		return cluster.Kubectl, nil
	} else {
		// key 不存在，进行初始化
		k := initKubectl(config, id)
		cluster = &clusterInst{
			ID:      id,
			Kubectl: k,
			Config:  config,
		}
		clusterInstances.clusters[id] = cluster

		client, err := kubernetes.NewForConfig(config)
		if err != nil {
			return nil, fmt.Errorf("RegisterByConfigWithID Error %s %v", id, err)
		}
		dynamicClient, err := dynamic.NewForConfig(config)
		if err != nil {
			return nil, fmt.Errorf("RegisterByConfigWithID Error %s %v", id, err)
		}
		cluster.Client = client                             // kubernetes 客户端
		cluster.DynamicClient = dynamicClient               // 动态客户端
		cluster.apiResources = k.initializeAPIResources()   // API 资源
		cluster.crdList = k.initializeCRDList()             // CRD列表
		cluster.callbacks = k.initializeCallbacks()         // 回调
		cluster.serverVersion = k.initializeServerVersion() // 服务器版本
		cluster.docs = doc.InitTrees(k.getOpenAPISchema())  // 文档
		if c.callbackRegisterFunc != nil {                  // 注册回调方法
			c.callbackRegisterFunc(Clusters())
		}
		return k, nil
	}
}

// GetClusterById 根据集群ID获取集群实例
func (c *ClusterInstances) GetClusterById(id string) *clusterInst {
	cluster, exists := c.clusters[id]
	if !exists {
		return nil
	}
	return cluster
}

// AllClusters 返回所有集群实例
func (c *ClusterInstances) AllClusters() map[string]*clusterInst {
	return c.clusters
}

// DefaultCluster 返回一个默认的 clusterInst 实例。
// 当 clusters 列表为空时，返回 nil。
// 首先尝试返回 ID 为 "InCluster" 的实例，如果不存在，
// 则尝试返回 ID 为 "default" 的实例。
// 如果上述两个实例都不存在，则返回 clusters 列表中的任意一个实例。
func (c *ClusterInstances) DefaultCluster() *clusterInst {
	// 检查 clusters 列表是否为空
	if len(c.clusters) == 0 {
		return nil
	}

	// 尝试获取 ID 为 "InCluster" 的集群实例
	id := "InCluster"
	cluster, exists := c.clusters[id]
	if exists {
		return cluster
	}

	// 尝试获取 ID 为 "default" 的集群实例
	id = "default"
	cluster, exists = c.clusters[id]
	if exists {
		return cluster
	}

	// 如果上述两个实例都不存在，遍历 clusters 列表，返回任意一个实例
	for _, v := range c.clusters {
		return v
	}

	// 如果 clusters 列表为空（理论上此时应已返回），则返回 nil
	return nil
}

// Show 显示所有集群信息
func (c *ClusterInstances) Show() {
	klog.Infof("Show Clusters\n")
	for k, v := range c.clusters {
		if v.serverVersion == nil {
			klog.Infof("%s=nil\n", k)
			continue
		}
		klog.Infof("%s[%s,%s.%s]=%s\n", k, v.serverVersion.Platform, v.serverVersion.Major, v.serverVersion.Minor, v.Config.Host)
	}
}
