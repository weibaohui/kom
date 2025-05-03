package kom

import (
	"fmt"
	"time"

	"github.com/weibaohui/kom/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/klog/v2"
)

// ContainerUsage 表示容器资源使用情况
type ContainerUsage struct {
	CPU        string `json:"cpu"`
	Memory     string `json:"memory"`
	CPUNano    int64  `json:"cpu_nano"`
	MemoryByte int64  `json:"memory_byte"`
}

// PodMetrics 表示容器指标数据
type PodMetrics struct {
	Name  string         `json:"name"`
	Usage ContainerUsage `json:"usage"`
}

// Top 获取容器资源使用情况 等同于  kubectl top pods ,获取列表
func (p *pod) Top() ([]*PodMetrics, error) {
	var inst []*unstructured.Unstructured
	var singlePod *unstructured.Unstructured
	stmt := p.kubectl.Statement
	cacheTime := stmt.CacheTTL
	if cacheTime == 0 {
		cacheTime = 5 * time.Second
	}
	var err error
	kubectl := p.kubectl.newInstance().WithCache(cacheTime).
		WithContext(stmt.Context).
		CRD("metrics.k8s.io", "v1beta1", "PodMetrics")
	if stmt.AllNamespace {
		kubectl = kubectl.AllNamespace()
	} else {
		kubectl = kubectl.Namespace(stmt.Namespace)
	}
	if stmt.Name != "" {
		err = kubectl.Name(stmt.Name).Get(&singlePod).Error
		if singlePod != nil {
			inst = append(inst, singlePod)
		}
	} else {
		err = kubectl.List(&inst).Error
	}

	if err != nil {
		klog.V(6).Infof("Get Top Pods  in ns %s  error %v\n", stmt.Namespace, err.Error())
		// 可能Metrics-Server 没有安装
		return nil, err
	}

	var result []*PodMetrics

	for _, item := range inst {
		if pm, err := SummarizePodMetrics(item); err == nil {
			result = append(result, pm)
		}
	}

	return result, nil
}
func (p *pod) Metrics() ([]*PodMetrics, error) {

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
func ExtractPodMetrics(u *unstructured.Unstructured, containerName string) ([]*PodMetrics, error) {
	containersRaw, found, err := unstructured.NestedSlice(u.Object, "containers")
	if err != nil {
		return nil, fmt.Errorf("failed to extract containers: %v", err)
	}
	if !found {
		return nil, fmt.Errorf("containers not found in object")
	}

	var result []*PodMetrics
	memTotal := resource.NewQuantity(0, resource.BinarySI)
	cpuTotal := resource.NewQuantity(0, resource.BinarySI)

	for _, c := range containersRaw {
		containerMap, ok := c.(map[string]interface{})
		if !ok {
			continue
		}
		if containerName != "" && containerMap["name"] != containerName {
			continue
		}

		containerMetric := &PodMetrics{
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
			cpuTotal.Add(cpuQty)
		}

		if memStr, ok := usage["memory"].(string); ok {
			memQty := resource.MustParse(memStr)
			memTotal.Add(memQty)
		}
	}

	result = append(result, &PodMetrics{
		Name: "total",
		Usage: ContainerUsage{
			CPU:        utils.FormatResource(*cpuTotal, corev1.ResourceCPU),
			CPUNano:    cpuTotal.MilliValue(),
			Memory:     utils.FormatResource(*memTotal, corev1.ResourceMemory),
			MemoryByte: memTotal.Value(),
		},
	})

	return result, nil
}

// SummarizePodMetrics 汇总Pod下的container的资源用量，返回标准结构
func SummarizePodMetrics(u *unstructured.Unstructured) (*PodMetrics, error) {
	containersRaw, found, err := unstructured.NestedSlice(u.Object, "containers")
	if err != nil {
		return nil, fmt.Errorf("failed to extract containers: %v", err)
	}
	if !found {
		return nil, fmt.Errorf("containers not found in object")
	}

	var result *PodMetrics
	memTotal := resource.NewQuantity(0, resource.BinarySI)
	cpuTotal := resource.NewQuantity(0, resource.BinarySI)

	for _, c := range containersRaw {
		containerMap, ok := c.(map[string]interface{})
		if !ok {
			continue
		}

		usage, ok := containerMap["usage"].(map[string]interface{})
		if !ok {
			continue
		}

		if cpuStr, ok := usage["cpu"].(string); ok {
			cpuQty := resource.MustParse(cpuStr)
			cpuTotal.Add(cpuQty)
		}

		if memStr, ok := usage["memory"].(string); ok {
			memQty := resource.MustParse(memStr)
			memTotal.Add(memQty)
		}
	}

	result = &PodMetrics{
		Name: u.GetName(),
		Usage: ContainerUsage{
			CPU:        utils.FormatResource(*cpuTotal, corev1.ResourceCPU),
			CPUNano:    cpuTotal.MilliValue(),
			Memory:     utils.FormatResource(*memTotal, corev1.ResourceMemory),
			MemoryByte: memTotal.Value(),
		},
	}

	return result, nil
}
