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
