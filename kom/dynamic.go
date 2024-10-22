package kom

import (
	"context"
	"fmt"

	"github.com/weibaohui/kom/kom/option"
	"github.com/weibaohui/kom/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func (kom *Kom) ListResources(ctx context.Context, kind string, ns string, opts ...option.ListOption) ([]unstructured.Unstructured, error) {
	gvr, namespaced := getGVR(kind)
	if gvr.Empty() {
		return nil, fmt.Errorf("不支持的资源类型: %s", kind)
	}

	listOptions := metav1.ListOptions{}
	for _, opt := range opts {
		opt(&listOptions)
	}

	var err error

	var list *unstructured.UnstructuredList
	if namespaced {
		list, err = kom.dynamicClient.Resource(gvr).Namespace(ns).List(ctx, listOptions)
	} else {
		list, err = kom.dynamicClient.Resource(gvr).List(ctx, listOptions)
	}
	if err != nil {
		return nil, err
	}
	var resources []unstructured.Unstructured
	for _, item := range list.Items {
		obj := item.DeepCopy()
		utils.RemoveManagedFields(obj)
		resources = append(resources, *obj)
	}

	return resources, nil
}

// getGVR 返回对应 string 的 GroupVersionResource
// 从k8s API接口中获取的值
// 如果同时存在多个version，则返回第一个
// 因此也有可能version不对
func getGVR(kind string) (gvr schema.GroupVersionResource, namespaced bool) {
	for _, resource := range apiResources {
		if resource.Kind == kind {
			version := resource.Version
			gvr = schema.GroupVersionResource{
				Group:    resource.Group,
				Version:  version,
				Resource: resource.Name, // 通常是 Kind 的复数形式
			}
			return gvr, resource.Namespaced
		}
	}
	return schema.GroupVersionResource{}, false
}

// isBuiltinResource 检查给定的资源种类是否为内置资源。
// 该函数通过遍历apiResources列表，对比每个列表项的Kind属性与给定的kind参数是否匹配。
// 如果找到匹配项，即表明该资源种类是内置资源，函数返回true；否则，返回false。
// 此函数主要用于资源种类的快速校验，以确定资源是否属于预定义的内置类型。
//
// 参数:
//
//	kind (string): 要检查的资源种类的名称。
//
// 返回值:
//
//	bool: 如果kind是内置资源种类之一，则返回true；否则返回false。
func isBuiltinResource(kind string) bool {
	for _, list := range apiResources {
		if list.Kind == kind {
			return true
		}
	}
	return false
}
func GetCRD(kind string, group string) (*unstructured.Unstructured, error) {
	// crdList 是共享变量，初始化时已加载
	for _, crd := range crdList {
		spec, found, err := unstructured.NestedMap(crd.Object, "spec")
		if err != nil || !found {
			continue
		}
		crdKind, found, err := unstructured.NestedString(spec, "names", "kind")
		if err != nil || !found {
			continue
		}
		crdGroup, found, err := unstructured.NestedString(spec, "group")
		if err != nil || !found {
			continue
		}
		if crdKind != kind || crdGroup != group {
			continue
		}
		return &crd, nil
	}
	return nil, fmt.Errorf("crd %s.%s not found", kind, group)
}

func GetGRVFromCRD(crd *unstructured.Unstructured) schema.GroupVersionResource {
	// 提取 GVR
	group := crd.Object["spec"].(map[string]interface{})["group"].(string)
	version := crd.Object["spec"].(map[string]interface{})["versions"].([]interface{})[0].(map[string]interface{})["name"].(string)
	resource := crd.Object["spec"].(map[string]interface{})["names"].(map[string]interface{})["plural"].(string)

	gvr := schema.GroupVersionResource{
		Group:    group,
		Version:  version,
		Resource: resource,
	}
	return gvr
}

// getGVKFromObj 获取对象的 GroupVersionKind
func getGVKFromObj(obj interface{}) (schema.GroupVersionKind, error) {
	switch o := obj.(type) {
	case *unstructured.Unstructured:
		return o.GroupVersionKind(), nil
	case runtime.Object:
		return o.GetObjectKind().GroupVersionKind(), nil
	default:
		return schema.GroupVersionKind{}, fmt.Errorf("不支持的类型%v", o)
	}
}

func ParseGVK2GVR(gvks []schema.GroupVersionKind, versions ...string) (gvr schema.GroupVersionResource, namespaced bool) {

	// 获取单个GVK
	gvk := getParsedGVK(gvks, versions...)

	// 获取GVR
	if isBuiltinResource(gvk.Kind) {
		// 内置资源
		return getGVR(gvk.Kind)
	} else {
		crd, err := GetCRD(gvk.Kind, gvk.Group)
		if err != nil {
			return
		}
		// 检查CRD是否是Namespaced
		namespaced = crd.Object["spec"].(map[string]interface{})["scope"].(string) == "Namespaced"
		gvr = GetGRVFromCRD(crd)
	}

	return
}
