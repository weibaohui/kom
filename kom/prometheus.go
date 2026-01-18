package kom

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/prometheus/client_golang/api"
	promv1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
)

// PrometheusService 提供基于当前集群的 Prometheus 访问能力。
type PrometheusService struct {
	kubectl *Kubectl
}

// Prometheus 从当前 Cluster/Kubectl 构造一个 Prometheus 服务访问器。
func (k *Kubectl) Prometheus() *PrometheusService {
	return &PrometheusService{kubectl: k}
}

// PromClient 表示一个具体的 Prometheus 实例（本地 Prom / Thanos 等）。
type PromClient struct {
	service *PrometheusService
	name    string
	address string
}

// DefaultClient 返回当前集群的默认 Prometheus 客户端。
// 实际地址解析逻辑由 PrometheusService.resolveAddress 负责。
func (s *PrometheusService) DefaultClient() *PromClient {
	return &PromClient{
		service: s,
		name:    "default",
	}
}

// Client 按名称返回当前集群下的 Prometheus 客户端。
// 地址解析同样由 PrometheusService.resolveAddress 负责。
func (s *PrometheusService) Client(name string) *PromClient {
	if name == "" {
		name = "default"
	}
	return &PromClient{
		service: s,
		name:    name,
	}
}

// WithAddress 使用显式地址构造一个临时 Prometheus 客户端，不依赖集群配置。
func (s *PrometheusService) WithAddress(addr string) *PromClient {
	return &PromClient{
		service: s,
		address: addr,
	}
}

// Expr 在指定 Prometheus 客户端上构造一个查询构建器。
func (c *PromClient) Expr(expr string) *PromQuery {
	return &PromQuery{
		client:        c,
		expr:          expr,
		labelMatchers: map[string]string{},
	}
}

// PromQuery 表示一次 Prometheus 查询的构建器。
type PromQuery struct {
	client *PromClient

	expr string

	queryTime *time.Time

	start *time.Time
	end   *time.Time
	step  *time.Duration

	timeout *time.Duration

	labelMatchers map[string]string
}

// WithTimeout 设置单次查询的超时时间。
func (q *PromQuery) WithTimeout(d time.Duration) *PromQuery {
	q.timeout = &d
	return q
}

// ForPod 为当前查询追加 Pod 级别的过滤条件（基于 namespace / pod 标签）。
func (q *PromQuery) ForPod(namespace, name string) *PromQuery {
	if q.labelMatchers == nil {
		q.labelMatchers = map[string]string{}
	}
	if namespace != "" {
		q.labelMatchers["namespace"] = namespace
	}
	if name != "" {
		q.labelMatchers["pod"] = name
	}
	return q
}

// ForDeployment 为当前查询追加 Deployment 级别的过滤条件。
func (q *PromQuery) ForDeployment(namespace, name string) *PromQuery {
	if q.labelMatchers == nil {
		q.labelMatchers = map[string]string{}
	}
	if namespace != "" {
		q.labelMatchers["namespace"] = namespace
	}
	if name != "" {
		q.labelMatchers["deployment"] = name
	}
	return q
}

// Query 在当前时间点执行瞬时查询，使用链路上通过 WithContext 设置的 context。
func (q *PromQuery) Query() (*PromResult, error) {
	ctx := q.getContext()
	apiClient, err := q.client.api()
	if err != nil {
		return nil, err
	}
	ts := time.Now()
	if q.queryTime != nil {
		ts = *q.queryTime
	}
	value, warnings, err := apiClient.Query(ctx, q.expr, ts)
	if err != nil {
		return nil, err
	}
	value = filterValueByLabels(value, q.labelMatchers)
	return &PromResult{
		value:    value,
		warnings: warnings,
	}, nil
}

// QueryAt 在指定时间点执行瞬时查询。
func (q *PromQuery) QueryAt(t time.Time) (*PromResult, error) {
	q.queryTime = &t
	return q.Query()
}

// QueryRange 执行区间查询，使用链路上的 context。
func (q *PromQuery) QueryRange(start, end time.Time, step time.Duration) (*PromResult, error) {
	ctx := q.getContext()
	apiClient, err := q.client.api()
	if err != nil {
		return nil, err
	}
	r := promv1.Range{
		Start: start,
		End:   end,
		Step:  step,
	}
	value, warnings, err := apiClient.QueryRange(ctx, q.expr, r)
	if err != nil {
		return nil, err
	}
	value = filterValueByLabels(value, q.labelMatchers)
	return &PromResult{
		value:    value,
		warnings: warnings,
	}, nil
}

// getContext 从 Kubectl.Statement 中获取链路上的 context，并应用超时设置。
func (q *PromQuery) getContext() context.Context {
	var ctx context.Context
	if q.client != nil && q.client.service != nil && q.client.service.kubectl != nil {
		ctx = q.client.service.kubectl.Statement.Context
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if q.timeout != nil {
		c, _ := context.WithTimeout(ctx, *q.timeout)
		return c
	}
	return ctx
}

// api 构造 Prometheus v1 API 客户端。
func (c *PromClient) api() (promv1.API, error) {
	addr := c.address
	if addr == "" {
		addr = c.service.resolveAddress(c.name)
	}
	if addr == "" {
		return nil, fmt.Errorf("prometheus address is not configured")
	}
	cli, err := api.NewClient(api.Config{Address: addr})
	if err != nil {
		return nil, err
	}
	return promv1.NewAPI(cli), nil
}

// resolveAddress 负责根据集群信息和 client 名称解析 Prometheus 地址。
// 当前版本仅作为占位实现，返回空字符串，调用方可通过 WithAddress 显式指定。
func (s *PrometheusService) resolveAddress(name string) string {
	// TODO: integrate with cluster configuration in future versions.
	_ = name
	return ""
}

// PromResult 封装 Prometheus 查询返回的结果。
type PromResult struct {
	value    model.Value
	warnings promv1.Warnings
}

// Raw 返回底层的 Prometheus model.Value 与 warnings。
func (r *PromResult) Raw() (model.Value, promv1.Warnings) {
	if r == nil {
		return nil, nil
	}
	return r.value, r.warnings
}

// Sample 表示一个时间序列样本点（向量模式下）。
type Sample struct {
	Metric    map[string]string
	Value     float64
	Timestamp time.Time
}

// Series 表示一条时间序列（矩阵模式下）。
type Series struct {
	Metric  map[string]string
	Samples []SamplePoint
}

// SamplePoint 表示矩阵模式下每个时间点的值。
type SamplePoint struct {
	Timestamp time.Time
	Value     float64
}

// AsScalar 试图将结果解析为单个标量值。
func (r *PromResult) AsScalar() (float64, bool) {
	if r == nil || r.value == nil {
		return 0, false
	}
	switch v := r.value.(type) {
	case *model.Scalar:
		return float64(v.Value), true
	case model.Vector:
		if len(v) == 1 {
			return float64(v[0].Value), true
		}
	}
	return 0, false
}

// AsVector 将结果转换为 Sample 列表（仅在结果为向量时有效）。
func (r *PromResult) AsVector() []Sample {
	if r == nil || r.value == nil {
		return nil
	}
	vec, ok := r.value.(model.Vector)
	if !ok {
		return nil
	}
	out := make([]Sample, 0, len(vec))
	for _, s := range vec {
		out = append(out, Sample{
			Metric:    metricToMap(s.Metric),
			Value:     float64(s.Value),
			Timestamp: s.Timestamp.Time(),
		})
	}
	return out
}

// AsMatrix 将结果转换为 Series 列表（仅在结果为矩阵时有效）。
func (r *PromResult) AsMatrix() []Series {
	if r == nil || r.value == nil {
		return nil
	}
	mat, ok := r.value.(model.Matrix)
	if !ok {
		return nil
	}
	out := make([]Series, 0, len(mat))
	for _, s := range mat {
		series := Series{
			Metric:  metricToMap(s.Metric),
			Samples: make([]SamplePoint, 0, len(s.Values)),
		}
		for _, v := range s.Values {
			series.Samples = append(series.Samples, SamplePoint{
				Timestamp: v.Timestamp.Time(),
				Value:     float64(v.Value),
			})
		}
		out = append(out, series)
	}
	return out
}

// AsString 以字符串形式返回结果，便于调试。
func (r *PromResult) AsString() string {
	if r == nil || r.value == nil {
		return ""
	}
	return r.value.String()
}

// metricToMap 将 Prometheus 的 Metric 转换为普通 map。
func metricToMap(m model.Metric) map[string]string {
	if len(m) == 0 {
		return nil
	}
	res := make(map[string]string, len(m))
	for k, v := range m {
		res[string(k)] = string(v)
	}
	return res
}

// filterValueByLabels 根据 labelMatchers 对 Vector/Matrix 结果做过滤。
func filterValueByLabels(v model.Value, matchers map[string]string) model.Value {
	if len(matchers) == 0 || v == nil {
		return v
	}
	switch typed := v.(type) {
	case model.Vector:
		out := make(model.Vector, 0, len(typed))
		for _, s := range typed {
			if metricMatches(s.Metric, matchers) {
				out = append(out, s)
			}
		}
		return out
	case model.Matrix:
		out := make(model.Matrix, 0, len(typed))
		for _, s := range typed {
			if metricMatches(s.Metric, matchers) {
				out = append(out, s)
			}
		}
		return out
	default:
		return v
	}
}

// metricMatches 判断某条 Metric 是否满足所有 label 匹配条件。
func metricMatches(m model.Metric, matchers map[string]string) bool {
	for k, v := range matchers {
		if m[model.LabelName(k)] != model.LabelValue(v) {
			return false
		}
	}
	return true
}

// QueryScalar 执行瞬时查询并期望返回单个标量结果。
func (q *PromQuery) QueryScalar() (float64, error) {
	res, err := q.Query()
	if err != nil {
		return 0, err
	}
	val, ok := res.AsScalar()
	if !ok {
		return 0, fmt.Errorf("result is not scalar")
	}
	return val, nil
}

// QueryVector 执行瞬时查询并将结果直接转换为 Sample 列表。
func (q *PromQuery) QueryVector() ([]Sample, error) {
	res, err := q.Query()
	if err != nil {
		return nil, err
	}
	return res.AsVector(), nil
}

// QueryMatrix 执行瞬时查询并将结果直接转换为 Series 列表。
func (q *PromQuery) QueryMatrix() ([]Series, error) {
	res, err := q.Query()
	if err != nil {
		return nil, err
	}
	return res.AsMatrix(), nil
}

// QuerySum 对当前表达式的结果进行 sum 聚合，并返回标量。
func (q *PromQuery) QuerySum() (float64, error) {
	vec, err := q.QueryVector()
	if err != nil {
		return 0, err
	}
	var sum float64
	for _, s := range vec {
		sum += s.Value
	}
	return sum, nil
}

// QueryAvg 对当前表达式的结果进行 avg 聚合，并返回标量。
func (q *PromQuery) QueryAvg() (float64, error) {
	vec, err := q.QueryVector()
	if err != nil {
		return 0, err
	}
	if len(vec) == 0 {
		return 0, nil
	}
	var sum float64
	for _, s := range vec {
		sum += s.Value
	}
	return sum / float64(len(vec)), nil
}

// QuerySumBy 按给定 label 组合进行 sum 聚合，返回每个分组的结果。
func (q *PromQuery) QuerySumBy(labels ...string) (map[string]float64, error) {
	vec, err := q.QueryVector()
	if err != nil {
		return nil, err
	}
	res := make(map[string]float64)
	for _, s := range vec {
		key := buildGroupKey(s.Metric, labels)
		res[key] += s.Value
	}
	return res, nil
}

// QueryAvgBy 按给定 label 组合进行 avg 聚合，返回每个分组的结果。
func (q *PromQuery) QueryAvgBy(labels ...string) (map[string]float64, error) {
	vec, err := q.QueryVector()
	if err != nil {
		return nil, err
	}
	sum := make(map[string]float64)
	count := make(map[string]int)
	for _, s := range vec {
		key := buildGroupKey(s.Metric, labels)
		sum[key] += s.Value
		count[key]++
	}
	res := make(map[string]float64, len(sum))
	for k, v := range sum {
		if c := count[k]; c > 0 {
			res[k] = v / float64(c)
		}
	}
	return res, nil
}

// buildGroupKey 根据指定的 labels 从 Metric 中取值并构造分组 key。
func buildGroupKey(metric map[string]string, labels []string) string {
	if len(labels) == 0 {
		return ""
	}
	parts := make([]string, 0, len(labels))
	for _, l := range labels {
		parts = append(parts, fmt.Sprintf("%s=%s", l, metric[l]))
	}
	return strings.Join(parts, ",")
}
