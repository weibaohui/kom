package kom

import (
	"fmt"
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

	err := d.kubectl.Get(&item).Error
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
