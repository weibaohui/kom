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
		err := kom.DefaultCluster().Resource(&item).
			Namespace("default").
			Name("nginx").
			Delete().Error
		if err != nil {
			t.Errorf("Deployment Clean(&item) error :%v", err)
		}
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
		t.Errorf(" Patch(&item) error :%v", err)
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
	if len(items) > 0 {
		t.Logf("List Pods count %d\n", len(items))
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
		WithLabelSelector("app=random").
		List(&items).Error
	if err != nil {
		t.Errorf("List Error %v\n", err)
	}
	if len(items) == 1 {
		t.Logf("List Pods count  [app=random] :%d\n", len(items))
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
		WithLabelSelector("app=random").
		WithLabelSelector("x=y").
		List(&items).Error
	if err != nil {
		t.Errorf("List Error %v\n", err)
	}
	if len(items) == 1 {
		t.Logf("List Pods count  [app=random,x=y] :%d\n", len(items))
	} else {
		t.Errorf("List Pods count,should %d,acctual %d", 1, len(items))
	}
}
func TestListPodByMultiLabelSelector2(t *testing.T) {
	var items []corev1.Pod
	var pod corev1.Pod
	err := kom.DefaultCluster().
		Resource(&pod).
		Namespace("default").
		WithLabelSelector("app=random").
		WithLabelSelector("x=y2").
		List(&items).Error
	if err != nil {
		t.Errorf("List Error %v\n", err)
	}
	if len(items) == 0 {
		t.Logf("List Pods count  [app=random,x=y2] :%d\n", len(items))
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
		WithFieldSelector("metadata.name=random").
		List(&items).Error
	if err != nil {
		t.Errorf("List Error %v\n", err)
	}
	if len(items) == 1 {
		t.Logf("List Pods count  [metadata.name=random] :%d\n", len(items))
	} else {
		t.Errorf("List Pods count,should %d,acctual %d", 1, len(items))
	}
}
func TestListPodByFieldSelectorNotExists(t *testing.T) {
	var items []corev1.Pod
	var pod corev1.Pod
	err := kom.DefaultCluster().
		Resource(&pod).
		Namespace("default").
		WithFieldSelector("metadata.name=random111").
		List(&items).Error
	if err != nil {
		t.Errorf("List Error %v\n", err)
	}
	if len(items) == 0 {
		t.Logf("List Pods count [metadata.name=random111] :%d\n", len(items))
	} else {
		t.Errorf("List Pods count,should %d,acctual %d", 1, len(items))
	}
}
func TestDelete(t *testing.T) {
	// 先创建一个pod，然后删除

	yaml := `apiVersion: v1
kind: Pod
metadata:
  name: delete-test
  namespace: default
spec:
  containers:
  - args:
    - |
      mkdir -p /var/log;
      while true; do
        random_char="A$(cat /dev/urandom | tr -dc 'a-zA-Z0-9' | head -c 1)";
        echo $random_char | tee -a /var/log/random_a.log;
        sleep 5;
      done
    command:
    - /bin/sh
    - -c
    image: alpine
    name: delete-test
`
	result := kom.DefaultCluster().Applier().Apply(yaml)
	for _, s := range result {
		t.Logf("%s\n", s)
	}

	// 等待创建成功
	if utils.WaitUntil(
		func() bool {
			var target corev1.Pod
			err := kom.DefaultCluster().Resource(&target).Namespace("default").
				Name("delete-test").Get(&target).Error
			if err != nil {
				return false
			}
			if target.Status.Phase == "Running" {
				fmt.Println("delete-test is running at", time.Now())
				return true
			}
			return false
		}, interval, timeout) {
		t.Logf("创建 Pod delete-test 成功")
	} else {
		t.Errorf("创建 Pod delete-test 失败")
	}

	// 删除
	var pod corev1.Pod
	err := kom.DefaultCluster().
		Resource(&pod).
		Namespace("default").
		Name("delete-test").
		Delete().Error
	if err != nil {
		t.Errorf("Delete Pod Error %v\n", err)
	}
	t.Logf("已经执行删除Pod delete-test 命令")
	if utils.WaitUntil(
		func() bool {
			t.Logf("尝试获取 Pod delete-test ")
			var target corev1.Pod
			kom.DefaultCluster().Resource(&target).
				Namespace("default").
				Name("delete-test").
				Get(&target)
			t.Logf("尝试获取 Pod delete-test Name= %s,status=%s ", target.Name, target.Status.Phase)

			if target.Name == "" {
				return true
			}

			return false
		}, interval, timeout) {
		t.Logf("删除 Pod delete-test 成功")
	} else {
		t.Errorf("删除 Pod delete-test 失败")
	}

}
