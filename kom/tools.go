package kom

import (
	"fmt"
	"strings"

	"github.com/duke-git/lancet/v2/slice"
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

func (u *tools) ParseGVK2GVR(gvks []schema.GroupVersionKind, versions ...string) (gvr schema.GroupVersionResource, namespaced bool) {
	// 获取单个GVK
	gvk := u.GetGVK(gvks, versions...)

	// 获取GVR
	if u.IsBuiltinResource(gvk.Kind) {
		// 内置资源
		return u.GetGVRByKind(gvk.Kind)
	} else {
		crd, err := u.GetCRD(gvk.Kind, gvk.Group)
		if err != nil {
			return
		}
		// 检查CRD是否是Namespaced
		namespaced = crd.Object["spec"].(map[string]interface{})["scope"].(string) == "Namespaced"
		gvr = u.GetGVRFromCRD(crd)
	}

	return
}

func (u *tools) GetGVK(gvks []schema.GroupVersionKind, versions ...string) (gvk schema.GroupVersionKind) {
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

// FindGVKByTableNameInApiResources 从 APIResource 列表中查找表名对应的 GVK
// APIResource 包含了CRD的内容
func (u *tools) FindGVKByTableNameInApiResources(tableName string) *schema.GroupVersionKind {

	for _, resource := range u.kubectl.parentCluster().apiResources {
		// 比较表名和资源名 (Name) 或 Kind
		if resource.Name == tableName || resource.Kind == tableName || resource.SingularName == tableName ||
			slice.Contain(resource.ShortNames, tableName) {
			// 构建 GroupVersionKind 并返回
			return &schema.GroupVersionKind{
				Group:   resource.Group,   // API 组
				Version: resource.Version, // 版本
				Kind:    resource.Kind,    // Kind
			}
		}
	}
	return nil // 没有匹配的资源
}

// FindGVKByTableNameInCRDList 从CRD列表中找到对应的表名的GVK
func (u *tools) FindGVKByTableNameInCRDList(tableName string) *schema.GroupVersionKind {

	for _, crd := range u.kubectl.parentCluster().crdList {
		// 从 CRD 对象中获取 "spec" 下的 names 字段
		specNames, found, err := unstructured.NestedMap(crd.Object, "spec", "names")
		if err != nil || !found {
			continue // 如果 spec.names 不存在，跳过当前 CRD
		}

		// 提取 kind 和 plural
		kind, _ := specNames["kind"].(string)
		plural, _ := specNames["plural"].(string)
		singular, _ := specNames["singular"].(string)
		shortNames, _, _ := unstructured.NestedStringSlice(crd.Object, "spec", "names", "shortNames")

		// 比较 tableName 是否匹配 kind 或 plural
		if tableName == kind || tableName == plural || tableName == singular || slice.Contain(shortNames, tableName) {
			// 提取 group 和 version
			group, _, _ := unstructured.NestedString(crd.Object, "spec", "group")
			versions, found, _ := unstructured.NestedSlice(crd.Object, "spec", "versions")
			if !found || len(versions) == 0 {
				continue
			}

			// 获取第一个版本的 name 字段
			versionMap, ok := versions[0].(map[string]interface{})
			if !ok {
				continue
			}
			version, _ := versionMap["name"].(string)

			// 返回 GVK
			return &schema.GroupVersionKind{
				Group:   group,
				Version: version,
				Kind:    kind,
			}
		}
	}
	return nil // 未找到匹配项
}
func (u *tools) ListAvailableTableNames() (names []string) {
	for _, resource := range u.kubectl.parentCluster().apiResources {
		// 比较表名和资源名 (Name) 或 Kind
		names = append(names, strings.ToLower(resource.Kind))
		for _, name := range resource.ShortNames {
			names = append(names, name)
		}
	}

	names = slice.Unique(names)
	names = slice.Filter(names, func(index int, item string) bool {
		return !strings.Contains(item, "Option")
	})
	slice.Sort(names, "asc")
	return names
}
