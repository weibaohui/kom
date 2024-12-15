package example

import (
	"testing"

	"github.com/weibaohui/kom/kom"
	"github.com/weibaohui/kom/utils"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestLabel(t *testing.T) {
	item := v1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "nginx-label",
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
		t.Logf("Deployment Create(&item) error :%v", err)
	}

	// 原始label

	err = kom.DefaultCluster().
		Resource(&item).Get(&item).Error
	if err != nil {
		t.Logf("Deployment Get(&item) error :%v", err)
	}
	t.Logf("原始label\n%s\n", utils.ToJSON(item.GetLabels()))

	// 创建后更新label；
	err = kom.DefaultCluster().
		Resource(&item).Ctl().
		Label("zhangsan=lisi")
	if err != nil {
		t.Logf("Deployment update label  error :%v", err)
	}

	// 检查新增label情况
	err = kom.DefaultCluster().
		Resource(&item).Get(&item).Error
	if err != nil {
		t.Logf("Deployment Get(&item) error :%v", err)
	}
	t.Logf("增加label\n%s\n", utils.ToJSON(item.GetLabels()))

	// 删除label；
	err = kom.DefaultCluster().
		Resource(&item).Ctl().
		Label("zhangsan-")
	if err != nil {
		t.Logf("Deployment update label delete error :%v", err)
	}
	// 检查删除label情况
	err = kom.DefaultCluster().
		Resource(&item).Get(&item).Error
	if err != nil {
		t.Logf("Deployment Get(&item) error :%v", err)
	}
	t.Logf("删除label\n%s\n", utils.ToJSON(item.GetLabels()))
}

func TestAnnotate(t *testing.T) {
	item := v1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "nginx-label",
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
		t.Logf("Deployment Create(&item) error :%v", err)
	}

	// 原始label

	err = kom.DefaultCluster().
		Resource(&item).Get(&item).Error
	if err != nil {
		t.Logf("Deployment Get(&item) error :%v", err)
	}
	t.Logf("原始Annotation\n%s\n", utils.ToJSON(item.GetAnnotations()))

	// 创建后更新label；
	err = kom.DefaultCluster().
		Resource(&item).Ctl().
		Annotate("zhangsan=lisi")
	if err != nil {
		t.Logf("Deployment update Annotation  error :%v", err)
	}

	// 检查新增label情况
	err = kom.DefaultCluster().
		Resource(&item).Get(&item).Error
	if err != nil {
		t.Logf("Deployment Get(&item) error :%v", err)
	}
	t.Logf("增加Annotation\n%s\n", utils.ToJSON(item.GetAnnotations()))

	// 删除Annotation；
	err = kom.DefaultCluster().
		Resource(&item).Ctl().
		Annotate("zhangsan-")
	if err != nil {
		t.Logf("Deployment update Annotation delete error :%v", err)
	}
	// 检查删除Annotation情况
	err = kom.DefaultCluster().
		Resource(&item).Get(&item).Error
	if err != nil {
		t.Logf("Deployment Get(&item) error :%v", err)
	}
	t.Logf("删除Annotation\n%s\n", utils.ToJSON(item.GetAnnotations()))
}
