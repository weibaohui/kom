package kom

import (
    "context"
    "fmt"
    "net/http"
    "net/url"
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

// init 初始化
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

// SetRegisterCallbackFunc 设置回调注册函数
func (c *ClusterInstances) SetRegisterCallbackFunc(callback func(cluster *ClusterInst) func()) {
	c.callbackRegisterFunc = callback
}

// RegisterByConfig 通过rest.Config注册集群
// 参数:
//   - config: Kubernetes rest.Config配置对象
//
// 返回值:
//   - *Kubectl: 成功时返回 Kubectl 实例，用于操作集群
//   - error: 失败时返回错误信息
//
// 此函数使用配置中的Host字段作为集群ID
func (c *ClusterInstances) RegisterByConfig(config *rest.Config, opts ...RegisterOption) (*Kubectl, error) {
    if config == nil {
        return nil, fmt.Errorf("config is nil")
    }
    host := config.Host

    return c.RegisterByConfigWithID(config, host, opts...)
}

// RegisterByConfigWithID 注册集群
func (c *ClusterInstances) RegisterByConfigWithID(config *rest.Config, id string, opts ...RegisterOption) (*Kubectl, error) {
    if config == nil {
        return nil, fmt.Errorf("config is nil")
    }
    // defaults
    config.QPS = 200
    config.Burst = 2000

    // collect registration options
    params := &RegisterParams{}
    for _, opt := range opts {
        if opt != nil {
            opt(params)
        }
    }

    // apply rest.Config options
    if params.QPS != nil {
        config.QPS = *params.QPS
    }
    if params.Burst != nil {
        config.Burst = *params.Burst
    }
    if params.Timeout > 0 {
        config.Timeout = params.Timeout
    }
    if params.UserAgent != "" {
        config.UserAgent = params.UserAgent
    }
    if params.TLSInsecure {
        config.TLSClientConfig.Insecure = true
    }
    if len(params.CACert) > 0 {
        config.TLSClientConfig.CAData = params.CACert
        config.TLSClientConfig.Insecure = false
    }
    if params.Impersonation != nil {
        config.Impersonate = *params.Impersonation
    }
    if params.ProxyFunc != nil {
        config.Proxy = params.ProxyFunc
    } else if params.ProxyURL != "" {
        if proxyURL, err := url.Parse(params.ProxyURL); err == nil {
            config.Proxy = func(*http.Request) (*url.URL, error) { return proxyURL, nil }
        } else {
            klog.V(4).Infof("invalid proxy url: %s, err: %v", params.ProxyURL, err)
        }
    }

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

    cacheCfg := params.CacheConfig
    if cacheCfg == nil {
        cacheCfg = &ristretto.Config[string, any]{
            NumCounters: 1e7,     // number of keys to track frequency of (10M).
            MaxCost:     1 << 30, // maximum cost of cache (1GB).
            BufferItems: 64,      // number of keys per Get buffer.
        }
    }
    cache, err := ristretto.NewCache(cacheCfg)
    if err != nil {
        return nil, fmt.Errorf("failed to create cache: %w", err)
    }
    cluster.Cache = cache

    // 启动CRD监控，有更新的时候，更新APIResources
    if !params.DisableCRDWatch {
        ctx, cf := context.WithCancel(context.Background())
        cluster.watchCRDCancelFunc = cf
        err = k.WatchCRDAndRefreshDiscovery(ctx)
        if err != nil {
            return nil, err
        }
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

// GetServerVersion 获取服务器版本信息
func (ci *ClusterInst) GetServerVersion() *version.Info {
	return ci.serverVersion
}
