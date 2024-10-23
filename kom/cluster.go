package kom

import (
	"fmt"

	"github.com/weibaohui/kom/kom/docer"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
)

var clusterInstances *ClusterInstances

type ClusterInstances struct {
	clusters             map[string]*ClusterInst
	callbackRegisterFunc func(clusters *ClusterInstances) func() // 用来注册回调参数的回调方法
}
type ClusterInst struct {
	ID            string
	Kom           *Kom
	Client        *kubernetes.Clientset
	Config        *rest.Config
	DynamicClient *dynamic.DynamicClient
	apiResources  []*metav1.APIResource        // 当前k8s已注册资源
	crdList       []*unstructured.Unstructured // 当前k8s已注册资源
	callbacks     *callbacks
	Docs          *docer.Docs
}

func Clusters() *ClusterInstances {
	return clusterInstances
}
func init() {
	clusterInstances = &ClusterInstances{
		clusters: make(map[string]*ClusterInst),
	}
}
func (c *ClusterInstances) InitInCluster() (*Kom, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("InCluster Error %v", err)
	}
	return c.InitByConfigWithID(config, "InCluster")
}
func (c *ClusterInstances) SetCallbackRegisterFunc(callback func(clusters *ClusterInstances) func()) {
	c.callbackRegisterFunc = callback
}
func (c *ClusterInstances) InitByPath(path string) (*Kom, error) {
	config, err := clientcmd.BuildConfigFromFlags("", path)
	if err != nil {
		return nil, fmt.Errorf("InitByPath Error %s %v", path, err)
	}
	return c.InitByConfig(config)
}
func (c *ClusterInstances) InitByPathWithID(path string, id string) (*Kom, error) {
	config, err := clientcmd.BuildConfigFromFlags("", path)
	if err != nil {
		return nil, fmt.Errorf("InitByPathWithID Error path:%s,id:%s,err:%v", path, id, err)
	}
	return c.InitByConfigWithID(config, id)
}
func (c *ClusterInstances) InitByConfig(config *rest.Config) (*Kom, error) {
	if config == nil {
		return nil, fmt.Errorf("config is nil")
	}
	host := config.Host

	return c.InitByConfigWithID(config, host)
}
func (c *ClusterInstances) InitByConfigWithID(config *rest.Config, id string) (*Kom, error) {
	if config == nil {
		return nil, fmt.Errorf("config is nil")
	}

	cluster, exists := clusterInstances.clusters[id]
	if exists {
		return cluster.Kom, nil
	} else {
		// key 不存在，进行初始化
		kom := InitConnectionByConfig(config, id)
		cluster = &ClusterInst{
			ID:     id,
			Kom:    kom,
			Config: config,
		}
		clusterInstances.clusters[id] = cluster

		client, err := kubernetes.NewForConfig(config)
		if err != nil {
			return nil, fmt.Errorf("InitByConfigWithID Error %s %v", id, err)
		}
		dynamicClient, err := dynamic.NewForConfig(config)
		if err != nil {
			return nil, fmt.Errorf("InitByConfigWithID Error %s %v", id, err)
		}
		cluster.Client = client
		cluster.DynamicClient = dynamicClient
		cluster.apiResources = kom.initializeAPIResources()
		cluster.crdList = kom.initializeCRDList()
		cluster.callbacks = kom.initializeCallbacks()
		cluster.Docs = docer.InitTrees(kom.GetOpenAPISchema())
		if c.callbackRegisterFunc != nil {
			c.callbackRegisterFunc(Clusters())
		}
		return kom, nil
	}
}
func (c *ClusterInstances) GetById(id string) *ClusterInst {
	cluster, exists := c.clusters[id]
	if !exists {
		return nil
	}
	return cluster
}
func (c *ClusterInstances) All() map[string]*ClusterInst {
	return c.clusters
}

// Default 返回一个默认的 ClusterInst 实例。
// 当 clusters 列表为空时，返回 nil。
// 首先尝试返回 ID 为 "InCluster" 的实例，如果不存在，
// 则尝试返回 ID 为 "default" 的实例。
// 如果上述两个实例都不存在，则返回 clusters 列表中的任意一个实例。
func (c *ClusterInstances) Default() *ClusterInst {
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

func (c *ClusterInstances) Show() {
	klog.Infof("Show Clusters\n")
	for k, _ := range c.clusters {
		klog.Infof("%s\n", k)
	}
}
