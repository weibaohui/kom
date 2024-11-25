package kom

import (
	"strings"

	v1 "k8s.io/api/apps/v1"
)

type deploy struct {
	kubectl *Kubectl
}

func (d *deploy) Restart() error {
	return d.kubectl.Ctl().Rollout().Restart()
}
func (d *deploy) Scale(replicas int32) error {
	return d.kubectl.Ctl().Scale(replicas)
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
