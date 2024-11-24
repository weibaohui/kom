package kom

import (
	"fmt"

	v1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
)

type statefulSet struct {
	kubectl *Kubectl
}

func (s *statefulSet) Scale(replicas int32) error {
	var item v1.StatefulSet

	err := s.kubectl.Get(&item).Error
	if err != nil {
		klog.Errorf("StatefulSet Scale Get %s/%s error :%v", s.kubectl.Statement.Namespace, s.kubectl.Statement.Name, err)
		s.kubectl.Error = err
		return s.kubectl.Error
	}
	patchData := fmt.Sprintf("{\"spec\":{\"replicas\":%d}}", replicas)
	err = s.kubectl.Resource(&item).
		Patch(&item, types.MergePatchType, patchData).Error
	if err != nil {
		klog.Errorf("StatefulSet Scale %s/%s error :%v", s.kubectl.Statement.Namespace, s.kubectl.Statement.Name, err)
		s.kubectl.Error = err
		return err
	}
	return s.kubectl.Error
}
