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

// ResourceUsage 获取Pod的资源使用情况，包括请求、限制、实时指标及占比
// 参数说明（可选）:
//   - denom ...UsageDenominator：用于指定“实时占比”的分母，支持：
//     DenominatorLimit：以 Pod 的 limit 值作为分母
//     DenominatorNode：以调度到的 Node 的可分配值作为分母
//     DenominatorAuto：自动选择分母（优先使用 limit，若未设置则使用 Node 可分配值）
//
// 计算规则：
//   - RequestFraction = request  / 节点可分配值
//   - LimitFraction   = limit    / 节点可分配值
//   - RealtimeFraction= realtime / 分母（按选项：limit 或节点；自动优先 limit）
//
// 说明：
//   - 默认分母为 DenominatorAuto。
//   - 为避免除零，当显式指定 DenominatorLimit 而资源未设置 limit 时，将回退到 Node 可分配值并打印日志。
//
// 来自 Kubernetes 官方文档：
// 1.CPU limit 是可选的
// 2.Memory limit 建议设置且要足够大
// 3.不要将 CPU request = limit
// 4.scheduler 只看 request，不看 limit
// CPU：必须 request，limit 可不设。
// Memory：必须 request，limit 需要设且要足够宽松。实际值建议基于历史 p95/p99 自动化生成。
func (p *pod) ResourceUsage(denom ...UsageDenominator) (*ResourceUsageResult, error) {

	var inst *v1.Pod
	cacheTime := p.kubectl.Statement.CacheTTL
	if cacheTime == 0 {
		cacheTime = 5 * time.Second
	}
	klog.V(6).Infof("Pod 资源用量缓存时间：%v\n", cacheTime)

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
		klog.V(6).Infof("获取 Pod/%s 资源用量失败：%v\n", p.kubectl.Statement.Name, err.Error())
		return nil, err
	}

	req, limit := resourcehelper.PodRequestsAndLimits(inst)
	if req == nil || limit == nil {
		return nil, fmt.Errorf("failed to get pod requests and limits")
	}

	nodeName := inst.Spec.NodeName
	if nodeName == "" {
		klog.V(6).Infof("获取 Pod/%s 资源用量失败：节点名称为空\n", p.kubectl.Statement.Name)
		return nil, fmt.Errorf("nodeName is empty")
	}
	var n *v1.Node
	err = p.kubectl.newInstance().WithContext(p.kubectl.Statement.Context).Resource(&v1.Node{}).
		WithCache(cacheTime).
		Name(nodeName).Get(&n).Error
	if err != nil {
		klog.V(6).Infof("获取 Node/%s 资源用量失败：%v\n", nodeName, err.Error())
		return nil, err
	}

	allocatable := n.Status.Capacity
	if len(n.Status.Allocatable) > 0 {
		allocatable = n.Status.Allocatable
	}

	klog.V(8).Infof("节点可分配资源信息：\n%s", utils.ToJSON(allocatable))

	cpuReq, cpuLimit, memoryReq, memoryLimit := req[v1.ResourceCPU], limit[v1.ResourceCPU], req[v1.ResourceMemory], limit[v1.ResourceMemory]
	cpuRealtime, memoryRealtime := realtimeMetrics[v1.ResourceCPU], realtimeMetrics[v1.ResourceMemory]

	fractionCpuReq := utils.FormatPercent(float64(cpuReq.MilliValue()) / float64(allocatable.Cpu().MilliValue()) * 100)
	fractionCpuLimit := utils.FormatPercent(float64(cpuLimit.MilliValue()) / float64(allocatable.Cpu().MilliValue()) * 100)
	fractionMemoryReq := utils.FormatPercent(float64(memoryReq.Value()) / float64(allocatable.Memory().Value()) * 100)
	fractionMemoryLimit := utils.FormatPercent(float64(memoryLimit.Value()) / float64(allocatable.Memory().Value()) * 100)

	// 选择分母（仅影响实时占比）
	selected := DenominatorAuto
	if len(denom) > 0 {
		selected = denom[0]
	}
	// 根据分母策略构造各资源的分母值
	chooseLimitOrNode := func(limitQ resource.Quantity, nodeQ resource.Quantity, resName string) resource.Quantity {
		if limitQ.IsZero() {
			klog.V(6).Infof("资源 %s 未设置 limit，分母回退为节点可分配值", resName)
			return nodeQ
		}
		return limitQ
	}
	var denomCPU, denomMemory resource.Quantity
	switch selected {
	case DenominatorLimit:
		denomCPU = chooseLimitOrNode(cpuLimit, *allocatable.Cpu(), "cpu")
		denomMemory = chooseLimitOrNode(memoryLimit, *allocatable.Memory(), "memory")
		klog.V(6).Infof("资源占比分母选择：使用 Limit")
	case DenominatorNode:
		denomCPU = *allocatable.Cpu()
		denomMemory = *allocatable.Memory()
		klog.V(6).Infof("资源占比分母选择：使用节点可分配值")
	case DenominatorAuto:
		if !cpuLimit.IsZero() {
			denomCPU = cpuLimit
		} else {
			denomCPU = *allocatable.Cpu()
		}
		if !memoryLimit.IsZero() {
			denomMemory = memoryLimit
		} else {
			denomMemory = *allocatable.Memory()
		}
		klog.V(6).Infof("资源占比分母选择：自动（优先 Limit，否则节点）")
	}

	// 实时占比按选定分母计算，分母为0时防御性置为0
	var fractionCpuRealtime, fractionMemoryRealtime string
	denomCpuMilli := (&denomCPU).MilliValue()
	if denomCpuMilli == 0 {
		klog.V(6).Infof("CPU 实时占比分母为0，已置为0")
		fractionCpuRealtime = utils.FormatPercent(0)
	} else {
		fractionCpuRealtime = utils.FormatPercent(float64(cpuRealtime.MilliValue()) / float64(denomCpuMilli) * 100)
	}

	denomMemVal := (&denomMemory).Value()
	if denomMemVal == 0 {
		klog.V(6).Infof("内存实时占比分母为0，已置为0")
		fractionMemoryRealtime = utils.FormatPercent(0)
	} else {
		fractionMemoryRealtime = utils.FormatPercent(float64(memoryRealtime.Value()) / float64(denomMemVal) * 100)
	}

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
		Requests: req,
		Limits:   limit,
		Realtime: realtimeMetrics,
		// 表格中的 Total 显示节点可分配值
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
