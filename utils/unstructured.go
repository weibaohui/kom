package utils

import (
	"fmt"
	"sort"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/yaml"
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

// ConvertUnstructuredToYAML 将 Unstructured 对象转换为 YAML 字符串
func ConvertUnstructuredToYAML(obj *unstructured.Unstructured) (string, error) {

	// Marshal Unstructured 对象为 JSON
	jsonBytes, err := obj.MarshalJSON()
	if err != nil {
		return "", fmt.Errorf("无法序列化 Unstructured 对象为 JSON: %v", err)
	}

	// 将 JSON 转换为 YAML
	yamlBytes, err := yaml.JSONToYAML(jsonBytes)
	if err != nil {
		return "", fmt.Errorf("无法将 JSON 转换为 YAML: %v", err)
	}

	return string(yamlBytes), nil
}
