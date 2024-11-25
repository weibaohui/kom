package kom

import (
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
)

type rollout struct {
	kubectl *Kubectl
}

func (d *rollout) Restart() error {

	kind := d.kubectl.Statement.GVK.Kind
	klog.V(8).Infof("Restart Kind=%s", kind)
	klog.V(8).Infof("Restart Resource=%s", d.kubectl.Statement.GVR.Resource)
	klog.V(8).Infof("Restart %s/%s", d.kubectl.Statement.Namespace, d.kubectl.Statement.Name)

	// 当前支持restart方法的资源有
	// Deployment
	// StatefulSet
	// DaemonSet
	// ReplicaSet
	if !(kind == "Deployment" || kind == "StatefulSet" || kind == "DaemonSet" || kind == "ReplicaSet") {
		d.kubectl.Error = fmt.Errorf("%s %s/%s restarting is not supported", kind, d.kubectl.Statement.Namespace, d.kubectl.Statement.Name)
		return d.kubectl.Error
	}

	var item interface{}
	patchData := fmt.Sprintf(`{"spec":{"template":{"metadata":{"annotations":{"kom.kubernetes.io/restartedAt":"%s"}}}}}`, time.Now().Format(time.DateTime))
	err := d.kubectl.Patch(&item, types.MergePatchType, patchData).Error
	if err != nil {
		d.kubectl.Error = fmt.Errorf("%s %s/%s restarting error %v", kind, d.kubectl.Statement.Namespace, d.kubectl.Statement.Name, err)
		return err
	}
	return d.kubectl.Error
}
