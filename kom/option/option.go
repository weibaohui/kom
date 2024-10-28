package option

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ListOption Functional options for listResources
type ListOption func(*metav1.ListOptions)

// WithLabelSelector 设置 LabelSelector
func WithLabelSelector(labelSelector string) ListOption {
	return func(lo *metav1.ListOptions) {
		if lo.LabelSelector != "" {
			lo.LabelSelector += "," + labelSelector
		} else {
			lo.LabelSelector = labelSelector
		}
	}
}

// WithFieldSelector 设置 FieldSelector
func WithFieldSelector(fieldSelector string) ListOption {
	return func(lo *metav1.ListOptions) {
		lo.FieldSelector = fieldSelector
	}
}
