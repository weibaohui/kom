package example

import (
	"context"
	"fmt"
	"time"

	"github.com/weibaohui/kom/kom"
	"github.com/weibaohui/kom/kom/applier"
	"github.com/weibaohui/kom/utils"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
)

func Example() {
	builtInExample()
	crdExample()
	YamlApplyDelete()
}
func YamlApplyDelete() {
	yaml := `apiVersion: v1
kind: ConfigMap
metadata:
  name: example-config
  namespace: default
data:
  key: value
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: example-deployment
  namespace: default
spec:
  replicas: 1
  selector:
    matchLabels:
      app: example
  template:
    metadata:
      labels:
        app: example
    spec:
      containers:
        - name: example-container
          image: nginx
`
	result := applier.Instance().WithContext(context.TODO()).Apply(yaml)
	for _, r := range result {
		fmt.Println(r)
	}
	result = applier.Instance().WithContext(context.TODO()).Apply(yaml)
	for _, r := range result {
		fmt.Println(r)
	}
	result = applier.Instance().WithContext(context.TODO()).Delete(yaml)
	for _, r := range result {
		fmt.Println(r)
	}
}
func crdExample() {

	var crontab unstructured.Unstructured
	crontab = unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "stable.example.com/v1",
			"kind":       "CronTab",
			"metadata": map[string]interface{}{
				"name":      "test-crontab",
				"namespace": "default",
			},
			"spec": map[string]interface{}{
				"cronSpec": "* * * * */8",
				"image":    "test-crontab-image",
			},
		},
	}
	// 删除CRD
	err := kom.Init().
		WithContext(context.TODO()).
		CRD("stable.example.com", "v1", "CronTab").
		Name(crontab.GetName()).
		Namespace(crontab.GetNamespace()).
		Delete().Error
	if err != nil {
		klog.Errorf("CronTab Delete(&item) error :%v", err)
	}
	err = kom.Init().
		WithContext(context.TODO()).
		CRD("stable.example.com", "v1", "CronTab").
		Name(crontab.GetName()).
		Namespace(crontab.GetNamespace()).
		Create(&crontab).Error
	if err != nil {
		fmt.Printf("CRD Get %v\n", err)
	}
	err = kom.Init().
		WithContext(context.TODO()).
		CRD("stable.example.com", "v1", "CronTab").
		Name(crontab.GetName()).
		Namespace(crontab.GetNamespace()).
		Get(&crontab).Error
	if err != nil {
		fmt.Printf("CRD Get %v\n", err)
	}

	var crontabList []unstructured.Unstructured
	err = kom.Init().
		WithContext(context.TODO()).
		CRD("stable.example.com", "v1", "CronTab").
		Namespace(crontab.GetNamespace()).
		List(&crontabList).Error
	fmt.Printf("CRD List  count %d\n", len(crontabList))
	for _, d := range crontabList {
		fmt.Printf("CRD  List Items foreach %s\n", d.GetName())
	}

	// 定义 Patch 内容
	patchData := `{
    "spec": {
        "image": "patch-image"
    },
    "metadata": {
        "labels": {
            "new-label": "new-value"
        }
    }
}`
	err = kom.Init().
		WithContext(context.TODO()).
		CRD("stable.example.com", "v1", "CronTab").
		Name(crontab.GetName()).
		Namespace(crontab.GetNamespace()).
		Get(&crontab).Error
	if err != nil {
		klog.Errorf("CronTab Get(&item) error :%v", err)
	}
	err = kom.Init().
		WithContext(context.TODO()).
		CRD("stable.example.com", "v1", "CronTab").
		Name(crontab.GetName()).
		Namespace(crontab.GetNamespace()).
		PatchData(patchData).
		PatchType(types.MergePatchType).
		Patch(&crontab).Error

	if err != nil {
		klog.Errorf("CronTab Patch(&item) error :%v", err)
	}
}
func builtInExample() {
	item := v1.Deployment{}
	err := kom.Init().
		WithContext(context.TODO()).
		Resource(&item).
		Namespace("default").
		Name("ci-755702-codexxx").
		Get(&item).Error
	if err != nil {
		klog.Errorf("Deployment Get(&item) error :%v", err)
	}
	fmt.Printf("Get Item %s\n", item.Spec.Template.Spec.Containers[0].Image)

	createItem := v1.Deployment{

		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-deploy",
			Namespace: "default",
		},
		Spec: v1.DeploymentSpec{
			Replicas: utils.Int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "test",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "test",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "test",
							Image: "nginx:1.14.2",
						},
					},
				},
			},
		},
	}
	err = kom.Init().
		WithContext(context.TODO()).
		Resource(&createItem).
		Create(&createItem).Error
	if err != nil {
		klog.Errorf("Deployment Create(&item) error :%v", err)
	}
	err = kom.Init().
		WithContext(context.TODO()).
		Resource(&createItem).
		Namespace(createItem.Namespace).
		Name(createItem.Name).
		Get(&createItem).Error
	if err != nil {
		klog.Errorf("Deployment Get(&item) error :%v", err)
	}
	if createItem.Spec.Template.Annotations == nil {
		createItem.Spec.Template.Annotations = map[string]string{}
	}
	createItem.Spec.Template.Annotations["kom.kubernetes.io/restartedAt"] = time.Now().Format(time.RFC3339)
	err = kom.Init().
		WithContext(context.TODO()).
		Resource(&createItem).
		Update(&createItem).Error
	if err != nil {
		klog.Errorf("Deployment Update(&item) error :%v", err)
	}
	// 定义 Patch 内容
	patchData := `{
    "spec": {
        "replicas": 5
    },
    "metadata": {
        "labels": {
            "new-label": "new-value"
        }
    }
}`
	err = kom.Init().
		WithContext(context.TODO()).
		Resource(&createItem).
		Namespace(createItem.Namespace).
		Name(createItem.Name).
		Get(&createItem).Error
	err = kom.Init().
		WithContext(context.TODO()).
		Resource(&createItem).
		PatchData(patchData).
		PatchType(types.MergePatchType).
		Patch(&createItem).Error
	if err != nil {
		klog.Errorf("Deployment Patch(&item) error :%v", err)
	}
	err = kom.Init().
		WithContext(context.TODO()).
		Resource(&createItem).
		Namespace(createItem.Namespace).
		Name(createItem.Name).
		Delete().Error

	var items []v1.Deployment
	err = kom.Init().
		WithContext(context.TODO()).
		Resource(&item).
		Namespace("default").
		List(&items).Error
	if err != nil {
		fmt.Printf("List Error %v\n", err)
	}
	fmt.Printf("List Deployment count %d\n", len(items))
	for _, d := range items {
		fmt.Printf("List Deployment  Items foreach %s,%s\n", d.Namespace, d.Spec.Template.Spec.Containers[0].Image)
	}

	err = kom.Init().
		WithContext(context.TODO()).
		Resource(&item).
		Namespace("default").
		List(&items, metav1.ListOptions{LabelSelector: "app=nginx"}).Error
	if err != nil {
		fmt.Printf("List Error %v\n", err)
	}
	fmt.Printf("List Deployment WithLabelSelector app=nginx count =%v \n", len(items))
	for _, d := range items {
		fmt.Printf("List Deployment WithLabelSelector Items foreach %s,%s\n", d.Namespace, d.Spec.Template.Spec.Containers[0].Image)
	}
}
