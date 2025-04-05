package kom

import (
	"time"

	"github.com/weibaohui/kom/utils"
	v1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
	resourcehelper "k8s.io/kubectl/pkg/util/resource"
)

// ResourceUsage 获取节点的资源使用情况，包括资源的请求和限制，还有当前使用占比
func (p *pod) ResourceUsage() *ResourceUsageResult {

	var inst *v1.Pod
	cacheTime := p.kubectl.Statement.CacheTTL
	if cacheTime == 0 {
		cacheTime = 5 * time.Second
	}
	klog.V(6).Infof("Pod ResourceUsage cacheTime %v\n", cacheTime)
	err := p.kubectl.newInstance().WithContext(p.kubectl.Statement.Context).Resource(&v1.Pod{}).
		Namespace(p.kubectl.Statement.Namespace).
		Name(p.kubectl.Statement.Name).
		WithCache(cacheTime).
		Get(&inst).Error
	if err != nil {
		klog.V(6).Infof("Get ResourceUsage in pod/%s  error %v\n", p.kubectl.Statement.Name, err.Error())
		return nil
	}

	nodeName := inst.Spec.NodeName
	if nodeName == "" {
		klog.V(6).Infof("Get Pod ResourceUsage in pod/%s  error %v\n", p.kubectl.Statement.Name, "nodeName is empty")
		return nil
	}
	var n *v1.Node
	err = p.kubectl.newInstance().WithContext(p.kubectl.Statement.Context).Resource(&v1.Node{}).
		WithCache(cacheTime).
		Name(nodeName).Get(&n).Error
	if err != nil {
		klog.V(6).Infof("Get Pod ResourceUsage in node/%s  error %v\n", nodeName, err.Error())
		return nil
	}

	req, limit := resourcehelper.PodRequestsAndLimits(inst)
	if req == nil || limit == nil {
		return nil
	}
	allocatable := n.Status.Capacity
	if len(n.Status.Allocatable) > 0 {
		allocatable = n.Status.Allocatable
	}

	klog.V(8).Infof("allocatable=:\n%s", utils.ToJSON(allocatable))
	cpuReq, cpuLimit, memoryReq, memoryLimit := req[v1.ResourceCPU], limit[v1.ResourceCPU], req[v1.ResourceMemory], limit[v1.ResourceMemory]
	fractionCpuReq := float64(cpuReq.MilliValue()) / float64(allocatable.Cpu().MilliValue()) * 100
	fractionCpuLimit := float64(cpuLimit.MilliValue()) / float64(allocatable.Cpu().MilliValue()) * 100
	fractionMemoryReq := float64(memoryReq.Value()) / float64(allocatable.Memory().Value()) * 100
	fractionMemoryLimit := float64(memoryLimit.Value()) / float64(allocatable.Memory().Value()) * 100

	usageFractions := map[v1.ResourceName]ResourceUsageFraction{
		v1.ResourceCPU: {
			RequestFraction: fractionCpuReq,
			LimitFraction:   fractionCpuLimit,
		},
		v1.ResourceMemory: {
			RequestFraction: fractionMemoryReq,
			LimitFraction:   fractionMemoryLimit,
		},
	}

	klog.V(6).Infof("%s\t%s\t\t%s (%d%%)\t%s (%d%%)\t%s (%d%%)\t%s (%d%%)\n", inst.Namespace, inst.Name,
		cpuReq.String(), int64(fractionCpuReq), cpuLimit.String(), int64(fractionCpuLimit),
		memoryReq.String(), int64(fractionMemoryReq), memoryLimit.String(), int64(fractionMemoryLimit))
	return &ResourceUsageResult{
		Requests:       req,
		Limits:         limit,
		Allocatable:    allocatable,
		UsageFractions: usageFractions,
	}
}
func (p *pod) ResourceUsageTable() []*ResourceUsageRow {
	usage := p.ResourceUsage()
	data, err := convertToTableData(usage)
	if err != nil {
		klog.V(6).Infof("convertToTableData error %v\n", err.Error())
		return make([]*ResourceUsageRow, 0)
	}
	return data
}
