package kom

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/duke-git/lancet/v2/slice"
	"github.com/weibaohui/kom/utils"
	corev1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
)

type node struct {
	kubectl *Kubectl
}

// Cordon node
// cordon 命令的核心功能是将节点标记为 Unschedulable。在此状态下，调度器（Scheduler）将不会向该节点分配新的 Pod。
func (d *node) Cordon() error {
	var item interface{}
	patchData := `{"spec":{"unschedulable":true}}`
	err := d.kubectl.Patch(&item, types.MergePatchType, patchData).Error
	return err
}

// UnCordon node
// uncordon 命令是 cordon 的逆操作，用于将节点从不可调度状态恢复为可调度状态。
func (d *node) UnCordon() error {
	var item interface{}
	patchData := `{"spec":{"unschedulable":null}}`
	err := d.kubectl.Patch(&item, types.MergePatchType, patchData).Error
	return err
}
func (d *node) Taint(str string) error {
	taint, err := parseTaint(str)
	if err != nil {
		return err
	}
	var original *corev1.Node
	err = d.kubectl.Get(&original).Error
	if err != nil {
		return err
	}
	taints := original.Spec.Taints
	if taints == nil || len(taints) == 0 {
		taints = []corev1.Taint{*taint}
	} else {
		taints = append(taints, *taint)
	}

	var item interface{}
	patchData := fmt.Sprintf(`{"spec":{"taints":%s}}`, utils.ToJSON(taints))
	err = d.kubectl.Patch(&item, types.MergePatchType, patchData).Error
	return err
}
func (d *node) UnTaint(str string) error {
	taint, err := parseTaint(str)
	if err != nil {
		return err
	}
	var original *corev1.Node
	err = d.kubectl.Get(&original).Error
	if err != nil {
		return err
	}
	taints := original.Spec.Taints
	if taints == nil || len(taints) == 0 {
		return fmt.Errorf("taint %s not found", str)
	}

	taints = slice.Filter(taints, func(index int, item corev1.Taint) bool {
		if item.Key != taint.Key {
			return true
		}
		return false
	})
	var item interface{}
	patchData := fmt.Sprintf(`{"spec":{"taints":%s}}`, utils.ToJSON(taints))
	err = d.kubectl.Patch(&item, types.MergePatchType, patchData).Error
	return err
}

// Drain node
// drain 通常在节点需要进行维护时使用。它不仅会标记节点为不可调度，还会逐一驱逐（Evict）该节点上的所有 Pod。
func (d *node) Drain() error {
	name := d.kubectl.Statement.Name

	// Step 1: 将节点标记为不可调度
	klog.V(8).Infof("node/%s  cordoned\n", name)
	err := d.Cordon()
	if err != nil {
		klog.V(8).Infof("node/%s  cordon error %v\n", name, err.Error())
		return err
	}

	// Step 2: 获取节点上的所有 Pod
	// 列出节点上的pod
	var podList []*corev1.Pod
	err = d.kubectl.newInstance().Resource(&corev1.Pod{}).
		WithFieldSelector(fmt.Sprintf("spec.nodeName=%s", name)).
		List(&podList).Error
	if err != nil {
		klog.V(8).Infof("list pods in node/%s  error %v\n", name, err.Error())
		return err
	}

	// Step 3: 驱逐所有可驱逐的 Pod
	for _, pod := range podList {
		if isDaemonSetPod(pod) || isMirrorPod(pod) {
			// 忽略 DaemonSet 和 Mirror Pod
			klog.V(8).Infof("ignore evict pod  %s/%s  \n", pod.Namespace, pod.Name)
			continue
		}
		klog.V(8).Infof("pod/%s eviction started", pod.Name)

		// 驱逐 Pod
		err := d.evictPod(pod)
		if err != nil {
			klog.V(8).Infof("failed to evict pod %s: %v", pod.Name, err)
			return fmt.Errorf("failed to evict pod %s: %v", pod.Name, err)
		}
		klog.V(8).Infof("pod/%s evictied", pod.Name)
	}

	// Step 4: 等待所有 Pod 被驱逐
	err = wait.PollImmediate(2*time.Second, 5*time.Minute, func() (bool, error) {
		var podList []*corev1.Pod
		err = d.kubectl.newInstance().Resource(&corev1.Pod{}).
			WithFieldSelector(fmt.Sprintf("spec.nodeName=%s", name)).
			List(&podList).Error
		if err != nil {
			klog.V(8).Infof("list pods in node/%s  error %v\n", name, err.Error())
			return false, err
		}
		for _, pod := range podList {
			if isDaemonSetPod(pod) || isMirrorPod(pod) {
				// 忽略 DaemonSet 和 Mirror Pod
				klog.V(8).Infof("ignore evict pod  %s/%s  \n", pod.Namespace, pod.Name)
				continue
			}
			klog.V(8).Infof("pod/%s eviction started", pod.Name)

			// 驱逐 Pod
			err := d.evictPod(pod)
			if err != nil {
				return false, fmt.Errorf("failed to evict pod %s: %v", pod.Name, err)
			}
			klog.V(8).Infof("pod/%s evictied", pod.Name)
		}
		return true, nil
	})
	if err != nil {
		return fmt.Errorf("timeout waiting for pods to be evicted: %w", err)
	}

	klog.V(8).Infof("node/%s drained", name)
	return nil
}

// 检查是否为 DaemonSet 创建的 Pod
func isDaemonSetPod(pod *corev1.Pod) bool {
	for _, owner := range pod.OwnerReferences {
		if owner.Kind == "DaemonSet" {
			return true
		}
	}
	return false
}

// 检查是否为 Mirror Pod
func isMirrorPod(pod *corev1.Pod) bool {
	_, exists := pod.Annotations["kubernetes.io/config.mirror"]
	return exists
}

// 驱逐 Pod
func (d *node) evictPod(pod *corev1.Pod) error {
	klog.V(8).Infof("evicting pod %s/%s \n", pod.Namespace, pod.Name)
	eviction := &policyv1.Eviction{
		ObjectMeta: metav1.ObjectMeta{
			Name:      pod.Name,
			Namespace: pod.Namespace,
		},
	}
	err := d.kubectl.Client().PolicyV1().Evictions(pod.Namespace).Evict(context.TODO(), eviction)

	// err := d.kubectl.newInstance().Resource(eviction).Create(eviction).Error
	if err != nil {
		return err
	}
	klog.V(8).Infof(" pod %s/%s evicted\n", pod.Namespace, pod.Name)
	return nil
}

// ParseTaint parses a taint string into a corev1.Taint structure.
func parseTaint(taintStr string) (*corev1.Taint, error) {
	// Split the input string into key-value-effect
	var key, value, effect string
	parts := strings.Split(taintStr, ":")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid taint format: %s", taintStr)
	}
	keyValue := parts[0]
	effect = parts[1]

	// Check the effect
	if effect != string(corev1.TaintEffectNoSchedule) &&
		effect != string(corev1.TaintEffectPreferNoSchedule) &&
		effect != string(corev1.TaintEffectNoExecute) {
		return nil, fmt.Errorf("invalid taint effect: %s", effect)
	}

	// Parse the key and value
	keyValueParts := strings.SplitN(keyValue, "=", 2)
	key = keyValueParts[0]
	if len(keyValueParts) == 2 {
		value = keyValueParts[1]
	}

	// Return the Taint structure
	return &corev1.Taint{
		Key:    key,
		Value:  value,
		Effect: corev1.TaintEffect(effect),
	}, nil
}
