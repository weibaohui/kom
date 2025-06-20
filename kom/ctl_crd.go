package kom

import (
	"fmt"

	autoscalingv2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type crd struct {
	kubectl *Kubectl
}

func (c *crd) ManagedPods() ([]*corev1.Pod, error) {
	var item *unstructured.Unstructured
	err := c.kubectl.Get(&item).Error
	if err != nil {
		return nil, err
	}
	// 通过rs 获取pod
	var podList []*corev1.Pod
	err = c.kubectl.newInstance().WithContext(c.kubectl.Statement.Context).WithCache(c.kubectl.Statement.CacheTTL).Resource(&corev1.Pod{}).
		Namespace(c.kubectl.Statement.Namespace).
		Where(fmt.Sprintf("metadata.ownerReferences.name='%s' and metadata.ownerReferences.kind='%s'", item.GetName(), item.GetKind())).
		List(&podList).Error
	return podList, err
}
func (c *crd) ManagedPod() (*corev1.Pod, error) {
	podList, err := c.ManagedPods()
	if err != nil {
		return nil, err
	}
	if len(podList) > 0 {
		return podList[0], nil
	}
	return nil, fmt.Errorf("未发现 CRD [%s]下的Pod", c.kubectl.Statement.GVK)
}

func (c *crd) HPAList() ([]*autoscalingv2.HorizontalPodAutoscaler, error) {
	// 通过rs 获取pod
	var list []*autoscalingv2.HorizontalPodAutoscaler
	err := c.kubectl.newInstance().WithCache(c.kubectl.Statement.CacheTTL).
		GVK("autoscaling", "v2", "HorizontalPodAutoscaler").
		Resource(&autoscalingv2.HorizontalPodAutoscaler{}).
		Namespace(c.kubectl.Statement.Namespace).
		Where(fmt.Sprintf("spec.scaleTargetRef.name='%s' and spec.scaleTargetRef.kind='%s'", c.kubectl.Statement.Name, c.kubectl.Statement.GVK.Kind)).
		List(&list).Error
	return list, err
}
