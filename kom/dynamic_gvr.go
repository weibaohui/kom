package kom

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// GetGVR 返回对应 string 的 GroupVersionResource
// 从k8s API接口中获取的值
// 如果同时存在多个version，则返回第一个
// 因此也有可能version不对
func (k8s *Kom) GetGVR(kind string) (gvr schema.GroupVersionResource, namespaced bool) {
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
func (k8s *Kom) IsBuiltinResource(kind string) bool {
	for _, list := range apiResources {
		if list.Kind == kind {
			return true
		}
	}
	return false
}
