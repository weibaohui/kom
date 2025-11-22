package kom

import (
    "fmt"

    "github.com/weibaohui/kom/utils"
    corev1 "k8s.io/api/core/v1"
    "k8s.io/apimachinery/pkg/api/resource"
)

// UsageDenominator 表示资源占比计算的分母选项
// DenominatorAuto：自动选择（优先使用 limit，否则使用节点可分配值）
// DenominatorNode：使用节点可分配值
// DenominatorLimit：使用 Pod 的 limit 值
type UsageDenominator int

const (
    DenominatorAuto UsageDenominator = iota
    DenominatorNode
    DenominatorLimit
)

func (d UsageDenominator) String() string {
    switch d {
    case DenominatorAuto:
        return "auto"
    case DenominatorNode:
        return "node"
    case DenominatorLimit:
        return "limit"
    default:
        return "unknown"
    }
}

// ResourceUsageFraction 定义单种资源的使用占比（相对于分母选项）
type ResourceUsageFraction struct {
    RequestFraction  string `json:"requestFraction"`  // 请求使用占比（百分比）
    LimitFraction    string `json:"limitFraction"`    // 限制使用占比（百分比）
    RealtimeFraction string `json:"realtimeFraction"` // 实时指标显示的占比（百分比）
}

// ResourceUsageResult 定义资源使用情况的结构体
// 存放Pod、Node 的资源使用情况
type ResourceUsageResult struct {
	Requests       map[corev1.ResourceName]resource.Quantity     `json:"requests"` // 请求用量
	Limits         map[corev1.ResourceName]resource.Quantity     `json:"limits"`   // 限制用量
	Realtime       map[corev1.ResourceName]resource.Quantity     `json:"realtime"`
    Allocatable    map[corev1.ResourceName]resource.Quantity     `json:"allocatable"`    // 节点可分配的实时值（用于表格 Total）
	UsageFractions map[corev1.ResourceName]ResourceUsageFraction `json:"usageFractions"` // 使用占比
}

// ResourceUsageRow 临时结构体，用于存储每一行数据
type ResourceUsageRow struct {
	ResourceType     string `json:"resourceType"`
	Total            string `json:"total"`
	Request          string `json:"request"`
	RequestFraction  string `json:"requestFraction"`
	Limit            string `json:"limit"`
	LimitFraction    string `json:"limitFraction"`
	Realtime         string `json:"realtime"`
	RealtimeFraction string `json:"realtimeFraction"`
}

func convertToTableData(result *ResourceUsageResult) ([]*ResourceUsageRow, error) {
	if result == nil {
		return nil, fmt.Errorf("result is nil")
	}
	var tableData []*ResourceUsageRow

	// 遍历资源类型（CPU、内存等），并生成表格行
	for _, resourceType := range []corev1.ResourceName{corev1.ResourceCPU, corev1.ResourceMemory, corev1.ResourceEphemeralStorage} {
		// 创建一行数据
		alc := result.Allocatable[resourceType]
		req := result.Requests[resourceType]
		lit := result.Limits[resourceType]
		realtime := result.Realtime[resourceType]
		row := &ResourceUsageRow{
			ResourceType: string(resourceType),
			Total:        utils.FormatResource(alc, resourceType),

			Request:  utils.FormatResource(req, resourceType),
			Limit:    utils.FormatResource(lit, resourceType),
			Realtime: utils.FormatResource(realtime, resourceType),

			RequestFraction:  result.UsageFractions[resourceType].RequestFraction,
			LimitFraction:    result.UsageFractions[resourceType].LimitFraction,
			RealtimeFraction: result.UsageFractions[resourceType].RealtimeFraction,
		}
		// 将行加入表格数据
		tableData = append(tableData, row)
	}

	return tableData, nil
}
