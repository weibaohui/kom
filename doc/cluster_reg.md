# KOM 集群注册与注册期 Option 高级用法

本文档介绍 KOM 的集群注册入口与“注册期 Option”高级用法。注册期选项仅在注册流程中被消费，不会影响后续 `Kubectl` 使用行为，也不会持久化到 `ClusterInst`。

## 统一的注册入口

所有注册入口统一接受可选参数 `opts ...RegisterOption`：

- `RegisterByConfig(config *rest.Config, opts ...RegisterOption) (*Kubectl, error)`
- `RegisterByConfigWithID(config *rest.Config, id string, opts ...RegisterOption) (*Kubectl, error)`
- `RegisterByPath(path string, opts ...RegisterOption) (*Kubectl, error)`
- `RegisterByPathWithID(path string, id string, opts ...RegisterOption) (*Kubectl, error)`
- `RegisterByString(str string, opts ...RegisterOption) (*Kubectl, error)`
- `RegisterByStringWithID(str string, id string, opts ...RegisterOption) (*Kubectl, error)`
- `RegisterInCluster(opts ...RegisterOption) (*Kubectl, error)`
- `RegisterAWSCluster(config *aws.EKSAuthConfig, opts ...RegisterOption) (*Kubectl, error)`
- `RegisterAWSClusterWithID(config *aws.EKSAuthConfig, clusterID string, opts ...RegisterOption) (*Kubectl, error)`
- `RegisterByTokenWithServerAndID(token string, server string, id string, opts ...RegisterOption) (*Kubectl, error)`

兼容包装：`RegisterByTokenWithServerAndIDLegacy(token, server, id, caData ...string)`，等价于传入 `RegisterCACert([]byte(caData[0]))`，建议迁移到新签名。

## 可用的注册期选项

- `RegisterProxyURL(string)`：设置 HTTP 代理（例如 `http://127.0.0.1:7890`）
- `RegisterProxyFunc(func(*http.Request) (*url.URL, error))`：自定义代理函数，优先级高于 `RegisterProxyURL`
- `RegisterTimeout(time.Duration)`：设置请求超时
- `RegisterQPS(float32)`：设置 QPS（默认 `200`）
- `RegisterBurst(int)`：设置 Burst（默认 `2000`）
- `RegisterUserAgent(string)`：设置 `User-Agent`
- `RegisterTLSInsecure()`：启用不安全 TLS（跳过证书校验）
- `RegisterCACert([]byte)`：设置 CA 证书（同时关闭 `Insecure`）
- `RegisterImpersonation(user string, groups []string, extra map[string][]string)`：设置冒充用户配置
- `RegisterDisableCRDWatch()`：禁用注册期的 CRD 监听与刷新（初始化更轻量）
- `RegisterCacheConfig(*ristretto.Config[string, any])`：自定义集群缓存配置

## 常用场景示例

### 1. Kubeconfig 路径注册（代理/QPS/Burst/关闭 CRD 监听）

```go
import "github.com/weibaohui/kom/kom"

defaultKubeConfig := "/Users/you/.kube/config"
_, err := kom.Clusters().RegisterByPathWithID(
    defaultKubeConfig,
    "default",
    kom.RegisterProxyURL("http://localhost:7890"),
    kom.RegisterDisableCRDWatch(),
    kom.RegisterBurst(400),
)
if err != nil { /* handle */ }
```

示例来源：`example/connect.go` 中的高级用法。

### 2. Token 注册 + 指定 CA + 超时

```go
import (
    "time"
    "github.com/weibaohui/kom/kom"
)

token  := "<YOUR_TOKEN>"
server := "https://your-cluster.example.com:6443"
caData := "<PEM_CA_DATA>"

kubectl, err := kom.Clusters().RegisterByTokenWithServerAndID(
    token,
    server,
    "prod",
    kom.RegisterCACert([]byte(caData)),
    kom.RegisterTimeout(10*time.Second),
    kom.RegisterUserAgent("kom-client/1.0"),
)
if err != nil { /* handle */ }
```

### 3. InCluster 注册（禁用 CRD 监听）

```go
import "github.com/weibaohui/kom/kom"

kubectl, err := kom.Clusters().RegisterInCluster(
    kom.RegisterDisableCRDWatch(),
)
if err != nil { /* handle */ }
```

### 4. AWS EKS 注册（统一传递选项）

```go
import (
    "time"
    "github.com/weibaohui/kom/kom"
    "github.com/weibaohui/kom/kom/aws"
)

cfg := &aws.EKSAuthConfig{
    Region:      "us-east-1",
    ClusterName: "my-eks",
    // 其他必要配置...
}
kubectl, err := kom.Clusters().RegisterAWSCluster(
    cfg,
    kom.RegisterTimeout(15 * time.Second),
    kom.RegisterQPS(100),
)
if err != nil { /* handle */ }
```

### 5. Impersonation（设置冒充用户）

```go
import "github.com/weibaohui/kom/kom"

_, err := kom.Clusters().RegisterByPathWithID(
    "/Users/you/.kube/config",
    "default",
    kom.RegisterImpersonation("alice", []string{"devs"}, map[string][]string{"tenant": {"blue"}}),
)
if err != nil { /* handle */ }
```

### 6. 自定义缓存配置

```go
import (
    "github.com/dgraph-io/ristretto/v2"
    "github.com/weibaohui/kom/kom"
)

cacheCfg := &ristretto.Config[string, any]{
    NumCounters: 1e6,
    MaxCost:     1 << 28, // 256MB
    BufferItems: 64,
}
_, err := kom.Clusters().RegisterByPathWithID(
    "/Users/you/.kube/config",
    "default",
    kom.RegisterCacheConfig(cacheCfg),
)
if err != nil { /* handle */ }
```

## 行为与优先级

- 默认速率限制：`QPS=200`、`Burst=2000`，若提供选项则覆盖默认值。
- `RegisterProxyFunc` 优先于 `RegisterProxyURL`（两者同时给定时使用函数）。
- `RegisterCACert` 会设置 `CAData` 并关闭 `Insecure`；同时使用 `RegisterTLSInsecure` 时，以 `RegisterCACert` 为准。
- `RegisterDisableCRDWatch` 仅影响注册期是否开启 CRD 监听与刷新。
- 选项仅在注册期生效，不会持久化到 `ClusterInst` 或 `Kubectl`。

## 兼容性与迁移

- 旧的 token 注册可变参数 `caData` 可使用兼容包装：
  - `RegisterByTokenWithServerAndIDLegacy(token, server, id, caData...)`
  - 等价为：`RegisterByTokenWithServerAndID(token, server, id, RegisterCACert([]byte(caData)))`
- 原不带 `opts` 的调用方式保持可用（`opts` 为可选）。

## 常见问题（FAQ）

- 为什么选项不生效？
  - 确认在对应的 `RegisterXXX(..., opts...)` 调用中确实传递了选项；代理仅支持标准 HTTP/HTTPS 代理，复杂逻辑请使用 `RegisterProxyFunc`。
- 可以同时使用 `TLSInsecure` 与 `CACert` 吗？
  - 可以，但若设置了 `RegisterCACert`，将优先进行证书校验（关闭 `Insecure`）。
- 禁用 CRD 监听有什么影响？
  - 初始化期间不会通过 CRD 变化刷新 API 资源发现，适用于资源类型稳定或强调轻量初始化的场景。

---

如需扩展更多注册期参数（例如更细的传输层配置或自定义 Dialer），欢迎提出需求以补充新的 `RegisterXXX` 选项构造函数。