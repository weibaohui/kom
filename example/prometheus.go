package example

import (
	"context"
	"time"

	"github.com/weibaohui/kom/kom"
	"k8s.io/klog/v2"
)

func PrometheusDefaultClient() {
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
		klog.V(6).Infof("查询失败（可能集群中未安装 Prometheus）: %v", err)
		return
	}

	// 打印结果
	klog.V(6).Infof("查询结果: %s", res.AsString())

	// 尝试获取向量结果
	samples := res.AsVector()
	klog.V(6).Infof("向量样本数: %d", len(samples))
	for i, sample := range samples {
		if i < 5 { // 只打印前5个
			klog.V(6).Infof("  样本 %d - Metric: %v, Value: %f", i+1, sample.Metric, sample.Value)
		}
	}
}

func PrometheusNamedClient() {
	ctx := context.Background()

	// 使用命名客户端（例如 "prometheus-k8s"）
	res, err := kom.DefaultCluster().
		WithContext(ctx).
		Prometheus().
		WithAddress("http://127.0.0.1:45224/").
		// Client("monitoring", "prometheus").
		Expr(`up`).
		WithTimeout(5 * time.Second).
		Query()

	if err != nil {
		klog.V(6).Infof("查询失败（可能集群中未安装指定的 Prometheus）: %v", err)
		return
	}

	klog.V(6).Infof("查询结果: %s", res.AsString())
}

func PrometheusWithAddress() {
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
		klog.V(6).Infof("查询失败: %v", err)
		return
	}

	klog.V(6).Infof("查询结果: %s", res.AsString())
}

func PrometheusQueryScalar() {
	ctx := context.Background()

	// 查询当前在线的 target 数量
	value, err := kom.DefaultCluster().
		WithContext(ctx).
		Prometheus().
		DefaultClient().
		Expr(`count(up == 1)`).
		QueryScalar()

	if err != nil {
		klog.V(6).Infof("查询失败: %v", err)
		return
	}

	klog.V(6).Infof("在线 target 数量: %f", value)
}

func PrometheusQueryRange() {
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
		klog.V(6).Infof("范围查询失败: %v", err)
		return
	}

	series := res.AsMatrix()
	klog.V(6).Infof("时间序列数: %d", len(series))
	for i, s := range series {
		if i < 3 { // 只打印前3个
			klog.V(6).Infof("  序列 %d - Metric: %v, 样本点数: %d", i+1, s.Metric, len(s.Samples))
		}
	}
}

func PrometheusForPod() {
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
		klog.V(6).Infof("查询失败: %v", err)
		return
	}

	samples := res.AsVector()
	klog.V(6).Infof("匹配的样本数: %d", len(samples))
	for _, sample := range samples {
		klog.V(6).Infof("  Metric: %v, Value: %f", sample.Metric, sample.Value)
	}
}

func PrometheusQuerySum() {
	ctx := context.Background()

	// 查询所有在线 target 的总和
	sum, err := kom.DefaultCluster().
		WithContext(ctx).
		Prometheus().
		DefaultClient().
		Expr(`up == 1`).
		QuerySum()

	if err != nil {
		klog.V(6).Infof("聚合查询失败: %v", err)
		return
	}

	klog.V(6).Infof("在线 target 总数: %f", sum)
}

func PrometheusQuerySumBy() {
	ctx := context.Background()

	// 按 job 分组统计在线 target
	result, err := kom.DefaultCluster().
		WithContext(ctx).
		Prometheus().
		DefaultClient().
		Expr(`up == 1`).
		QuerySumBy("job")

	if err != nil {
		klog.V(6).Infof("分组聚合查询失败: %v", err)
		return
	}

	klog.V(6).Infof("按 job 分组的结果:")
	for job, count := range result {
		klog.V(6).Infof("  %s: %f", job, count)
	}
}
