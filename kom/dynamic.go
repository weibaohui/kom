package kom

import (
	"context"
	"fmt"

	"github.com/weibaohui/kom/kom/option"
	"github.com/weibaohui/kom/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// listResources 列出指定资源类型的所有对象
func (k *Kubectl) listResources(ctx context.Context, kind string, ns string, opts ...option.ListOption) (resources []*unstructured.Unstructured, err error) {
	gvr, namespaced := k.Tools().GetGVRByKind(kind)
	if gvr.Empty() {
		return nil, fmt.Errorf("不支持的资源类型: %s", kind)
	}

	listOptions := metav1.ListOptions{}
	for _, opt := range opts {
		opt(&listOptions)
	}

	var list *unstructured.UnstructuredList
	if namespaced {
		list, err = k.DynamicClient().Resource(gvr).Namespace(ns).List(ctx, listOptions)
	} else {
		list, err = k.DynamicClient().Resource(gvr).List(ctx, listOptions)
	}
	if err != nil {
		return nil, err
	}
	for _, item := range list.Items {
		obj := item.DeepCopy()
		utils.RemoveManagedFields(obj)
		resources = append(resources, obj)
	}

	return resources, nil
}

func (k *Kubectl) parseGVK2GVR(gvks []schema.GroupVersionKind, versions ...string) (gvr schema.GroupVersionResource, namespaced bool) {
	// 获取单个GVK
	gvk := k.Tools().GetParsedGVK(gvks, versions...)

	// 获取GVR
	if k.Tools().IsBuiltinResource(gvk.Kind) {
		// 内置资源
		return k.Tools().GetGVRByKind(gvk.Kind)
	} else {
		crd, err := k.Tools().GetCRD(gvk.Kind, gvk.Group)
		if err != nil {
			return
		}
		// 检查CRD是否是Namespaced
		namespaced = crd.Object["spec"].(map[string]interface{})["scope"].(string) == "Namespaced"
		gvr = k.Tools().GetGVRFromCRD(crd)
	}

	return
}
