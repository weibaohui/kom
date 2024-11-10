package kom

import (
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type tools struct {
	kubectl *Kubectl
}

// ConvertRuntimeObjectToTypedObject 是一个通用的转换函数，将 runtime.Object 转换为指定的目标类型
func (u *tools) ConvertRuntimeObjectToTypedObject(obj runtime.Object, target interface{}) error {
	// 将 obj 断言为 *unstructured.Unstructured 类型
	unstructuredObj, ok := obj.(*unstructured.Unstructured)
	if !ok {
		return fmt.Errorf("无法将对象转换为 *unstructured.Unstructured 类型")
	}

	// 使用 DefaultUnstructuredConverter 将 unstructured 数据转换为具体类型
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(unstructuredObj.Object, target)
	if err != nil {
		return fmt.Errorf("无法将对象转换为目标类型: %v", err)
	}

	return nil
}
func (u *tools) ConvertRuntimeObjectToUnstructuredObject(obj runtime.Object) (*unstructured.Unstructured, error) {
	// 将 obj 断言为 *unstructured.Unstructured 类型
	unstructuredObj, ok := obj.(*unstructured.Unstructured)
	if !ok {
		return nil, fmt.Errorf("无法将对象转换为 *unstructured.Unstructured 类型")
	}

	return unstructuredObj, nil
}
func (u *tools) GetGVRByGVK(gvk schema.GroupVersionKind) (gvr schema.GroupVersionResource, namespaced bool) {
	apiResources := u.kubectl.Status().APIResources()
	for _, resource := range apiResources {
		if resource.Kind == gvk.Kind &&
			resource.Version == gvk.Version &&
			resource.Group == gvk.Group {
			gvr = schema.GroupVersionResource{
				Group:    resource.Group,
				Version:  resource.Version,
				Resource: resource.Name, // 通常是 Kind 的复数形式
			}
			return gvr, resource.Namespaced
		}
	}
	return schema.GroupVersionResource{}, false
}

// getGVR 返回对应 string 的 GroupVersionResource
// 从k8s API接口中获取的值
// 如果同时存在多个version，则返回第一个
// 因此也有可能version不对
func (u *tools) GetGVRByKind(kind string) (gvr schema.GroupVersionResource, namespaced bool) {
	apiResources := u.kubectl.Status().APIResources()
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

// IsBuiltinResource 检查给定的资源种类是否为内置资源。
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
func (u *tools) IsBuiltinResource(kind string) bool {
	apiResources := u.kubectl.Status().APIResources()
	for _, list := range apiResources {
		if list.Kind == kind {
			return true
		}
	}
	return false
}

func (u *tools) GetCRD(kind string, group string) (*unstructured.Unstructured, error) {

	crdList := u.kubectl.Status().CRDList()
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
		return crd, nil
	}
	return nil, fmt.Errorf("crd %s.%s not found", kind, group)
}

func (u *tools) GetParsedGVK(gvks []schema.GroupVersionKind, versions ...string) (gvk schema.GroupVersionKind) {
	if len(gvks) == 0 {
		return schema.GroupVersionKind{}
	}
	if len(versions) > 0 {
		// 指定了版本
		v := versions[0]
		for _, g := range gvks {
			if g.Version == v {
				return schema.GroupVersionKind{
					Kind:    g.Kind,
					Group:   g.Group,
					Version: g.Version,
				}
			}
		}
	} else {
		// 取第一个
		return schema.GroupVersionKind{
			Kind:    gvks[0].Kind,
			Group:   gvks[0].Group,
			Version: gvks[0].Version,
		}
	}
	return
}

// GetGVKFromObj 获取对象的 GroupVersionKind
func (u *tools) GetGVKFromObj(obj interface{}) (schema.GroupVersionKind, error) {
	switch o := obj.(type) {
	case *unstructured.Unstructured:
		return o.GroupVersionKind(), nil
	case runtime.Object:
		return o.GetObjectKind().GroupVersionKind(), nil
	default:
		return schema.GroupVersionKind{}, fmt.Errorf("不支持的类型%v", o)
	}
}

func (u *tools) GetGVRFromCRD(crd *unstructured.Unstructured) schema.GroupVersionResource {
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
