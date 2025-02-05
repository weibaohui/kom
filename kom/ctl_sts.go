package kom

import (
	"fmt"

	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

type statefulSet struct {
	kubectl *Kubectl
}

func (s *statefulSet) Restart() error {
	return s.kubectl.Ctl().Rollout().Restart()
}
func (s *statefulSet) Scale(replicas int32) error {
	return s.kubectl.Ctl().Scale(replicas)
}

func (s *statefulSet) Stop() error {
	return s.kubectl.Ctl().Scaler().Stop()
}
func (s *statefulSet) Restore() error {
	return s.kubectl.Ctl().Scaler().Restore()
}

func (s *statefulSet) ManagedPods() ([]*corev1.Pod, error) {
	//先找到sts
	var sts v1.StatefulSet
	err := s.kubectl.WithCache(s.kubectl.Statement.CacheTTL).Resource(&sts).Get(&sts).Error

	if err != nil {
		return nil, err
	}
	// 通过sts 获取pod
	var podList []*corev1.Pod
	err = s.kubectl.newInstance().WithCache(s.kubectl.Statement.CacheTTL).Resource(&corev1.Pod{}).
		Namespace(s.kubectl.Statement.Namespace).
		Where(fmt.Sprintf("metadata.ownerReferences.name='%s' and metadata.ownerReferences.kind='%s'", sts.GetName(), "StatefulSet")).
		List(&podList).Error
	return podList, err
}
func (s *statefulSet) ManagedPod() (*corev1.Pod, error) {
	podList, err := s.ManagedPods()
	if err != nil {
		return nil, err
	}
	if len(podList) > 0 {
		return podList[0], nil
	}
	return nil, fmt.Errorf("未发现StatefulSet[%s]下的Pod", s.kubectl.Statement.Name)
}
