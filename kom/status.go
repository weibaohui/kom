package kom

import (
	"context"
	"fmt"
	"strings"
	"sync"
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
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
	"k8s.io/utils/strings/slices"
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

// GetResourceCountSummary 获取集群内资源状态统计数据
// Resource                                          Namespaced Count
// ---------------------------------------------------------------------------
// v1/pods                                           true       192
// apps/v1/deployments                               true       43
// batch/v1/jobs                                     true       11
// networking.k8s.io/v1/ingresses                    true       22
// monitoring.coreos.com/v1/servicemonitors          true       8
// apiextensions.k8s.io/v1/customresourcedefinitions false      15
func (s *status) GetResourceCountSummary(cacheSeconds int) (map[schema.GroupVersionResource]int, error) {
	d := time.Duration(cacheSeconds) * time.Second
	return utils.GetOrSetCache(s.kubectl.ClusterCache(), "GetResourceCountSummary", d, func() (map[schema.GroupVersionResource]int, error) {
		ctx := s.kubectl.Statement.Context

		config := s.kubectl.RestConfig()
		dynamicClient, err := dynamic.NewForConfig(config)
		if err != nil {
			return nil, err
		}

		// Create clients
		discoveryClient, err := discovery.NewDiscoveryClientForConfig(config)
		if err != nil {
			return nil, err
		}
		apiGroupResources, err := restmapper.GetAPIGroupResources(discoveryClient)
		if err != nil {
			return nil, err
		}

		var (
			summary = make(map[schema.GroupVersionResource]int)
			mu      sync.Mutex
			wg      sync.WaitGroup
			sema    = make(chan struct{}, 5) // 最多5个并发list
		)

		for _, group := range apiGroupResources {
			for v, resources := range group.VersionedResources {
				for _, resource := range resources {
					// Skip subresources (e.g., pods/status)
					if strings.Contains(resource.Name, "/") || !slices.Contains(resource.Verbs, "list") {
						continue
					}

					gvr := schema.GroupVersionResource{
						Group:    group.Group.Name,
						Version:  v,
						Resource: resource.Name,
					}

					wg.Add(1)
					sema <- struct{}{} // 获取信号量
					go func(namespaced bool, gvr schema.GroupVersionResource) {
						defer wg.Done()
						defer func() {
							<-sema // 释放信号量
						}()
						count, err := countResources(ctx, dynamicClient, gvr, namespaced)
						if err != nil {
							klog.V(6).Infof("[ResourceCount][error] err=%v", err)
							return
						}
						mu.Lock()
						summary[gvr] = count
						mu.Unlock()
					}(resource.Namespaced, gvr)
				}
			}
		}
		wg.Wait()
		return summary, nil
	})
}
func countResources(ctx context.Context, client dynamic.Interface, gvr schema.GroupVersionResource, namespaced bool) (int, error) {
	total := 0
	var continueToken string

	for {
		var list *unstructured.UnstructuredList
		var err error

		if namespaced {
			list, err = client.Resource(gvr).Namespace(metav1.NamespaceAll).List(ctx, metav1.ListOptions{
				Limit:    500,
				Continue: continueToken,
			})
		} else {
			list, err = client.Resource(gvr).List(ctx, metav1.ListOptions{
				Limit:    500,
				Continue: continueToken,
			})
		}
		if err != nil {
			return total, err
		}
		total += len(list.Items)

		continueToken = list.GetContinue()
		if continueToken == "" {
			break
		}
	}

	return total, nil
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
	ck, err := utils.GetOrSetCache(k.ClusterCache(), "crdList", ttl, func() (ret []*unstructured.Unstructured, err error) {
		crdList, err := k.listResources(context.TODO(), "CustomResourceDefinition", "")
		return crdList, err
	})
	if err != nil {
		return nil
	}
	return ck

}
func (k *Kubectl) WatchCRDAndRefreshDiscovery(ctx context.Context) error {
	klog.V(6).Infof("Watching CRD resources")
	cfg := k.RestConfig()

	apixClient, err := apixclient.NewForConfig(cfg)
	if err != nil {
		return fmt.Errorf("failed to create apiextensions client: %v", err)
	}

	var started atomic.Bool

	// 创建一个刷新请求通道（带缓冲，防抖）
	refreshCh := make(chan struct{}, 1)
	go func() {
		var (
			debounceTimer *time.Timer
			mu            sync.Mutex
		)

		triggerRefresh := func() {
			if !started.Load() {
				klog.V(8).Infof("Skipping refresh (initial load)")
				return
			}
			klog.V(6).Infof("Refreshing API resources due to CRD change")
			k.Status().SetAPIResources(k.initializeAPIResources())
		}

		for {
			select {
			case <-ctx.Done():
				return
			case <-refreshCh:
				mu.Lock()
				if debounceTimer != nil {
					// 如果已有定时器，重置它（延迟合并）
					if debounceTimer.Stop() {
						<-debounceTimer.C
					}
				}
				debounceTimer = time.AfterFunc(2*time.Second, func() {
					triggerRefresh()
				})
				mu.Unlock()
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

	// 创建一个带超时的context，默认30秒超时
	syncCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if ok := cache.WaitForCacheSync(syncCtx.Done(), crdInformer.HasSynced); !ok {
		klog.Warningf("Failed to sync CRD informer cache within 30 seconds")
	}

	// started.Store 防止 初始化阶段重复刷新
	// informer 启动后会把现有 CRD 全量同步一遍，此时会产生大量的 “Add” 事件。
	// 如果没有 started，worker 会在第一次 sync 阶段被疯狂触发，从而浪费资源。
	// 有了 started.Store(true) 之后，只有在第一次初始化加载完成后，才允许后续事件触发刷新。

	klog.V(6).Infof("Initial CRD sync done, loading API resources")
	k.Status().SetAPIResources(k.initializeAPIResources())
	started.Store(true)

	klog.V(6).Infof("CRD watcher initialized and running")
	return nil
}

func (k *Kubectl) initializeAPIResources() (apiResources []*metav1.APIResource) {
	klog.V(6).Infof("Loading API resources")
	// 提取ApiResources
	k.Client()
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
