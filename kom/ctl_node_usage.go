package kom

import (
	"fmt"
	"time"

	"github.com/weibaohui/kom/utils"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/klog/v2"
	resourcehelper "k8s.io/kubectl/pkg/util/resource"
)

// NodeUsage 表示容器资源使用情况
type NodeUsage struct {
	CPU        string `json:"cpu"`
	Memory     string `json:"memory"`
	CPUNano    int64  `json:"cpu_nano"`
	MemoryByte int64  `json:"memory_byte"`
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
func (d *node) ResourceUsage() (*ResourceUsageResult, error) {

	// 计算实时值
	realtimeMetrics := make(map[v1.ResourceName]resource.Quantity)

	if metrics, err := d.Metrics(); err == nil {
		cpu := metrics.CPU
		if cpu != "" {
			if cpuQty, err := resource.ParseQuantity(cpu); err == nil {
				realtimeMetrics[v1.ResourceCPU] = cpuQty
			}
		}
		memory := metrics.Memory
		if memory != "" {
			if memQty, err := resource.ParseQuantity(memory); err == nil {
				realtimeMetrics[v1.ResourceMemory] = memQty
			}
		}

	}

	reqs, limits := d.TotalRequestsAndLimits()
	if reqs == nil || limits == nil {
		return nil, fmt.Errorf("getPodsTotalRequestsAndLimits error")
	}
	cacheTime := d.getCacheTTL()
	n, err := d.getNodeWithCache(cacheTime)
	if err != nil {
		klog.V(6).Infof("Get ResourceUsage in node/%s  error %v\n", d.kubectl.Statement.Name, err.Error())
		return nil, err
	}

	allocatable := n.Status.Capacity
	if len(n.Status.Allocatable) > 0 {
		allocatable = n.Status.Allocatable
	}

	klog.V(8).Infof("allocatable=:\n%s", utils.ToJSON(allocatable))
	cpuReqs, cpuLimits, memoryReqs, memoryLimits, ephemeralstorageReqs, ephemeralstorageLimits :=
		reqs[corev1.ResourceCPU], limits[corev1.ResourceCPU], reqs[corev1.ResourceMemory], limits[corev1.ResourceMemory], reqs[corev1.ResourceEphemeralStorage], limits[corev1.ResourceEphemeralStorage]
	cpuRealtime, memoryRealtime := realtimeMetrics[corev1.ResourceCPU], realtimeMetrics[corev1.ResourceMemory]

	// 计算CPU 使用率
	fractionCpuReqs := ""
	fractionCpuLimits := ""
	fractionCpuRealtime := ""
	if allocatable.Cpu().MilliValue() != 0 {
		fractionCpuReqs = utils.FormatPercent(float64(cpuReqs.MilliValue()) / float64(allocatable.Cpu().MilliValue()) * 100)
		fractionCpuLimits = utils.FormatPercent(float64(cpuLimits.MilliValue()) / float64(allocatable.Cpu().MilliValue()) * 100)
		fractionCpuRealtime = utils.FormatPercent(float64(cpuRealtime.MilliValue()) / float64(allocatable.Cpu().MilliValue()) * 100)
	}

	// 计算内存 使用率
	fractionMemoryReqs := ""
	fractionMemoryLimits := ""
	fractionMemoryRealtime := ""
	if allocatable.Memory().Value() != 0 {
		fractionMemoryReqs = utils.FormatPercent(float64(memoryReqs.Value()) / float64(allocatable.Memory().Value()) * 100)
		fractionMemoryLimits = utils.FormatPercent(float64(memoryLimits.Value()) / float64(allocatable.Memory().Value()) * 100)
		fractionMemoryRealtime = utils.FormatPercent(float64(memoryRealtime.Value()) / float64(allocatable.Memory().Value()) * 100)
	}

	// 计算存储 使用率
	fractionEphemeralStorageReqs := ""
	fractionEphemeralStorageLimits := ""
	if allocatable.StorageEphemeral().Value() != 0 {
		fractionEphemeralStorageReqs = utils.FormatPercent(float64(ephemeralstorageReqs.Value()) / float64(allocatable.StorageEphemeral().Value()) * 100)
		fractionEphemeralStorageLimits = utils.FormatPercent(float64(ephemeralstorageLimits.Value()) / float64(allocatable.StorageEphemeral().Value()) * 100)
	}

	usageFractions := map[corev1.ResourceName]ResourceUsageFraction{
		corev1.ResourceCPU: {
			RequestFraction:  fractionCpuReqs,
			LimitFraction:    fractionCpuLimits,
			RealtimeFraction: fractionCpuRealtime,
		},
		corev1.ResourceMemory: {
			RequestFraction:  fractionMemoryReqs,
			LimitFraction:    fractionMemoryLimits,
			RealtimeFraction: fractionMemoryRealtime,
		},
		corev1.ResourceEphemeralStorage: {
			RequestFraction: fractionEphemeralStorageReqs,
			LimitFraction:   fractionEphemeralStorageLimits,
		},
	}

	return &ResourceUsageResult{
		Requests:       reqs,
		Limits:         limits,
		Realtime:       realtimeMetrics,
		Allocatable:    allocatable,
		UsageFractions: usageFractions,
	}, nil
}
func (d *node) ResourceUsageTable() ([]*ResourceUsageRow, error) {
	usage, err := d.ResourceUsage()
	if err != nil {
		return nil, err
	}
	data, err := convertToTableData(usage)
	if err != nil {
		klog.V(6).Infof("convertToTableData error %v\n", err.Error())
		return nil, err
	}
	return data, nil
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

func (d *node) Metrics() (*NodeUsage, error) {

	var inst *unstructured.Unstructured
	stmt := d.kubectl.Statement
	cacheTime := stmt.CacheTTL
	if cacheTime == 0 {
		cacheTime = 5 * time.Second
	}
	err := d.kubectl.newInstance().
		WithContext(stmt.Context).
		CRD("metrics.k8s.io", "v1beta1", "NodeMetrics").
		Name(stmt.Name).
		WithCache(cacheTime).
		Get(&inst).Error
	if err != nil {
		klog.V(6).Infof("Get ResourceUsage in Node/%s  error %v\n", stmt.Name, err.Error())
		// 可能Metrics-Server 没有安装
		return nil, err
	}
	klog.V(6).Infof("inst = %s", inst)
	usage, err := ExtractNodeMetrics(inst)
	if err != nil {
		return nil, err
	}

	return usage, nil
}

// ExtractNodeMetrics 提取 containers 字段，返回标准结构
func ExtractNodeMetrics(u *unstructured.Unstructured) (*NodeUsage, error) {
	usageRaw, found, err := unstructured.NestedMap(u.Object, "usage")
	if err != nil {
		return nil, fmt.Errorf("failed to extract containers: %v", err)
	}
	if !found {
		return nil, fmt.Errorf("containers not found in object")
	}
	klog.V(6).Infof("usageraw %s", usageRaw)
	nodeUsage := &NodeUsage{
		CPU:    usageRaw[corev1.ResourceCPU.String()].(string),
		Memory: usageRaw[corev1.ResourceMemory.String()].(string),
	}
	klog.V(6).Infof("node/%s resource usage\n", utils.ToJSON(nodeUsage))
	return nodeUsage, nil
}

// NodeMetrics 表示节点指标数据
type NodeMetrics struct {
	Name  string    `json:"name"`
	Usage NodeUsage `json:"usage"`
}

// Top 获取节点资源使用情况，等同于 kubectl top nodes，返回节点列表
func (d *node) Top() ([]*NodeMetrics, error) {
	var inst []*unstructured.Unstructured
	var singleNode *unstructured.Unstructured
	stmt := d.kubectl.Statement
	cacheTime := stmt.CacheTTL
	if cacheTime == 0 {
		cacheTime = 5 * time.Second
	}
	var err error
	kubectl := d.kubectl.newInstance().WithCache(cacheTime).
		WithContext(stmt.Context).
		CRD("metrics.k8s.io", "v1beta1", "NodeMetrics")

	if stmt.Name != "" {
		err = kubectl.Name(stmt.Name).Get(&singleNode).Error
		if singleNode != nil {
			inst = append(inst, singleNode)
		}
	} else {
		err = kubectl.List(&inst).Error
	}

	if err != nil {
		klog.V(6).Infof("Get Top Nodes  error %v\n", err.Error())
		// 可能Metrics-Server 没有安装
		return nil, err
	}

	var result []*NodeMetrics

	for _, item := range inst {

		memTotal := resource.NewQuantity(0, resource.BinarySI)
		cpuTotal := resource.NewQuantity(0, resource.BinarySI)
		if usage, ok := item.Object["usage"].(map[string]interface{}); ok {
			if cpuStr, ok := usage[corev1.ResourceCPU.String()].(string); ok {
				cpuQty := resource.MustParse(cpuStr)
				cpuTotal.Add(cpuQty)
			}
			if memStr, ok := usage[corev1.ResourceMemory.String()].(string); ok {
				memQty := resource.MustParse(memStr)
				memTotal.Add(memQty)
			}
		}
		result = append(result, &NodeMetrics{
			Name: item.GetName(),
			Usage: NodeUsage{
				CPU:        utils.FormatResource(*cpuTotal, corev1.ResourceCPU),
				CPUNano:    cpuTotal.MilliValue(),
				Memory:     utils.FormatResource(*memTotal, corev1.ResourceMemory),
				MemoryByte: memTotal.Value(),
			},
		})

	}

	return result, nil
}
