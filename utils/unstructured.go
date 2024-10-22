package utils

import (
	"sort"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func SortByCreationTime(items []unstructured.Unstructured) []unstructured.Unstructured {
	sort.Slice(items, func(i, j int) bool {
		ti := items[i].GetCreationTimestamp()
		tj := items[j].GetCreationTimestamp()
		return ti.After(tj.Time)
	})
	return items
}

// RemoveManagedFields 删除 unstructured.Unstructured 对象中的 metadata.managedFields 字段
func RemoveManagedFields(obj *unstructured.Unstructured) {
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
