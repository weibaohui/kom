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

func (d *rollout) logInfo(action string) {
	kind := d.kubectl.Statement.GVK.Kind
	resource := d.kubectl.Statement.GVR.Resource
	namespace := d.kubectl.Statement.Namespace
	name := d.kubectl.Statement.Name
	klog.V(8).Infof("%s Kind=%s", action, kind)
	klog.V(8).Infof("%s Resource=%s", action, resource)
	klog.V(8).Infof("%s %s/%s", action, namespace, name)
}
func (d *rollout) handleError(kind string, namespace string, name string, action string, err error) error {
	if err != nil {
		d.kubectl.Error = fmt.Errorf("%s %s/%s %s error %v", kind, namespace, name, action, err)
		return err
	}
	return nil
}
func (d *rollout) checkResourceKind(kind string) error {
	if !(kind == "Deployment" || kind == "StatefulSet" || kind == "DaemonSet" || kind == "ReplicaSet") {
		d.kubectl.Error = fmt.Errorf("%s %s/%s operation is not supported", kind, "<namespace>", "<name>")
		return d.kubectl.Error
	}
	return nil
}
func (d *rollout) Restart() error {

	kind := d.kubectl.Statement.GVK.Kind
	d.logInfo("Restart")

	if err := d.checkResourceKind(kind); err != nil {
		return err
	}

	var item interface{}
	patchData := fmt.Sprintf(`{"spec":{"template":{"metadata":{"annotations":{"kom.kubernetes.io/restartedAt":"%s"}}}}}`, time.Now().Format(time.DateTime))
	err := d.kubectl.Patch(&item, types.MergePatchType, patchData).Error
	return d.handleError(kind, d.kubectl.Statement.Namespace, d.kubectl.Statement.Name, "restarting", err)
}
func (d *rollout) Pause() error {
	kind := d.kubectl.Statement.GVK.Kind
	d.logInfo("Restart")

	if err := d.checkResourceKind(kind); err != nil {
		return err
	}

	var item interface{}
	patchData := `{"spec":{"paused":true}}`
	err := d.kubectl.Patch(&item, types.MergePatchType, patchData).Error
	return d.handleError(kind, d.kubectl.Statement.Namespace, d.kubectl.Statement.Name, "pause", err)

}
func (d *rollout) Resume() error {
	kind := d.kubectl.Statement.GVK.Kind
	d.logInfo("Restart")

	if err := d.checkResourceKind(kind); err != nil {
		return err
	}

	var item interface{}
	patchData := `{"spec":{"paused":null}}`
	err := d.kubectl.Patch(&item, types.MergePatchType, patchData).Error
	return d.handleError(kind, d.kubectl.Statement.Namespace, d.kubectl.Statement.Name, "resume", err)

}
