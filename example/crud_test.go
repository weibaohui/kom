package example

import (
	"fmt"
	"testing"
	"time"

	"github.com/weibaohui/kom/kom"
	"github.com/weibaohui/kom/utils"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
)

func TestCreate(t *testing.T) {
	item := v1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "nginx",
			Namespace: "default",
			Labels: map[string]string{
				"app": "nginx",
				"m":   "n",
			},
		},
		Spec: v1.DeploymentSpec{
			Replicas: utils.Int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "nginx",
					"m":   "n",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "nginx",
						"m":   "n",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "nginx",
							Image: "nginx:1.14.2",
						},
					},
				},
			},
		},
	}
	// 创建 test-deploy
	err := kom.DefaultCluster().
		Resource(&item).
		Create(&item).Error
	if err != nil {
		t.Errorf("Deployment Create(&item) error :%v", err)
	}

	if utils.WaitUntil(
		func() bool {
			var target v1.Deployment
			kom.DefaultCluster().Resource(&target).
				Namespace("default").
				Name("nginx").
				Get(&target)
			if target.Spec.Template.Spec.Containers[0].Name == "nginx" {
				// 达成测试条件
				return true
			}
			return false
		}, interval, timeout) {
		t.Logf("创建Deploy nginx 成功")
	} else {
		t.Errorf("创建Deploy nginx 失败")
	}
}
func TestUpdate(t *testing.T) {

	var pod corev1.Pod
	err := kom.DefaultCluster().
		Resource(&pod).
		Namespace("default").
		Name("random").
		Get(&pod).Error
	if err != nil {
		t.Errorf("Deployment Get(&item) error :%v", err)
	}

	// 更新
	if pod.Annotations == nil {
		pod.Annotations = map[string]string{}
	}
	annotation := "kom.kubernetes.io/updatedAt"
	pod.Annotations[annotation] = time.Now().Format(time.RFC3339)
	err = kom.DefaultCluster().
		Resource(&pod).
		Update(&pod).Error
	if err != nil {
		klog.Errorf("Deployment Update(&item) error :%v", err)
	}

	if utils.WaitUntil(
		func() bool {
			var target corev1.Pod
			kom.DefaultCluster().Resource(&target).
				Namespace("default").
				Name("random").
				Get(&target)
			for k, _ := range target.Annotations {
				if k == annotation {
					// 找到设置的标签
					// 达成测试条件
					return true
				}
			}

			return false
		}, interval, timeout) {
		t.Logf("更新Pod random 成功")
	} else {
		t.Errorf("更新Pod random 失败")
	}
}
func TestPatch(t *testing.T) {
	// 定义 Patch 内容
	patchData := `{
    "metadata": {
        "labels": {
            "new-label": "new-value",
            "x": "y"
        }
    }
}`
	var pod corev1.Pod
	// Patch test-deploy
	err := kom.DefaultCluster().
		Resource(&pod).
		Namespace("default").
		Name("random").
		Get(&pod).Error
	err = kom.DefaultCluster().
		Resource(&pod).
		Patch(&pod, types.MergePatchType, patchData).Error
	if err != nil {
		t.Errorf("Deployment Patch(&item) error :%v", err)
	}

	if utils.WaitUntil(
		func() bool {
			var target corev1.Pod
			kom.DefaultCluster().Resource(&target).
				Namespace("default").
				Name("random").
				Get(&target)
			for k, _ := range target.Labels {
				if k == "x" {
					// 找到设置的标签
					// 达成测试条件
					return true
				}
			}

			return false
		}, interval, timeout) {
		t.Logf("更新Pod random 成功")
	} else {
		t.Errorf("更新Pod random 失败")
	}
}
func TestListPod(t *testing.T) {
	var items []corev1.Pod
	var pod corev1.Pod
	err := kom.DefaultCluster().
		Resource(&pod).
		Namespace("default").
		List(&items).Error
	if err != nil {
		t.Errorf("List Error %v\n", err)
	}
	if len(items) == 1 {
		fmt.Printf("List Pods count %d\n", len(items))
	} else {
		t.Errorf("List Pods count,should %d,acctual %d", 1, len(items))
	}
}
func TestListPodByLabelSelector(t *testing.T) {
	var items []corev1.Pod
	var pod corev1.Pod
	err := kom.DefaultCluster().
		Resource(&pod).
		Namespace("default").
		List(&items, metav1.ListOptions{LabelSelector: "app=random"}).Error
	if err != nil {
		t.Errorf("List Error %v\n", err)
	}
	if len(items) == 1 {
		fmt.Printf("List Pods count %d\n", len(items))
	} else {
		t.Errorf("List Pods count,should %d,acctual %d", 1, len(items))
	}
}
func TestListPodByMultiLabelSelector(t *testing.T) {
	var items []corev1.Pod
	var pod corev1.Pod
	err := kom.DefaultCluster().
		Resource(&pod).
		Namespace("default").
		List(&items, metav1.ListOptions{LabelSelector: "app=random,x=y"}).Error
	if err != nil {
		t.Errorf("List Error %v\n", err)
	}
	if len(items) == 1 {
		fmt.Printf("List Pods count %d\n", len(items))
	} else {
		t.Errorf("List Pods count,should %d,acctual %d", 1, len(items))
	}
}
func TestListPodByFieldSelector(t *testing.T) {
	var items []corev1.Pod
	var pod corev1.Pod
	err := kom.DefaultCluster().
		Resource(&pod).
		Namespace("default").
		List(&items, metav1.ListOptions{FieldSelector: "metadata.name=random"}).Error
	if err != nil {
		t.Errorf("List Error %v\n", err)
	}
	if len(items) == 1 {
		fmt.Printf("List Pods count %d\n", len(items))
	} else {
		t.Errorf("List Pods count,should %d,acctual %d", 1, len(items))
	}
}
func TestDelete(t *testing.T) {
	var pod corev1.Pod
	err := kom.DefaultCluster().
		Resource(&pod).
		Namespace("default").
		Name("random").
		Delete().Error
	if err != nil {
		t.Errorf("Delete Pod Error %v\n", err)
	}

	if utils.WaitUntil(
		func() bool {
			var target corev1.Pod
			kom.DefaultCluster().Resource(&target).
				Namespace("default").
				Name("random").
				Get(&target)
			if &target == nil {
				return true
			}

			return false
		}, interval, timeout) {
		t.Logf("删除 Pod random 成功")
	} else {
		t.Errorf("删除 Pod random 失败")
	}

}
