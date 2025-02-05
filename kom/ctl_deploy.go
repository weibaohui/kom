package kom

import (
	"fmt"
	"strings"

	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

type deploy struct {
	kubectl *Kubectl
}

func (d *deploy) Stop() error {
	return d.kubectl.Ctl().Scaler().Stop()
}
func (d *deploy) Restore() error {
	return d.kubectl.Ctl().Scaler().Restore()
}
func (d *deploy) Restart() error {
	return d.kubectl.Ctl().Rollout().Restart()
}
func (d *deploy) Scale(replicas int32) error {
	return d.kubectl.Ctl().Scale(replicas)
}
func (d *deploy) ManagedPods() ([]*corev1.Pod, error) {
	//先找到rs
	rs, err := d.ManagedLatestReplicaSet()
	if err != nil {
		return nil, err
	}
	// 通过rs 获取pod
	var podList []*corev1.Pod
	err = d.kubectl.newInstance().Resource(&corev1.Pod{}).
		Namespace(d.kubectl.Statement.Namespace).
		Where(fmt.Sprintf("metadata.ownerReferences.name='%s' and metadata.ownerReferences.kind='%s'", rs.GetName(), "ReplicaSet")).
		List(&podList).Error
	return podList, err
}
func (d *deploy) ManagedPod() (*corev1.Pod, error) {
	podList, err := d.ManagedPods()
	if err != nil {
		return nil, err
	}
	if len(podList) > 0 {
		return podList[0], nil
	}
	return nil, fmt.Errorf("未发现Deployment[%s]下的Pod", d.kubectl.Statement.Name)
}

// 最新部署版本的RS
func (d *deploy) ManagedLatestReplicaSet() (*v1.ReplicaSet, error) {
	var item v1.Deployment
	err := d.kubectl.Resource(&item).Get(&item).Error

	if err != nil {
		return nil, err
	}

	var rsList []*v1.ReplicaSet
	err = d.kubectl.newInstance().Resource(&v1.ReplicaSet{}).
		Namespace(d.kubectl.Statement.Namespace).
		Where(fmt.Sprintf("metadata.ownerReferences.name='%s' and metadata.ownerReferences.kind='%s'", d.kubectl.Statement.Name, "Deployment")).
		List(&rsList).Error
	if err != nil {
		return nil, err
	}

	//先看有几个rs，如果有一个，那么就这一个了
	if len(rsList) == 1 {
		return rsList[0], nil
	}

	//如果有多个rs，那么需要通过 revision 过滤rs list
	//寻找Deploy上的注解
	//metadata:
	//   annotations:
	//     deployment.kubernetes.io/revision: "50"
	var revision string
	for _, v := range item.GetAnnotations() {
		if strings.HasPrefix(v, "deployment.kubernetes.io/revision") {
			// 找到 revision
			// deployment.kubernetes.io/revision: "50"
			// 50
			s := strings.Split(v, ":")
			if len(s) == 2 {
				revision = s[1]
				break
			}
		}
	}

	for _, rs := range rsList {
		if rs.Annotations["deployment.kubernetes.io/revision"] == revision {
			return rs, nil
		}
	}
	return nil, fmt.Errorf("未发现Deployment[%s]下的最新的RS", item.GetName())
}

func (d *deploy) ReplaceImageTag(targetContainerName string, tag string) (*v1.Deployment, error) {
	var item v1.Deployment
	err := d.kubectl.Resource(&item).Get(&item).Error

	if err != nil {
		return nil, err
	}

	for i := range item.Spec.Template.Spec.Containers {
		c := &item.Spec.Template.Spec.Containers[i]
		if c.Name == targetContainerName {
			c.Image = replaceImageTag(c.Image, tag)
		}
	}
	err = d.kubectl.Resource(&item).Update(&item).Error
	return &item, err
}

// replaceImageTag 替换镜像的 tag
func replaceImageTag(imageName, newTag string) string {
	// 检查镜像名称是否包含 tag
	if strings.Contains(imageName, ":") {
		// 按照 ":" 分割镜像名称和 tag
		parts := strings.Split(imageName, ":")
		// 使用新的 tag 替换旧的 tag
		return parts[0] + ":" + newTag
	} else {
		// 如果镜像名称中没有 tag，直接添加新的 tag
		return imageName + ":" + newTag
	}
}
