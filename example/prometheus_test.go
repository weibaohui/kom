package example

import (
	"context"
	"testing"
	"time"

	"github.com/weibaohui/kom/kom"
)

// TestPrometheusDefaultClient 测试默认 Prometheus 客户端的自动发现功能
func TestPrometheusDefaultClient(t *testing.T) {
	ctx := context.Background()

	// 使用默认客户端，应该自动发现集群中的 Prometheus 实例
	res, err := kom.DefaultCluster().
		WithContext(ctx).
		Prometheus().
		DefaultClient().
		Expr(`up`).
		WithTimeout(5 * time.Second).
		Query()

	if err != nil {
		t.Logf("查询失败（可能集群中未安装 Prometheus）: %v", err)
		return
	}

	// 打印结果
	t.Logf("查询结果: %s", res.AsString())

	// 尝试获取向量结果
	samples := res.AsVector()
	t.Logf("向量样本数: %d", len(samples))
	for i, sample := range samples {
		if i < 5 { // 只打印前5个
			t.Logf("  样本 %d - Metric: %v, Value: %f", i+1, sample.Metric, sample.Value)
		}
	}
}

// TestPrometheusNamedClient 测试命名的 Prometheus 客户端
func TestPrometheusNamedClient(t *testing.T) {
	ctx := context.Background()

	// 使用命名客户端（例如 "prometheus-k8s"）
	res, err := kom.DefaultCluster().
		WithContext(ctx).
		Prometheus().
		Client("monitoring", "prometheus").
		Expr(`up`).
		WithTimeout(5 * time.Second).
		Query()

	if err != nil {
		t.Logf("查询失败（可能集群中未安装指定的 Prometheus）: %v", err)
		return
	}

	t.Logf("查询结果: %s", res.AsString())
}

// TestPrometheusWithAddress 测试使用显式地址的 Prometheus 客户端
func TestPrometheusWithAddress(t *testing.T) {
	ctx := context.Background()

	// 使用显式地址（如果有外部 Prometheus 实例）
	// 注意：这里的地址需要根据实际情况修改
	res, err := kom.DefaultCluster().
		WithContext(ctx).
		Prometheus().
		WithAddress("http://prometheus.monitoring.svc:9090").
		Expr(`up`).
		WithTimeout(5 * time.Second).
		Query()

	if err != nil {
		t.Logf("查询失败: %v", err)
		return
	}

	t.Logf("查询结果: %s", res.AsString())
}

// TestPrometheusQueryScalar 测试标量查询
func TestPrometheusQueryScalar(t *testing.T) {
	ctx := context.Background()

	// 查询当前在线的 target 数量
	value, err := kom.DefaultCluster().
		WithContext(ctx).
		Prometheus().
		DefaultClient().
		Expr(`count(up == 1)`).
		QueryScalar()

	if err != nil {
		t.Logf("查询失败: %v", err)
		return
	}

	t.Logf("在线 target 数量: %f", value)
}

// TestPrometheusQueryRange 测试范围查询
func TestPrometheusQueryRange(t *testing.T) {
	ctx := context.Background()

	// 查询过去5分钟的数据
	start := time.Now().Add(-5 * time.Minute)
	end := time.Now()

	res, err := kom.DefaultCluster().
		WithContext(ctx).
		Prometheus().
		DefaultClient().
		Expr(`up`).
		QueryRange(start, end, 30*time.Second)

	if err != nil {
		t.Logf("范围查询失败: %v", err)
		return
	}

	series := res.AsMatrix()
	t.Logf("时间序列数: %d", len(series))
	for i, s := range series {
		if i < 3 { // 只打印前3个
			t.Logf("  序列 %d - Metric: %v, 样本点数: %d", i+1, s.Metric, len(s.Samples))
		}
	}
}

// TestPrometheusForPod 测试针对 Pod 的查询过滤
func TestPrometheusForPod(t *testing.T) {
	ctx := context.Background()

	// 查询特定 Pod 的指标
	res, err := kom.DefaultCluster().
		WithContext(ctx).
		Prometheus().
		DefaultClient().
		Expr(`up`).
		ForPod("kube-system", "kube-apiserver-*").
		Query()

	if err != nil {
		t.Logf("查询失败: %v", err)
		return
	}

	samples := res.AsVector()
	t.Logf("匹配的样本数: %d", len(samples))
	for _, sample := range samples {
		t.Logf("  Metric: %v, Value: %f", sample.Metric, sample.Value)
	}
}

// TestPrometheusQuerySum 测试聚合查询
func TestPrometheusQuerySum(t *testing.T) {
	ctx := context.Background()

	// 查询所有在线 target 的总和
	sum, err := kom.DefaultCluster().
		WithContext(ctx).
		Prometheus().
		DefaultClient().
		Expr(`up == 1`).
		QuerySum()

	if err != nil {
		t.Logf("聚合查询失败: %v", err)
		return
	}

	t.Logf("在线 target 总数: %f", sum)
}

// TestPrometheusQuerySumBy 测试分组聚合查询
func TestPrometheusQuerySumBy(t *testing.T) {
	ctx := context.Background()

	// 按 job 分组统计在线 target
	result, err := kom.DefaultCluster().
		WithContext(ctx).
		Prometheus().
		DefaultClient().
		Expr(`up == 1`).
		QuerySumBy("job")

	if err != nil {
		t.Logf("分组聚合查询失败: %v", err)
		return
	}

	t.Logf("按 job 分组的结果:")
	for job, count := range result {
		t.Logf("  %s: %f", job, count)
	}
}
