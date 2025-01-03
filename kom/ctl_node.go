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
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
	resourcehelper "k8s.io/kubectl/pkg/util/resource"
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
		return item.Key != taint.Key
	})
	var item interface{}
	patchData := fmt.Sprintf(`{"spec":{"taints":%s}}`, utils.ToJSON(taints))
	err = d.kubectl.Patch(&item, types.MergePatchType, patchData).Error
	return err
}

// Drain node
// drain 通常在节点需要进行维护时使用。它不仅会标记节点为不可调度，还会逐一驱逐（Evict）该节点上的所有 Pod。
func (d *node) Drain() error {
	// todo 增加--force的处理，也就强制驱逐所有pod，即便是不满足PDB
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
func (d *node) RunningPods() ([]*corev1.Pod, error) {
	cacheTime := d.kubectl.Statement.CacheTTL
	if cacheTime == 0 {
		cacheTime = 5 * time.Second
	}
	var podList []*corev1.Pod
	// status.phase!=Succeeded,status.phase!=Failed
	err := d.kubectl.newInstance().Resource(&corev1.Pod{}).
		Where("spec.nodeName=? and status.phase!=Succeeded and status.phase!=Failed", d.kubectl.Statement.Name).
		WithCache(cacheTime).List(&podList).Error
	if err != nil {
		klog.V(6).Infof("list pods in node/%s  error %v\n", d.kubectl.Statement.Name, err.Error())
		return nil, err
	}
	return podList, nil
}
func (d *node) TotalRequestsAndLimits() (map[corev1.ResourceName]resource.Quantity, map[corev1.ResourceName]resource.Quantity) {
	pods, err := d.RunningPods()
	if err != nil {
		klog.V(6).Infof("Get TotalRequestsAndLimits in node/%s  error %v\n", d.kubectl.Statement.Name, err.Error())
		return nil, nil
	}
	return getPodsTotalRequestsAndLimits(pods)
}

// ResourceUsage 获取节点的资源使用情况，包括资源的请求和限制，还有当前使用占比
func (d *node) ResourceUsage() *ResourceUsageResult {

	reqs, limits := d.TotalRequestsAndLimits()
	if reqs == nil || limits == nil {
		return nil
	}
	cacheTime := d.kubectl.Statement.CacheTTL
	if cacheTime == 0 {
		cacheTime = 5 * time.Second
	}
	var n *corev1.Node
	err := d.kubectl.newInstance().Resource(&corev1.Node{}).
		Name(d.kubectl.Statement.Name).WithCache(cacheTime).Get(&n).Error
	if err != nil {
		klog.V(6).Infof("Get ResourceUsage in node/%s  error %v\n", d.kubectl.Statement.Name, err.Error())
		return nil
	}

	allocatable := n.Status.Capacity
	if len(n.Status.Allocatable) > 0 {
		allocatable = n.Status.Allocatable
	}

	klog.V(8).Infof("allocatable=:\n%s", utils.ToJSON(allocatable))
	cpuReqs, cpuLimits, memoryReqs, memoryLimits, ephemeralstorageReqs, ephemeralstorageLimits :=
		reqs[corev1.ResourceCPU], limits[corev1.ResourceCPU], reqs[corev1.ResourceMemory], limits[corev1.ResourceMemory], reqs[corev1.ResourceEphemeralStorage], limits[corev1.ResourceEphemeralStorage]

	// 计算CPU 使用率
	fractionCpuReqs := float64(0)
	fractionCpuLimits := float64(0)
	if allocatable.Cpu().MilliValue() != 0 {
		fractionCpuReqs = float64(cpuReqs.MilliValue()) / float64(allocatable.Cpu().MilliValue()) * 100
		klog.V(6).Infof("cpuReqs=%f，allocatable.Cpu=%f, f=%f \n", float64(cpuReqs.MilliValue()), float64(allocatable.Cpu().MilliValue()), fractionCpuReqs)
		fractionCpuLimits = float64(cpuLimits.MilliValue()) / float64(allocatable.Cpu().MilliValue()) * 100
	}

	// 计算内存 使用率
	fractionMemoryReqs := float64(0)
	fractionMemoryLimits := float64(0)
	if allocatable.Memory().Value() != 0 {
		fractionMemoryReqs = float64(memoryReqs.Value()) / float64(allocatable.Memory().Value()) * 100
		fractionMemoryLimits = float64(memoryLimits.Value()) / float64(allocatable.Memory().Value()) * 100
	}

	// 计算存储 使用率
	fractionEphemeralStorageReqs := float64(0)
	fractionEphemeralStorageLimits := float64(0)
	if allocatable.StorageEphemeral().Value() != 0 {
		fractionEphemeralStorageReqs = float64(ephemeralstorageReqs.Value()) / float64(allocatable.StorageEphemeral().Value()) * 100
		fractionEphemeralStorageLimits = float64(ephemeralstorageLimits.Value()) / float64(allocatable.StorageEphemeral().Value()) * 100
	}

	usageFractions := map[corev1.ResourceName]ResourceUsageFraction{
		corev1.ResourceCPU: {
			RequestFraction: fractionCpuReqs,
			LimitFraction:   fractionCpuLimits,
		},
		corev1.ResourceMemory: {
			RequestFraction: fractionMemoryReqs,
			LimitFraction:   fractionMemoryLimits,
		},
		corev1.ResourceEphemeralStorage: {
			RequestFraction: fractionEphemeralStorageReqs,
			LimitFraction:   fractionEphemeralStorageLimits,
		},
	}
	klog.V(6).Infof("node/%s resource usage\n", d.kubectl.Statement.Name)
	klog.V(6).Infof("%s\t%s (%d%%)\t%s (%d%%)\n",
		corev1.ResourceCPU, cpuReqs.String(), int64(fractionCpuReqs), cpuLimits.String(), int64(fractionCpuLimits))
	klog.V(6).Infof("%s\t%s (%d%%)\t%s (%d%%)\n",
		corev1.ResourceMemory, memoryReqs.String(), int64(fractionMemoryReqs), memoryLimits.String(), int64(fractionMemoryLimits))
	klog.V(6).Infof("%s\t%s (%d%%)\t%s (%d%%)\n",
		corev1.ResourceEphemeralStorage, ephemeralstorageReqs.String(), int64(fractionEphemeralStorageReqs), ephemeralstorageLimits.String(), int64(fractionEphemeralStorageLimits))

	return &ResourceUsageResult{
		Requests:       reqs,
		Limits:         limits,
		Allocatable:    allocatable,
		UsageFractions: usageFractions,
	}
}
func (d *node) ResourceUsageTable() []*ResourceUsageRow {

	usage := d.ResourceUsage()
	data, err := convertToTableData(usage)
	if err != nil {
		klog.V(6).Infof("convertToTableData error %v\n", err.Error())
		return make([]*ResourceUsageRow, 0)
	}
	return data
}

// IPUsage 计算节点上IP数量状态，返回节点IP总数，已用数量，可用数量
func (d *node) IPUsage() (total, used, available int) {
	var n *corev1.Node
	err := d.kubectl.newInstance().Resource(&corev1.Node{}).
		Name(d.kubectl.Statement.Name).WithCache(5 * time.Second).Get(&n).Error
	if err != nil {
		klog.V(6).Infof("Get ResourceUsage in node/%s  error %v\n", d.kubectl.Statement.Name, err.Error())
		return 0, 0, 0
	}
	// 计算总数
	cidr := n.Spec.PodCIDR
	count, err := utils.CidrTotalIPs(cidr)
	if err != nil {
		klog.V(6).Infof("Get ResourceUsage in node/%s  error %v\n", d.kubectl.Statement.Name, err.Error())
		return 0, 0, 0
	}
	total = count

	// 计算PodIP数量，
	var podList []*corev1.Pod
	err = d.kubectl.newInstance().Resource(&corev1.Pod{}).
		Where("spec.nodeName=?", d.kubectl.Statement.Name).
		WithCache(5 * time.Second).List(&podList).Error
	if err != nil {
		klog.V(6).Infof("list pods in node/%s  error %v\n", d.kubectl.Statement.Name, err.Error())
		return 0, 0, 0
	}

	podList = slice.Filter(podList, func(index int, item *corev1.Pod) bool {
		return item.Status.PodIP != ""
	})
	used = len(podList)
	available = total - used
	return

}

func getPodsTotalRequestsAndLimits(podList []*corev1.Pod) (reqs map[corev1.ResourceName]resource.Quantity, limits map[corev1.ResourceName]resource.Quantity) {
	reqs, limits = map[corev1.ResourceName]resource.Quantity{}, map[corev1.ResourceName]resource.Quantity{}
	for _, pod := range podList {
		podReqs, podLimits := resourcehelper.PodRequestsAndLimits(pod)
		for podReqName, podReqValue := range podReqs {
			if value, ok := reqs[podReqName]; !ok {
				reqs[podReqName] = podReqValue.DeepCopy()
			} else {
				value.Add(podReqValue)
				reqs[podReqName] = value
			}
		}
		for podLimitName, podLimitValue := range podLimits {
			if value, ok := limits[podLimitName]; !ok {
				limits[podLimitName] = podLimitValue.DeepCopy()
			} else {
				value.Add(podLimitValue)
				limits[podLimitName] = value
			}
		}
	}
	return
}
