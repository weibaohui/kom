# Token 注册集群功能

本文档介绍如何使用 token 来注册 Kubernetes 集群。

## 功能概述

KOM 提供了通过 token 注册集群的方法：

1. `RegisterByTokenWithServerAndID` - 带服务器地址的 token 注册

## 使用方法
 
### 带服务器地址的 Token 注册

```go
import "github.com/weibaohui/kom/kom"

// 完整注册（推荐方式）
token := "<YOUR_TOKEN>"
server := "https://your-cluster.example.com:6443"
clusterID := "production-cluster"
caData := "<YOUR_CA_DATA>"
// 当需要指定 CA 证书时，使用注册期 Option：
kubectl, err := kom.Clusters().RegisterByTokenWithServerAndID(
    token,
    server,
    clusterID,
    kom.RegisterCACert([]byte(caData)),
)
if err != nil {
    log.Fatalf("Failed to register cluster: %v", err)
}
```

**参数说明：**
- `token`: Kubernetes 集群的访问令牌 (Bearer Token)
- `server`: Kubernetes API 服务器地址
- `clusterID`: 集群的唯一标识符
- `caData`: 可选的 CA 证书数据，用于启用 TLS 验证
 
## 获取 Token

### 方法 1: 使用 Service Account

```bash
# 创建 Service Account
kubectl create serviceaccount kom-user

# 创建 ClusterRoleBinding
kubectl create clusterrolebinding kom-user-binding \
  --clusterrole=cluster-admin \
  --serviceaccount=default:kom-user

# 获取 Token
kubectl create token kom-user
```

### 方法 2: 从现有 Secret 获取

```bash
# 查找 Service Account 的 Secret
kubectl get secrets

# 获取 Token
kubectl get secret <secret-name> -o jsonpath='{.data.token}' | base64 -d
```

## 错误处理

所有的 token 注册函数都会进行参数验证：

- `token` 不能为空
- `server` 不能为空（适用于带服务器地址的函数）
- `clusterID` 不能为空

如果参数验证失败，函数会返回相应的错误信息。

## 安全注意事项

1. **Token 安全**: 请妥善保管您的 token，不要在代码中硬编码
2. **TLS 验证**: 生产环境中建议启用 TLS 验证（`insecure=false`）
3. **权限控制**: 为 Service Account 分配最小必要权限

## 示例代码

完整的示例代码可以在 `example/token_register_test.go` 文件中找到。

## 常见问题

### Q: Token 过期了怎么办？
A: 需要重新生成 token 并重新注册集群。

### Q: 如何验证注册是否成功？
A: 可以尝试调用 kubectl 的方法，例如：
```go
pods, err := kubectl.Pod().List()
if err != nil {
    log.Printf("Failed to list pods: %v", err)
} else {
    log.Printf("Successfully connected, found %d pods", len(pods))
}
```

### Q: 支持哪些类型的 Kubernetes 集群？
A: 支持所有标准的 Kubernetes 集群，包括：
- 自建集群
- 云服务商托管集群（EKS、GKE、AKS 等）
- 本地开发集群（minikube、kind 等）