package kom

import (
	"fmt"
	"math/big"
	"time"

	"github.com/weibaohui/kom/utils"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/klog/v2"
	resourcehelper "k8s.io/kubectl/pkg/util/resource"
)

// ContainerUsage 表示容器资源使用情况
type ContainerUsage struct {
	CPU    string `json:"cpu"`
	Memory string `json:"memory"`
}

// PodMetrics 表示容器指标数据
type PodMetrics struct {
	Name  string         `json:"name"`
	Usage ContainerUsage `json:"usage"`
}

func (p *pod) Metrics() ([]PodMetrics, error) {

	var inst unstructured.Unstructured
	stmt := p.kubectl.Statement
	cacheTime := stmt.CacheTTL
	containerName := stmt.ContainerName
	if cacheTime == 0 {
		cacheTime = 5 * time.Second
	}
	err := p.kubectl.newInstance().
		WithContext(stmt.Context).
		CRD("metrics.k8s.io", "v1beta1", "PodMetrics").
		Namespace(stmt.Namespace).
		Name(stmt.Name).
		WithCache(cacheTime).
		Get(&inst).Error
	if err != nil {
		klog.V(6).Infof("Get ResourceUsage in pod/%s  error %v\n", stmt.Name, err.Error())
		// 可能Metrics-Server 没有安装
		return nil, err
	}

	containers, err := ExtractPodMetrics(&inst, containerName)
	if err != nil {
		return nil, err
	}

	return containers, nil
}

// ExtractPodMetrics 提取 containers 字段，返回标准结构
func ExtractPodMetrics(u *unstructured.Unstructured, containerName string) ([]PodMetrics, error) {
	containersRaw, found, err := unstructured.NestedSlice(u.Object, "containers")
	if err != nil {
		return nil, fmt.Errorf("failed to extract containers: %v", err)
	}
	if !found {
		return nil, fmt.Errorf("containers not found in object")
	}

	var result []PodMetrics
	cpuTotal := big.NewInt(0)
	memTotal := big.NewInt(0)

	for _, c := range containersRaw {
		containerMap, ok := c.(map[string]interface{})
		if !ok {
			continue
		}
		if containerName != "" && containerMap["name"] != containerName {
			continue
		}

		containerMetric := PodMetrics{
			Name:  containerMap["name"].(string),
			Usage: ContainerUsage{},
		}

		if usage, ok := containerMap["usage"].(map[string]interface{}); ok {
			if cpuStr, ok := usage["cpu"].(string); ok {
				containerMetric.Usage.CPU = cpuStr
			}
			if memStr, ok := usage["memory"].(string); ok {
				containerMetric.Usage.Memory = memStr
			}
		}

		result = append(result, containerMetric)

		usage, ok := containerMap["usage"].(map[string]interface{})
		if !ok {
			continue
		}

		if cpuStr, ok := usage["cpu"].(string); ok {
			cpuQty := resource.MustParse(cpuStr)
			cpuTotal.Add(cpuTotal, cpuQty.AsDec().UnscaledBig())
		}

		if memStr, ok := usage["memory"].(string); ok {
			memQty := resource.MustParse(memStr)
			memTotal.Add(memTotal, memQty.AsDec().UnscaledBig())
		}
	}

	// // 计算成 float64 形式（CPU: m核；内存: Mi）
	// cpuMilli := new(big.Float).Quo(new(big.Float).SetInt(cpuTotal), big.NewFloat(1_000_000))
	// memMi := new(big.Float).Quo(new(big.Float).SetInt(memTotal), big.NewFloat(1024*1024))
	//
	// // 格式化字符串形式
	// cpuFormatted, _ := cpuMilli.Float64()
	// memFormatted, _ := memMi.Float64()

	result = append(result, PodMetrics{
		Name: "total",
		Usage: ContainerUsage{
			CPU:    fmt.Sprintf("%.0fn", cpuTotal),  // 毫核
			Memory: fmt.Sprintf("%.0fKi", memTotal), // MiB
		},
	})

	return result, nil
}

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
	fractionCpuReq := float64(cpuReq.MilliValue()) / float64(allocatable.Cpu().MilliValue()) * 100
	fractionCpuLimit := float64(cpuLimit.MilliValue()) / float64(allocatable.Cpu().MilliValue()) * 100
	fractionCpuRealtime := float64(cpuRealtime.MilliValue()) / float64(allocatable.Cpu().MilliValue()) * 100

	fractionMemoryReq := float64(memoryReq.Value()) / float64(allocatable.Memory().Value()) * 100
	fractionMemoryLimit := float64(memoryLimit.Value()) / float64(allocatable.Memory().Value()) * 100
	fractionMemoryRealtime := float64(memoryRealtime.Value()) / float64(allocatable.Memory().Value()) * 100

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

	klog.V(6).Infof("%s\t%s\t\t%s (%d%%)\t%s (%d%%)\t%s (%d%%)\t%s (%d%%)\n", inst.Namespace, inst.Name,
		cpuReq.String(), int64(fractionCpuReq), cpuLimit.String(), int64(fractionCpuLimit),
		memoryReq.String(), int64(fractionMemoryReq), memoryLimit.String(), int64(fractionMemoryLimit))

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
