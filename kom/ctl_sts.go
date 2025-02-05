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

func (r *statefulSet) ManagedPods() ([]*corev1.Pod, error) {
	//先找到sts
	var sts v1.StatefulSet
	err := r.kubectl.Resource(&sts).Get(&sts).Error

	if err != nil {
		return nil, err
	}
	// 通过sts 获取pod
	var podList []*corev1.Pod
	err = r.kubectl.newInstance().Resource(&corev1.Pod{}).
		Namespace(r.kubectl.Statement.Namespace).
		Where(fmt.Sprintf("metadata.ownerReferences.name='%s' and metadata.ownerReferences.kind='%s'", sts.GetName(), "StatefulSet")).
		List(&podList).Error
	return podList, err
}
func (r *statefulSet) ManagedPod() (*corev1.Pod, error) {
	podList, err := r.ManagedPods()
	if err != nil {
		return nil, err
	}
	if len(podList) > 0 {
		return podList[0], nil
	}
	return nil, fmt.Errorf("未发现StatefulSet[%s]下的Pod", r.kubectl.Statement.Name)
}
