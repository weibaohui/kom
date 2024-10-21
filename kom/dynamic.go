package kom

import (
	"context"
	"fmt"
	"sort"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// ListOption Functional options for ListResources
type ListOption func(*metav1.ListOptions)

// WithLabelSelector 设置 LabelSelector
func WithLabelSelector(labelSelector string) ListOption {
	return func(lo *metav1.ListOptions) {
		lo.LabelSelector = labelSelector
	}
}

// WithFieldSelector 设置 FieldSelector
func WithFieldSelector(fieldSelector string) ListOption {
	return func(lo *metav1.ListOptions) {
		lo.FieldSelector = fieldSelector
	}
}

func (k8s *Kom) ListResources(ctx context.Context, kind string, ns string, opts ...ListOption) ([]unstructured.Unstructured, error) {
	gvr, namespaced := k8s.GetGVR(kind)
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
		list, err = k8s.dynamicClient.Resource(gvr).Namespace(ns).List(ctx, listOptions)
	} else {
		list, err = k8s.dynamicClient.Resource(gvr).List(ctx, listOptions)
	}
	if err != nil {
		return nil, err
	}
	var resources []unstructured.Unstructured
	for _, item := range list.Items {
		obj := item.DeepCopy()
		k8s.RemoveManagedFields(obj)
		resources = append(resources, *obj)
	}

	return sortByCreationTime(resources), nil
}

// RemoveManagedFields 删除 unstructured.Unstructured 对象中的 metadata.managedFields 字段
func (k8s *Kom) RemoveManagedFields(obj *unstructured.Unstructured) {
	// 获取 metadata
	metadata, found, err := unstructured.NestedMap(obj.Object, "metadata")
	if err != nil || !found {
		return
	}

	// 删除 managedFields
	delete(metadata, "managedFields")

	// 更新 metadata
	err = unstructured.SetNestedMap(obj.Object, metadata, "metadata")
	if err != nil {
		return
	}
}

// sortByCreationTime 按创建时间排序资源
func sortByCreationTime(items []unstructured.Unstructured) []unstructured.Unstructured {
	sort.Slice(items, func(i, j int) bool {
		ti := items[i].GetCreationTimestamp()
		tj := items[j].GetCreationTimestamp()
		return ti.After(tj.Time)
	})
	return items
}
