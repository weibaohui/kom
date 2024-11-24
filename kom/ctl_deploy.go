package kom

import (
	"fmt"
	"strings"
	"time"

	v1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
)

type deploy struct {
	kubectl *Kubectl
}

func (d *deploy) Restart() error {
	var item v1.Deployment
	err := d.kubectl.Get(&item).Error
	if err != nil {
		klog.Errorf("Deployment Restart %s/%s error :%v", d.kubectl.Statement.Namespace, d.kubectl.Statement.Name, err)
		d.kubectl.Error = err
		return d.kubectl.Error
	}
	patchData := `{
	"spec": {
		"template": {
			"metadata": {
				"annotations": {
						"kom.kubernetes.io/restartedAt": "%s"
				}
			}
		}
	}
}`
	patchData = fmt.Sprintf(patchData, time.Now().Format(time.DateTime))

	err = d.kubectl.Resource(&item).
		Patch(&item, types.MergePatchType, patchData).Error
	if err != nil {
		klog.Errorf("Deployment Restart %s/%s error :%v", d.kubectl.Statement.Namespace, d.kubectl.Statement.Name, err)
		d.kubectl.Error = err
		return err
	}
	return d.kubectl.Error
}
func (d *deploy) Scale(replicas int32) error {
	var item v1.Deployment

	err := d.kubectl.Resource(&item).Get(&item).Error
	if err != nil {
		klog.Errorf("Deployment Scale Get %s/%s error :%v", d.kubectl.Statement.Namespace, d.kubectl.Statement.Name, err)
		d.kubectl.Error = err
		return d.kubectl.Error
	}
	patchData := fmt.Sprintf("{\"spec\":{\"replicas\":%d}}", replicas)
	err = d.kubectl.Resource(&item).
		Patch(&item, types.MergePatchType, patchData).Error
	if err != nil {
		klog.Errorf("Deployment Scale %s/%s error :%v", d.kubectl.Statement.Namespace, d.kubectl.Statement.Name, err)
		d.kubectl.Error = err
		return err
	}
	return d.kubectl.Error
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
