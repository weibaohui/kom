package kom

import (
	"context"
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	openapi_v2 "github.com/google/gnostic-models/openapiv2"
	"github.com/weibaohui/kom/kom/describe"
	"github.com/weibaohui/kom/kom/doc"
	"github.com/weibaohui/kom/utils"
	apixclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	apixinformers "k8s.io/apiextensions-apiserver/pkg/client/informers/externalversions"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/version"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
)

type status struct {
	kubectl *Kubectl
}

func (s *status) SetAPIResources(apiResources []*metav1.APIResource) {
	cluster := s.kubectl.parentCluster()
	cluster.apiResources = apiResources
}
func (s *status) APIResources() []*metav1.APIResource {
	cluster := s.kubectl.parentCluster()
	return cluster.apiResources
}
func (s *status) CRDList() []*unstructured.Unstructured {
	return s.kubectl.initializeCRDList(time.Minute * 10)
}
func (s *status) Docs() *doc.Docs {
	cluster := s.kubectl.parentCluster()
	return cluster.docs
}
func (s *status) ServerVersion() *version.Info {
	cluster := s.kubectl.parentCluster()
	return cluster.serverVersion
}
func (s *status) DescriberMap() map[schema.GroupKind]describe.ResourceDescriber {
	cluster := s.kubectl.parentCluster()
	return cluster.describerMap
}
func (s *status) OpenAPISchema() *openapi_v2.Document {
	cluster := s.kubectl.parentCluster()
	return cluster.openAPISchema
}

func (s *status) IsGatewayAPISupported() bool {
	list := s.CRDList()
	name := "gateways.gateway.networking.k8s.io"
	for _, crd := range list {
		if crd.GetName() == name {
			return true
		}
	}
	return false
}

// IsCRDSupportedByName 判断CRD是否存在
func (s *status) IsCRDSupportedByName(name string) bool {
	list := s.CRDList()
	for _, crd := range list {
		if crd.GetName() == name {
			return true
		}
	}
	return false
}

// 获取版本信息
func (k *Kubectl) initializeServerVersion() *version.Info {
	versionInfo, err := k.Client().Discovery().ServerVersion()
	if err != nil {
		klog.V(2).Infof("Error getting server version: %v\n", err)
	}
	return versionInfo
}

func (k *Kubectl) getOpenAPISchema() *openapi_v2.Document {
	openAPISchema, err := k.Client().Discovery().OpenAPISchema()
	if err != nil {
		klog.V(2).Infof("Error fetching OpenAPI schema: %v\n", err)
		return nil
	}
	return openAPISchema
}

func (k *Kubectl) initializeCRDList(ttl time.Duration) []*unstructured.Unstructured {
	cache, err := utils.GetOrSetCache(k.ClusterCache(), "crdList", ttl, func() (ret []*unstructured.Unstructured, err error) {
		crdList, err := k.listResources(context.TODO(), "CustomResourceDefinition", "")
		return crdList, err
	})
	if err != nil {
		return nil
	}
	return cache

}
func (k *Kubectl) WatchCRDAndRefreshDiscovery(ctx context.Context) error {
	klog.V(6).Infof("Watching CRD resources")

	cfg := k.RestConfig()

	apixClient, err := apixclient.NewForConfig(cfg)
	if err != nil {
		return fmt.Errorf("failed to create apiextensions client: %v", err)
	}

	rawDiscovery, err := discovery.NewDiscoveryClientForConfig(cfg)
	if err != nil {
		return fmt.Errorf("failed to create discovery client: %v", err)
	}

	cachedDiscovery := memory.NewMemCacheClient(rawDiscovery)

	var started atomic.Bool

	// 创建一个刷新请求通道（带缓冲，防抖）
	refreshCh := make(chan struct{}, 1)

	// 后台刷新 worker，只处理信号，不管来源
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-refreshCh:
				if !started.Load() {
					klog.V(8).Infof("Skipping refresh (initial load)")
					continue
				}
				klog.V(6).Infof("Refreshing API resources due to CRD change")
				cachedDiscovery.Invalidate()
				k.Status().SetAPIResources(k.initializeAPIResources())
			}
		}
	}()

	// 创建 informer
	factory := apixinformers.NewSharedInformerFactory(apixClient, 0)
	crdInformer := factory.Apiextensions().V1().CustomResourceDefinitions().Informer()

	_, _ = crdInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			select {
			case refreshCh <- struct{}{}:
			default:
				// 已有信号在队列中，不重复推送
			}
		},
		DeleteFunc: func(obj interface{}) {
			select {
			case refreshCh <- struct{}{}:
			default:
			}
		},
	})

	factory.Start(ctx.Done())

	if ok := cache.WaitForCacheSync(ctx.Done(), crdInformer.HasSynced); !ok {
		return fmt.Errorf("failed to sync CRD informer cache")
	}

	klog.V(6).Infof("Initial CRD sync done, loading API resources")
	k.Status().SetAPIResources(k.initializeAPIResources())
	started.Store(true)

	klog.V(6).Infof("CRD watcher initialized and running")
	return nil
}

func (k *Kubectl) initializeAPIResources() (apiResources []*metav1.APIResource) {
	klog.V(6).Infof("Loading API resources")
	// 提取ApiResources
	_, lists, _ := k.Client().Discovery().ServerGroupsAndResources()
	for _, list := range lists {
		resources := list.APIResources
		ver := list.GroupVersionKind().Version
		group := list.GroupVersionKind().Group
		groupVersion := list.GroupVersion
		gvs := strings.Split(groupVersion, "/")
		if len(gvs) == 2 {
			group = gvs[0]
			ver = gvs[1]
		} else {
			// 只有version的情况"v1"
			ver = groupVersion
		}

		for _, resource := range resources {
			resource.Group = group
			resource.Version = ver
			apiResources = append(apiResources, &resource)
		}
	}
	return apiResources
}
func (k *Kubectl) initializeDescriberMap() map[schema.GroupKind]describe.ResourceDescriber {
	return describe.InitializeDescriberMap(k.RestConfig())
}
