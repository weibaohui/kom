package kom

import (
	"fmt"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
)

type scale struct {
	kubectl *Kubectl
}

func (d *scale) Scale(replicas int32) error {

	kind := d.kubectl.Statement.GVK.Kind
	klog.V(8).Infof("scale Kind=%s", kind)
	klog.V(8).Infof("scale Resource=%s", d.kubectl.Statement.GVR.Resource)
	klog.V(8).Infof("scale %s/%s", d.kubectl.Statement.Namespace, d.kubectl.Statement.Name)

	// 当前支持restart方法的资源有
	// Deployment
	// StatefulSet
	// ReplicaSet
	// ReplicationController

	if !isSupportedKind(kind, []string{"Deployment", "StatefulSet", "ReplicationController", "ReplicaSet"}) {
		d.kubectl.Error = fmt.Errorf("%s %s/%s Scale is not supported", kind, d.kubectl.Statement.Namespace, d.kubectl.Statement.Name)
		return d.kubectl.Error
	}

	var item interface{}
	patchData := fmt.Sprintf("{\"spec\":{\"replicas\":%d}}", replicas)
	err := d.kubectl.Patch(&item, types.MergePatchType, patchData).Error
	if err != nil {
		d.kubectl.Error = fmt.Errorf("%s %s/%s scale error %v", kind, d.kubectl.Statement.Namespace, d.kubectl.Statement.Name, err)
		return err
	}
	return d.kubectl.Error
}
