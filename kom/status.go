package kom

import (
	"context"
	"strings"

	"github.com/google/gnostic-models/openapiv2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/klog/v2"
)

// // 获取版本信息
// versionInfo, err := clientset.Discovery().ServerVersion()
// if err != nil {
// log.Fatalf("Error getting server version: %s", err.Error())
// }
func (k *Kom) GetOpenAPISchema() *openapi_v2.Document {
	openAPISchema, err := k.Client().DiscoveryClient.OpenAPISchema()
	if err != nil {
		klog.V(2).Infof("Error fetching OpenAPI schema: %v\n", err)
		return nil
	}
	return openAPISchema
}

func (k *Kom) initializeCRDList() []*unstructured.Unstructured {
	crdList, _ := k.ListResources(context.TODO(), "CustomResourceDefinition", "")
	return crdList
}
func (k *Kom) initializeAPIResources() (apiResources []*metav1.APIResource) {
	// 提取ApiResources
	_, lists, _ := k.Client().Discovery().ServerGroupsAndResources()
	for _, list := range lists {
		resources := list.APIResources
		version := list.GroupVersionKind().Version
		group := list.GroupVersionKind().Group
		groupVersion := list.GroupVersion
		gvs := strings.Split(groupVersion, "/")
		if len(gvs) == 2 {
			group = gvs[0]
			version = gvs[1]
		} else {
			// 只有version的情况"v1"
			version = groupVersion
		}

		for _, resource := range resources {
			resource.Group = group
			resource.Version = version
			apiResources = append(apiResources, &resource)
		}
	}
	return apiResources
}