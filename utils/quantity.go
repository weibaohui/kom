package utils

import (
	"fmt"

	"k8s.io/apimachinery/pkg/api/resource"
)

// FormatResource 格式化 resource.Quantity 为人类可读的格式
// Example:
// 示例：内存格式化
// q1 := resource.MustParse("8127096Ki")
// fmt.Println("Formatted memory:", utils.FormatResource(q1))
//
// // 示例：内存格式化（大于 Gi）
// q2 := resource.MustParse("256Gi")
// fmt.Println("Formatted memory:", utils.FormatResource(q2))
//
// // 示例：CPU 格式化
// q3 := resource.MustParse("500m") // CPU 百分之一核
// fmt.Println("Formatted CPU:", q3.String()) // CPU 不需要转换，直接原格式即可
func FormatResource(q resource.Quantity) string {
	value := q.Value()
	format := q.Format

	switch format {
	case resource.BinarySI: // Ki, Mi, Gi, etc.
		return formatBinarySI(value)
	case resource.DecimalSI: // K, M, G, etc.
		return formatDecimalSI(value)
	default:
		return q.String() // 返回原始格式
	}
}

// formatBinarySI 将二进制格式转换为易读格式 (Ki, Mi, Gi)
func formatBinarySI(value int64) string {
	const (
		Ki = 1024
		Mi = Ki * 1024
		Gi = Mi * 1024
		Ti = Gi * 1024
	)
	switch {
	case value >= Ti:
		return fmt.Sprintf("%.2fTi", float64(value)/float64(Ti))
	case value >= Gi:
		return fmt.Sprintf("%.2fGi", float64(value)/float64(Gi))
	case value >= Mi:
		return fmt.Sprintf("%.2fMi", float64(value)/float64(Mi))
	case value >= Ki:
		return fmt.Sprintf("%.2fKi", float64(value)/float64(Ki))
	default:
		return fmt.Sprintf("%d", value)
	}
}

// formatDecimalSI 将十进制格式转换为易读格式 (K, M, G)
func formatDecimalSI(value int64) string {
	const (
		K = 1000
		M = K * 1000
		G = M * 1000
		T = G * 1000
	)
	switch {
	case value >= T:
		return fmt.Sprintf("%.2fT", float64(value)/float64(T))
	case value >= G:
		return fmt.Sprintf("%.2fG", float64(value)/float64(G))
	case value >= M:
		return fmt.Sprintf("%.2fM", float64(value)/float64(M))
	case value >= K:
		return fmt.Sprintf("%.2fK", float64(value)/float64(K))
	default:
		return fmt.Sprintf("%d", value)
	}
}
