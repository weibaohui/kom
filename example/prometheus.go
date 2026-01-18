package example

import (
	"context"
	"time"

	"github.com/weibaohui/kom/kom"
	"k8s.io/klog/v2"
)

func PrometheusNamedClient() {
	ctx := context.Background()

	// 使用命名客户端（例如 "prometheus-k8s"）
	res, err := kom.DefaultCluster().
		WithContext(ctx).
		Prometheus().
		WithAddress("http://127.0.0.1:45972/").
		Expr(`up`).
		WithTimeout(5 * time.Second).
		Query()

	if err != nil {
		klog.V(6).Infof("查询失败（可能集群中未安装指定的 Prometheus）: %v", err)
		return
	}

	klog.V(6).Infof("查询结果: %s", res.AsString())
}

func PrometheusQuery() {
	ctx := context.Background()

	value, err := kom.DefaultCluster().
		WithContext(ctx).
		Prometheus().
		WithAddress("http://127.0.0.1:43422/").
		Expr(`sum by (instance) (irate(node_cpu_seconds_total{mode!="idle"}[1m])) / sum by (instance) (irate(node_cpu_seconds_total[1m])) * 100`).
		QueryRange(time.Now().Add(-1*time.Hour), time.Now(), time.Minute)

	if err != nil {
		klog.V(6).Infof("查询失败: %v", err)
		return
	}

	klog.V(6).Infof(" value: %s", value.AsString())
}
