package example

import (
	"fmt"
	"testing"

	"github.com/weibaohui/kom/kom"
	"github.com/weibaohui/kom/utils"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestRestart(t *testing.T) {
	name := "nginx-restart"
	item := v1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "default",
			Labels: map[string]string{
				"app": name,
			},
		},
		Spec: v1.DeploymentSpec{
			Replicas: utils.Int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": name,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": name,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  name,
							Image: "nginx:1.14.2",
						},
					},
				},
			},
		},
	}
	// 先创建Deploy
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
				Name(name).
				Get(&target)
			if target.Spec.Template.Spec.Containers[0].Name == name {
				// 达成测试条件
				return true
			}
			return false
		}, interval, timeout) {
		t.Logf("创建Deploy nginx 成功")

	} else {
		t.Errorf("创建Deploy nginx 失败")
		return
	}

	// 重启
	err = kom.DefaultCluster().Resource(&item).
		Namespace("default").Name(name).Ctl().Rollout().Restart()

	if err != nil {
		t.Errorf("Deployment Restart(&item) error :%v", err)
	}
	t.Logf("Restart Deploy nginx 成功")

	// 清理
	err = kom.DefaultCluster().Resource(&item).
		Namespace("default").
		Name(name).
		Delete().Error
	if err != nil {
		t.Errorf("Deployment Restart Clean(&item) error :%v", err)
	}
}
func TestHistory(t *testing.T) {
	name := "nginx-history"
	item := v1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "default",
			Labels: map[string]string{
				"app": name,
			},
		},
		Spec: v1.DeploymentSpec{
			Replicas: utils.Int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": name,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": name,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  name,
							Image: "nginx:1.14.2",
						},
					},
				},
			},
		},
	}
	// 先创建Deploy
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
				Name(name).
				Get(&target)
			if target.Spec.Template.Spec.Containers[0].Name == name {
				// 达成测试条件
				return true
			}
			return false
		}, interval, timeout) {
		t.Logf("创建Deploy nginx 成功")

	} else {
		t.Errorf("创建Deploy nginx 失败")
		return
	}

	// 重启
	err = kom.DefaultCluster().Resource(&item).
		Namespace("default").Name(name).Ctl().Rollout().Restart()

	history, err := kom.DefaultCluster().Resource(&v1.Deployment{}).
		Namespace("default").Name(name).Ctl().Rollout().History()
	t.Log(history)
	t.Log(history)

	if err != nil {
		t.Errorf("Deployment History(&item) error :%v", err)
	}
	t.Logf("History Deploy nginx 成功")

	// 清理
	err = kom.DefaultCluster().Resource(&item).
		Namespace("default").
		Name(name).
		Delete().Error
	if err != nil {
		t.Errorf("Deployment History Clean(&item) error :%v", err)
	}
}

func TestScale(t *testing.T) {
	name := "nginx-scale"
	item := v1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "default",
			Labels: map[string]string{
				"app": name,
			},
		},
		Spec: v1.DeploymentSpec{
			Replicas: utils.Int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": name,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": name,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  name,
							Image: "nginx:1.14.2",
						},
					},
				},
			},
		},
	}
	// 先创建Deploy
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
				Name(name).
				Get(&target)
			if target.Spec.Template.Spec.Containers[0].Name == name {
				// 达成测试条件
				return true
			}
			return false
		}, interval, timeout) {
		t.Logf("创建Deploy nginx 成功")

	} else {
		t.Errorf("创建Deploy nginx 失败")
		return
	}

	// scale
	replicas := int32(2)
	err = kom.DefaultCluster().Resource(&item).Ctl().Scale(replicas)

	if err != nil {
		t.Errorf("Deployment Scale  error :%v", err)
	}

	err = kom.DefaultCluster().Resource(&item).
		Namespace("default").Name(name).Get(&item).Error
	if err != nil {
		t.Errorf("Deployment Get(&item) error :%v", err)
		return
	}
	if *item.Spec.Replicas != replicas {
		t.Errorf("Deployment Scale replicas error :expected=%d,actual=%d", replicas, *item.Spec.Replicas)
		return
	}

	t.Logf("Scale Deploy  成功")

	// 清理
	err = kom.DefaultCluster().Resource(&item).
		Namespace("default").
		Name(name).
		Delete().Error
	if err != nil {
		t.Errorf("Deployment Scale Clean  error :%v", err)
	}
}
func TestReplaceTag(t *testing.T) {
	name := "nginx-replace-tag"
	item := v1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "default",
			Labels: map[string]string{
				"app": name,
			},
		},
		Spec: v1.DeploymentSpec{
			Replicas: utils.Int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": name,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": name,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  name,
							Image: "nginx:1.14.2",
						},
					},
				},
			},
		},
	}
	// 先创建Deploy
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
				Name(name).
				Get(&target)
			if target.Spec.Template.Spec.Containers[0].Name == name {
				// 达成测试条件
				return true
			}
			return false
		}, interval, timeout) {
		t.Logf("创建Deploy nginx 成功")

	} else {
		t.Errorf("创建Deploy nginx 失败")
		return
	}

	// scale
	newTag := "alpine"
	_, err = kom.DefaultCluster().Resource(&item).Ctl().
		Deployment().ReplaceImageTag(name, newTag)

	if err != nil {
		t.Errorf("Deployment ReplaceImageTag  error :%v", err)
	}

	err = kom.DefaultCluster().Resource(&item).
		Namespace("default").Name(name).Get(&item).Error
	if err != nil {
		t.Errorf("Deployment Get(&item) error :%v", err)
		return
	}
	if item.Spec.Template.Spec.Containers[0].Image != fmt.Sprintf("nginx:%s", newTag) {
		t.Errorf("Deployment ReplaceImageTag error :expected=%s,actual=%s", newTag, item.Spec.Template.Spec.Containers[0].Image)
		return
	}

	t.Logf(" Deploy ReplaceImageTag 成功")

	// 清理
	err = kom.DefaultCluster().Resource(&item).
		Namespace("default").
		Name(name).
		Delete().Error
	if err != nil {
		t.Errorf("Deployment ReplaceImageTag Clean  error :%v", err)
	}
}
