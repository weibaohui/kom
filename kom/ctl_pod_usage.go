package kom

import (
	"fmt"
	"time"

	"github.com/weibaohui/kom/utils"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/klog/v2"
	resourcehelper "k8s.io/kubectl/pkg/util/resource"
)

// ResourceUsage 获取节点的资源使用情况，包括资源的请求和限制，还有当前使用占比
func (p *pod) ResourceUsage() (*ResourceUsageResult, error) {

	var inst *v1.Pod
	cacheTime := p.kubectl.Statement.CacheTTL
	if cacheTime == 0 {
		cacheTime = 5 * time.Second
	}
	klog.V(6).Infof("Pod ResourceUsage cacheTime %v\n", cacheTime)

	// 计算实时值
	realtimeMetrics := make(map[v1.ResourceName]resource.Quantity)

	if podMetrics, err := p.Metrics(); err == nil {
		for _, metric := range podMetrics {
			if metric.Name != "total" {
				continue
			}
			if cpuQty, err := resource.ParseQuantity(metric.Usage.CPU); err == nil {
				realtimeMetrics[v1.ResourceCPU] = cpuQty
			}
			if memQty, err := resource.ParseQuantity(metric.Usage.Memory); err == nil {
				realtimeMetrics[v1.ResourceMemory] = memQty
			}

		}
	}

	err := p.kubectl.newInstance().WithContext(p.kubectl.Statement.Context).Resource(&v1.Pod{}).
		Namespace(p.kubectl.Statement.Namespace).
		Name(p.kubectl.Statement.Name).
		WithCache(cacheTime).
		Get(&inst).Error
	if err != nil {
		klog.V(6).Infof("Get ResourceUsage in pod/%s  error %v\n", p.kubectl.Statement.Name, err.Error())
		return nil, err
	}

	nodeName := inst.Spec.NodeName
	if nodeName == "" {
		klog.V(6).Infof("Get Pod ResourceUsage in pod/%s  error %v\n", p.kubectl.Statement.Name, "nodeName is empty")
		return nil, fmt.Errorf("nodeName is empty")
	}
	var n *v1.Node
	err = p.kubectl.newInstance().WithContext(p.kubectl.Statement.Context).Resource(&v1.Node{}).
		WithCache(cacheTime).
		Name(nodeName).Get(&n).Error
	if err != nil {
		klog.V(6).Infof("Get Pod ResourceUsage in node/%s  error %v\n", nodeName, err.Error())
		return nil, err
	}

	req, limit := resourcehelper.PodRequestsAndLimits(inst)
	if req == nil || limit == nil {
		return nil, fmt.Errorf("failed to get pod requests and limits")
	}
	allocatable := n.Status.Capacity
	if len(n.Status.Allocatable) > 0 {
		allocatable = n.Status.Allocatable
	}

	klog.V(8).Infof("allocatable=:\n%s", utils.ToJSON(allocatable))

	cpuReq, cpuLimit, memoryReq, memoryLimit := req[v1.ResourceCPU], limit[v1.ResourceCPU], req[v1.ResourceMemory], limit[v1.ResourceMemory]
	cpuRealtime, memoryRealtime := realtimeMetrics[v1.ResourceCPU], realtimeMetrics[v1.ResourceMemory]

	fractionCpuReq := utils.FormatPercent(float64(cpuReq.MilliValue()) / float64(allocatable.Cpu().MilliValue()) * 100)
	fractionCpuLimit := utils.FormatPercent(float64(cpuLimit.MilliValue()) / float64(allocatable.Cpu().MilliValue()) * 100)
	fractionCpuRealtime := utils.FormatPercent(float64(cpuRealtime.MilliValue()) / float64(allocatable.Cpu().MilliValue()) * 100)

	fractionMemoryReq := utils.FormatPercent(float64(memoryReq.Value()) / float64(allocatable.Memory().Value()) * 100)
	fractionMemoryLimit := utils.FormatPercent(float64(memoryLimit.Value()) / float64(allocatable.Memory().Value()) * 100)
	fractionMemoryRealtime := utils.FormatPercent(float64(memoryRealtime.Value()) / float64(allocatable.Memory().Value()) * 100)

	usageFractions := map[v1.ResourceName]ResourceUsageFraction{
		v1.ResourceCPU: {
			RequestFraction:  fractionCpuReq,
			LimitFraction:    fractionCpuLimit,
			RealtimeFraction: fractionCpuRealtime,
		},
		v1.ResourceMemory: {
			RequestFraction:  fractionMemoryReq,
			LimitFraction:    fractionMemoryLimit,
			RealtimeFraction: fractionMemoryRealtime,
		},
	}

	return &ResourceUsageResult{
		Requests:       req,
		Limits:         limit,
		Realtime:       realtimeMetrics,
		Allocatable:    allocatable,
		UsageFractions: usageFractions,
	}, nil
}
func (p *pod) ResourceUsageTable() ([]*ResourceUsageRow, error) {
	usage, err := p.ResourceUsage()
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
