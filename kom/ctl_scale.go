package kom

import (
	"fmt"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
)

type scale struct {
	kubectl *Kubectl
}

func (s *scale) Scale(replicas int32) error {

	kind := s.kubectl.Statement.GVK.Kind
	klog.V(8).Infof("scale Kind=%s", kind)
	klog.V(8).Infof("scale Resource=%s", s.kubectl.Statement.GVR.Resource)
	klog.V(8).Infof("scale %s/%s", s.kubectl.Statement.Namespace, s.kubectl.Statement.Name)

	// 当前支持restart方法的资源有
	// Deployment
	// StatefulSet
	// ReplicaSet
	// ReplicationController

	if !isSupportedKind(kind, []string{"Deployment", "StatefulSet", "ReplicationController", "ReplicaSet"}) {
		s.kubectl.Error = fmt.Errorf("%s %s/%s Scale is not supported", kind, s.kubectl.Statement.Namespace, s.kubectl.Statement.Name)
		return s.kubectl.Error
	}

	var item interface{}
	patchData := fmt.Sprintf("{\"spec\":{\"replicas\":%d}}", replicas)
	err := s.kubectl.Patch(&item, types.MergePatchType, patchData).Error
	if err != nil {
		s.kubectl.Error = fmt.Errorf("%s %s/%s scale error %v", kind, s.kubectl.Statement.Namespace, s.kubectl.Statement.Name, err)
		return err
	}
	return s.kubectl.Error
}
