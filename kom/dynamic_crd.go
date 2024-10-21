package kom

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func (k8s *Kom) GetCRD(ctx context.Context, kind string, group string) (*unstructured.Unstructured, error) {
	// todo 放到变量中，避免每次获取
	crdList, err := k8s.ListResources(ctx, "CustomResourceDefinition", "")
	if err != nil {
		return nil, err
	}
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
		return &crd, nil
	}
	return nil, fmt.Errorf("crd %s.%s not found", kind, group)
}

func (k8s *Kom) getGRVFromCRD(crd *unstructured.Unstructured) schema.GroupVersionResource {
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
