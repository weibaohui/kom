package kom

import (
	"context"
	"fmt"

	"github.com/weibaohui/kom/kom/option"
	"github.com/weibaohui/kom/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func (k8s *Kom) ListResources(ctx context.Context, kind string, ns string, opts ...option.ListOption) ([]unstructured.Unstructured, error) {
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
		utils.RemoveManagedFields(obj)
		resources = append(resources, *obj)
	}

	return resources, nil
}
