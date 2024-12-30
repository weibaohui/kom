package kom

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

// ResourceUsageFraction 定义单种资源的使用占比
type ResourceUsageFraction struct {
	RequestFraction float64 `json:"requestFraction"` // 请求使用占比（百分比）占总可分配值的比例
	LimitFraction   float64 `json:"limitFraction"`   // 限制使用占比（百分比）占总可分配值的比例
}

// ResourceUsageResult 定义资源使用情况的结构体
// 存放Pod、Node 的资源使用情况
type ResourceUsageResult struct {
	Requests       map[corev1.ResourceName]resource.Quantity     `json:"requests"`       // 请求用量
	Limits         map[corev1.ResourceName]resource.Quantity     `json:"limits"`         // 限制用量
	Allocatable    map[corev1.ResourceName]resource.Quantity     `json:"allocatable"`    // 节点可分配的实时值
	UsageFractions map[corev1.ResourceName]ResourceUsageFraction `json:"usageFractions"` // 使用占比
}
