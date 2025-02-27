package kom

import (
	"fmt"

	v1 "k8s.io/api/apps/v1"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
)

type replicaSet struct {
	kubectl *Kubectl
}

func (r *replicaSet) Restart() error {
	return r.kubectl.Ctl().Rollout().Restart()
}
func (r *replicaSet) Scale(replicas int32) error {
	return r.kubectl.Ctl().Scale(replicas)
}
func (r *replicaSet) Stop() error {
	return r.kubectl.Ctl().Scaler().Stop()
}
func (r *replicaSet) Restore() error {
	return r.kubectl.Ctl().Scaler().Restore()
}

func (r *replicaSet) ManagedPods() ([]*corev1.Pod, error) {
	// 先找到rs
	var rs v1.ReplicaSet
	err := r.kubectl.WithCache(r.kubectl.Statement.CacheTTL).Resource(&rs).Get(&rs).Error

	if err != nil {
		return nil, err
	}
	// 通过rs 获取pod
	var podList []*corev1.Pod
	err = r.kubectl.newInstance().WithCache(r.kubectl.Statement.CacheTTL).Resource(&corev1.Pod{}).
		Namespace(r.kubectl.Statement.Namespace).
		Where(fmt.Sprintf("metadata.ownerReferences.name='%s' and metadata.ownerReferences.kind='%s'", rs.GetName(), "ReplicaSet")).
		List(&podList).Error
	return podList, err
}
func (r *replicaSet) ManagedPod() (*corev1.Pod, error) {
	podList, err := r.ManagedPods()
	if err != nil {
		return nil, err
	}
	if len(podList) > 0 {
		return podList[0], nil
	}
	return nil, fmt.Errorf("未发现ReplicaSet[%s]下的Pod", r.kubectl.Statement.Name)
}
func (r *replicaSet) HPAList() ([]*autoscalingv2.HorizontalPodAutoscaler, error) {
	// 通过rs 获取pod
	var list []*autoscalingv2.HorizontalPodAutoscaler
	err := r.kubectl.newInstance().WithCache(r.kubectl.Statement.CacheTTL).Resource(&autoscalingv2.HorizontalPodAutoscaler{}).
		Namespace(r.kubectl.Statement.Namespace).
		Where(fmt.Sprintf("spec.scaleTargetRef.name='%s' and spec.scaleTargetRef.kind='%s'", r.kubectl.Statement.Name, "ReplicaSet")).
		List(&list).Error
	return list, err
}
