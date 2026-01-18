# Prometheus 查询 SDK 设计文档（草案）

## 1. 设计目标

- **统一入口**：基于 `kom.Cluster` 抽象，提供从“集群 → Prometheus → 查询”的统一链式调用。
- **简单易用**：用户只关心 PromQL 语句和查询参数，不需要关心底层 HTTP 客户端和地址拼接。
- **支持多集群 / 多 Prometheus 实例**：同一个集群可以有多个 Prometheus（如本地 Prom、Thanos 等），也可以针对不同集群分别查询。
- **与 Go 风格一致**：大量使用 `context.Context`、链式 builder、明确的错误返回。

---

## 2. 基本使用方式概览

### 2.1 从默认集群执行瞬时查询

```go
ctx := context.Background()

res, err := kom.
    DefaultCluster().
    WithContext(ctx).
    Prometheus().
    WithInClusterEndpoint("monitoring", "prometheus").
    Expr(`sum(rate(http_requests_total[5m]))`).
    Query()
```

### 2.2 从指定集群执行查询

```go
res, err := kom.
    Cluster("prod-cn-beijing").
    WithContext(ctx).
    Prometheus().
    WithInClusterEndpoint("monitoring", "prometheus").
    Expr(`up`).
    Query()
```
 

### 2.3 使用临时指定 Prometheus 地址

无需事先在配置中注册，直接指定地址：

```go
res, err := kom.
    DefaultCluster().
    WithContext(ctx).
    Prometheus().
    WithAddress("http://prometheus.monitoring.svc:9090").
    Expr(`sum(kube_pod_container_status_ready)`).
    Query()
```

---

## 3. API 分层设计（用户视角）

### 3.1 Cluster 层

假定现有能力：

- **`kom.DefaultCluster()`**：返回默认集群 `Cluster`。
- **`kom.Cluster(name string)`**：按名称/ID 返回指定 `Cluster`。

在此基础上，为 `Cluster` 新增：

- **`Cluster.Prometheus()`**  
  返回“Prometheus 服务访问器”，用于在该集群上下文下构造 Prometheus client 和查询。

示例（同时演示 `WithContext` 传入上下文）：

```go
ctx := context.Background()

cluster := kom.DefaultCluster().
    WithContext(ctx)

prom := cluster.Prometheus()
_ = prom // 后续获取 client、构造查询
```

> 通过在 `Cluster` 层调用 `WithContext(ctx)`，后续的 Prometheus 查询（`Query` / `QueryRange` / `QuerySum` / `QueryXXX` 等）都会复用同一个 `context.Context`。

### 3.2 Prometheus 服务层

从 `cluster.Prometheus()` 获得的对象，向外暴露：

- **`WithInClusterEndpoint("monitoring", "prometheus")`**
  - 返回当前集群的**默认** Prometheus client。
  - 默认 client 的地址和认证等信息来源于 kom 的配置或集群元数据（annotation/configmap 等）。
 
- **`WithAddress(addr string)`**
  - 基于给定 HTTP 地址构造一个临时 client。

示例：

```go
client := kom.
    DefaultCluster().
    Prometheus().
    WithInClusterEndpoint("monitoring", "prometheus")
 
tmpClient := kom.
    DefaultCluster().
    Prometheus().
    WithAddress("http://10.0.0.1:9090")
```

### 3.3 Prometheus Client 层

Prometheus client 用于构建查询：

- **`Expr(expr string)`**
  - 接收一段 PromQL 字符串。
  - 返回一个“查询构建器”（Query Builder）。

示例：

```go
q := kom.
    DefaultCluster().
    Prometheus().
    WithInClusterEndpoint("monitoring", "prometheus").
    Expr(`sum(rate(http_requests_total{job="api"}[5m]))`)
```

> 第一版以纯 `Expr` 为主，后续可扩展更高级的 Metric Builder（如 `Metric().Labels().Rate()` 等）。

---

## 4. 查询构建器（Query Builder）层

`Expr` 返回的 Query Builder 负责设置查询选项并执行查询。

### 4.1 瞬时查询（Instant Query）

- **`Query()`**  
  在**当前时间点**执行 PromQL 查询，使用链路上通过 `WithContext` 设置的 `context.Context`。

- （可选扩展）**`QueryAt(ts time.Time)`**  
  在指定时间点执行瞬时查询，同样复用链路上的 `context.Context`。

示例：

```go
res, err := kom.
    DefaultCluster().
    WithContext(ctx).
    Prometheus().
    WithInClusterEndpoint("monitoring", "prometheus").
    Expr(`sum(rate(http_requests_total[5m]))`).
    Query()
```

> 后续可在 Query Builder 上增加若干快捷方法（如 `QuerySum()` / `QueryXXX()`），同样不再接收 `ctx` 参数，而是复用链路上的 `WithContext`。

### 4.2 范围查询（Range Query）

- 统一采用：**`QueryRange(start, end time.Time, step time.Duration)`**  
  使用链路上通过 `WithContext` 设置的 `context.Context` 执行区间查询。

示例（以 `Range` 命名为例）：

```go
start := time.Now().Add(-1 * time.Hour)
end   := time.Now()
step  := time.Minute

res, err := kom.
    DefaultCluster().
    WithContext(ctx).
    Prometheus().
    WithInClusterEndpoint("monitoring", "prometheus").
    Expr(`sum(rate(http_requests_total[5m]))`).
    QueryRange(start, end, step)
```

### 4.3 附加选项（链式设置）

在 Query Builder 上可以附加一些可选设置：

- **`WithTimeout(d time.Duration)`**  
  设置单次查询的超时时间。

- **`WithDedup(enabled bool)`**  
  针对 Thanos / Cortex 等支持去重的后端。

- **`WithPartialResponse(enabled bool)`**  
  是否允许返回部分结果。

示例：

```go
res, err := kom.
    DefaultCluster().
    WithContext(ctx).
    Prometheus().
    WithInClusterEndpoint("monitoring", "prometheus").
    Expr(`sum(rate(http_requests_total[5m]))`).
    WithTimeout(3 * time.Second).
    QueryRange(start, end, time.Minute)
```

> 第一版可以先实现 `WithTimeout`，其他选项视后端对接情况逐步增加。

---

## 5. 查询结果设计（PromResult）

为提高易用性，接口层建议返回一个包装后的结果结构，如：

- **`type PromResult struct { ... }`**  
  封装底层 Prometheus HTTP API 的结果（类型可以是 `model.Value` 等）。

对外提供便捷方法，例如：

- **`Raw()`**  
  返回原始 Prometheus HTTP API 响应对象（或 JSON）。

- **`AsScalar() (float64, bool)`**  
  当结果为 scalar 时返回具体数值及是否成功转换。

- **`AsVector() []Sample`**  
  当结果为 vector 时，返回样本数组。

- **`AsMatrix() []Series`**  
  当结果为 matrix 时，返回多条时间序列。

- **`AsString() string`**  
  以人类可读形式返回结果，便于快速调试打印。

示例：

```go
res, err := kom.
    DefaultCluster().
    WithContext(ctx).
    Prometheus().
    WithInClusterEndpoint("monitoring", "prometheus").
    Expr(`up`).
    Query()
if err != nil {
    // 处理错误
    return
}

fmt.Println(res.AsString()) // 简单打印

samples := res.AsVector()
for _, s := range samples {
    fmt.Printf("metric=%v value=%v\n", s.Metric, s.Value)
}
```

> 设计倾向：**对外暴露 `PromResult` 这一层封装**，内部可以自由替换底层实现，避免直接泄露第三方库类型。

---

## 6. 错误处理约定

- **统一签名**：  
  - 瞬时查询：`Query() (*PromResult, error)`  
  - 范围查询：`QueryRange(start, end, step) (*PromResult, error)`

- **错误来源**：  
  - HTTP 层错误：4xx / 5xx。  
  - PromQL 语法错误。  
  - 查询超时、上下文取消等。

- **扩展方向**：  
  - 未来可定义 `PrometheusError` 接口或若干 sentinel errors：  
    - 例如 `ErrTimeout`、`ErrBadRequest` 等，便于调用方 `errors.Is` 判断。

---

## 7. 并发与复用约定

- **Cluster / Prometheus 服务对象 / Prometheus Client 对象**  
  设计目标：**可在多 goroutine 安全复用**（内部负责线程安全）。

- **Query Builder 对象**  
  建议：作为**一次性对象**使用，不保证同时在多个 goroutine 中并发调用。  
  可以支持在同一 goroutine 中多次调用 `Query()` / `QueryRange()`（例如不同时间点重复查询），但推荐“一次构建，一次使用”的模式。

示例：多个 goroutine 共享 client，各自构建 query：

```go
client := kom.
    DefaultCluster().
    WithContext(ctx).
    Prometheus().
    WithInClusterEndpoint("monitoring", "prometheus")

wg := sync.WaitGroup{}
for i := 0; i < 10; i++ {
    wg.Add(1)
    go func() {
        defer wg.Done()
        res, err := client.
            Expr(`up`).
            Query()
        // 处理 res/err
    }()
}
wg.Wait()
```

---

## 8. 与 K8s / 多集群关系

- **`Cluster.Prometheus()` 的职责**：  
  在当前集群上下文中解析出 Prometheus 的访问方式，可能是：
  - 集群内 `Service`：如 `prometheus-k8s.monitoring.svc:9090`；
  - 聚合查询层：如 Thanos Query / Query Frontend；
  - 外部暴露的地址。

- **多集群场景**：  
  不同集群可以有各自的 Prometheus 配置，通过 `kom.Cluster("...").Prometheus().WithInClusterEndpoint("monitoring", "prometheus")` 访问：

```go
prodRes,  _ := kom.Cluster("prod").WithContext(ctx).Prometheus().WithInClusterEndpoint("monitoring", "prometheus").Expr(`up`).Query()
stageRes, _ := kom.Cluster("stage").WithContext(ctx).Prometheus().WithInClusterEndpoint("monitoring", "prometheus").Expr(`up`).Query()
```
 
---

## 9. 综合示例

### 9.1 查询 HTTP QPS（瞬时）

```go
ctx := context.Background()

res, err := kom.
    DefaultCluster().
    WithContext(ctx).
    Prometheus().
    WithInClusterEndpoint("monitoring", "prometheus").
    Expr(`sum(rate(http_requests_total{job="api"}[5m]))`).
    Query()
if err != nil {
    log.Fatalf("query prometheus failed: %v", err)
}

fmt.Println("QPS:", res.AsScalar())
```

### 9.2 查询 CPU 使用率曲线（范围）

```go
ctx := context.Background()
start := time.Now().Add(-30 * time.Minute)
end   := time.Now()

res, err := kom.
    DefaultCluster().
    WithContext(ctx).
    Prometheus().
    WithInClusterEndpoint("monitoring", "prometheus").
    Expr(`sum(rate(container_cpu_usage_seconds_total{namespace="default"}[2m]))`).
    WithTimeout(5 * time.Second).
    QueryRange(start, end, 30*time.Second)
if err != nil {
    log.Fatalf("query range failed: %v", err)
}

series := res.AsMatrix()
for _, s := range series {
    fmt.Printf("metric=%v, points=%d\n", s.Metric, len(s.Samples))
}
```
 
---

## 10. 待确认问题

当前仍有少量设计细节可以在实现阶段再细化：

- **1）快捷方法设计（聚合 + 结果形态）**：
  - 在 Query Builder 上提供结果形态快捷方法：`QueryScalar()` / `QueryVector()` / `QueryMatrix()`，在不改变 PromQL 的前提下直接返回目标类型。
  - 在此基础上再提供带聚合语义的快捷方法：`QuerySum()` / `QueryAvg()` / `QueryMin()` / `QueryMax()` 等，对当前表达式包一层聚合并直接返回标量结果。
  - 对于需要按 label 维度聚合的场景，增加 `QuerySumBy(labels ...string)` / `QueryAvgBy(labels ...string)` 等方法，返回例如 `map[string]float64` 形式的聚合结果。
  - 以上所有方法均不接收 `ctx`，统一复用链路上的 `WithContext(ctx)`。

- **2）资源级快捷过滤（Pod / Deployment 等）**：
  - 在 Query Builder 上提供资源级过滤方法，例如：`ForPod(namespace, name string)`、`ForDeployment(namespace, name string)`，后续可扩展 `ForStatefulSet` / `ForDaemonSet` / `ForNode` 等。
  - 这些方法只负责在当前 PromQL 表达式上追加合适的 label 过滤（例如 `{namespace="default",pod="mypod"}` 或 `{namespace="default",deployment="my-deploy"}`），不改变上下文。
  - 资源过滤方法与 `Query` / `QueryRange` 以及聚合快捷方法可以自由组合，例如：

    ```go
    // 针对单个 Pod 的 QPS（sum 聚合 + Scalar 结果）
    qps, err := kom.
        DefaultCluster().
        WithContext(ctx).
        Prometheus().
        WithInClusterEndpoint("monitoring", "prometheus").
        Expr(`rate(http_requests_total[5m])`).
        ForPod("default", "mypod").
        QuerySum()

    // 针对某个 Deployment，按 Pod 维度聚合 QPS
    values, err := kom.
        DefaultCluster().
        WithContext(ctx).
        Prometheus().
        WithInClusterEndpoint("monitoring", "prometheus").
        Expr(`rate(http_requests_total[5m])`).
        ForDeployment("default", "my-deploy").
        QuerySumBy("pod")
    ```

- **3）结果封装层级**：当前文档采用对外暴露 `PromResult` 包装类型（内部隐藏 Prometheus 官方类型）的方案，后续可以根据实际使用情况再决定是否暴露更多底层类型或提供更多 helper。

- **4）多 Prometheus 实例配置方式**：每个集群下的多 Prometheus 实例（如 `default` / `thanos-global` 等）的配置载体（配置文件 / 注解 / 代码注册）可以在实现阶段结合现有 kom 配置体系再做详细设计。