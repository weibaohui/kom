package kom

import (
	"fmt"
	"time"

	v1 "k8s.io/api/apps/v1"
	"k8s.io/klog/v2"
)

type ctl struct {
	kubectl *Kubectl
}

type deploy struct {
	kubectl *Kubectl
}

func (c *ctl) Deployment() *deploy {
	if c.kubectl.Statement.GVR.Empty() {
		c.kubectl.Statement.Error = fmt.Errorf("请先调用Resource()、CRD()、GVR()等方法")
	}
	return &deploy{
		kubectl: c.kubectl,
	}
}

func (d *deploy) Restart() error {
	var item v1.Deployment

	err := d.kubectl.Get(&item).Error
	if err != nil {
		klog.Errorf("Deployment Restart Get(&item) error :%v", err)
		d.kubectl.Error = err
		return d.kubectl.Error
	}

	if item.Spec.Template.Annotations == nil {
		item.Spec.Template.Annotations = map[string]string{}
	}
	item.Spec.Template.Annotations["kom.kubernetes.io/restartedAt"] = time.Now().Format(time.RFC3339)
	err = d.kubectl.Update(&item).Error
	if err != nil {
		klog.Errorf("Deployment Restart Update(&item) error :%v", err)
		d.kubectl.Error = err
		return d.kubectl.Error
	}
	return d.kubectl.Error
}
