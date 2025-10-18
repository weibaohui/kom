# Multi-Cluster Support in SSE Mode

本文档介绍如何在SSE模式下使用kom支持多个Kubernetes集群。

## 概述

kom现在支持在SSE模式下同时管理多个Kubernetes集群，允许通过kubeconfig文件或内容来配置和连接多个集群。这个功能使得kom可以部署一次就服务多个k8s集群。

## 功能特性

- ✅ 支持通过kubeconfig文件路径注册集群
- ✅ 支持通过kubeconfig内容注册集群
- ✅ 支持从目录批量加载kubeconfig文件
- ✅ 支持动态注册和注销集群
- ✅ 支持指定默认集群
- ✅ 完整的集群信息展示
- ✅ 与现有MCP工具完全兼容

## 配置方式

### 方式1：手动指定kubeconfig文件

```go
package main

import (
    "github.com/weibaohui/kom/mcp"
)

func main() {
    cfg := &mcp.ServerConfig{
        Name:    "kom multi-cluster server",
        Version: "1.0.0",
        Port:    9096,
        Mode:    mcp.ServerModeSSE,
        Kubeconfigs: []mcp.KubeconfigConfig{
            {
                ID:        "cluster1",
                Path:      "/path/to/cluster1-kubeconfig.yaml",
                IsDefault: true,
            },
            {
                ID:   "cluster2",
                Path: "/path/to/cluster2-kubeconfig.yaml",
            },
        },
    }
    
    mcp.RunMCPServerWithOption(&cfg)
}
```

### 方式2：从目录自动加载

```go
package main

import (
    "github.com/weibaohui/kom/mcp"
    "k8s.io/klog/v2"
)

func main() {
    // 从目录加载所有kubeconfig文件
    configs, err := mcp.LoadKubeconfigsFromDirectory("/path/to/kubeconfigs")
    if err != nil {
        klog.Errorf("Failed to load kubeconfigs: %v", err)
        return
    }

    cfg := &mcp.ServerConfig{
        Name:        "kom multi-cluster server",
        Version:     "1.0.0",
        Port:        9096,
        Mode:        mcp.ServerModeSSE,
        Kubeconfigs: configs,
    }
    
    mcp.RunMCPServerWithOption(&cfg)
}
```

### 方式3：使用kubeconfig内容

```go
package main

import (
    "github.com/weibaohui/kom/mcp"
)

func main() {
    kubeconfigContent := `apiVersion: v1
clusters:
- cluster:
    server: https://cluster1.example.com:6443
  name: cluster1
contexts:
- context:
    cluster: cluster1
    user: admin
  name: cluster1
current-context: cluster1
kind: Config
users:
- name: admin
  user:
    token: your-token-here`

    cfg := &mcp.ServerConfig{
        Name:    "kom multi-cluster server",
        Version: "1.0.0",
        Port:    9096,
        Mode:    mcp.ServerModeSSE,
        Kubeconfigs: []mcp.KubeconfigConfig{
            {
                ID:        "cluster1",
                Content:   kubeconfigContent,
                IsDefault: true,
            },
        },
    }
    
    mcp.RunMCPServerWithOption(&cfg)
}
```

## 动态集群管理

kom提供了MCP工具来动态管理集群，无需重启服务：

### 列出所有集群

```bash
# 通过MCP客户端调用
{
  "method": "tools/call",
  "params": {
    "name": "list_k8s_clusters"
  }
}
```

### 注册新集群

```bash
# 通过kubeconfig文件
{
  "method": "tools/call",
  "params": {
    "name": "register_k8s_cluster",
    "arguments": {
      "cluster_id": "new-cluster",
      "kubeconfig_path": "/path/to/kubeconfig.yaml",
      "is_default": false
    }
  }
}

# 通过kubeconfig内容
{
  "method": "tools/call",
  "params": {
    "name": "register_k8s_cluster",
    "arguments": {
      "cluster_id": "new-cluster",
      "kubeconfig_content": "apiVersion: v1\n...",
      "is_default": false
    }
  }
}
```

### 注销集群

```bash
{
  "method": "tools/call",
  "params": {
    "name": "unregister_k8s_cluster",
    "arguments": {
      "cluster_id": "cluster-to-remove"
    }
  }
}
```

## 使用多集群

一旦配置了多个集群，所有现有的MCP工具都会自动支持多集群操作：

### 指定集群操作

```bash
# 获取指定集群的Pod列表
{
  "method": "tools/call",
  "params": {
    "name": "list_k8s_pods",
    "arguments": {
      "cluster": "cluster1",
      "namespace": "default"
    }
  }
}

# 获取指定集群的节点信息
{
  "method": "tools/call",
  "params": {
    "name": "list_k8s_nodes",
    "arguments": {
      "cluster": "cluster2"
    }
  }
}
```

### 默认集群操作

如果不指定`cluster`参数，系统会使用默认集群：

```bash
# 使用默认集群
{
  "method": "tools/call",
  "params": {
    "name": "list_k8s_pods",
    "arguments": {
      "namespace": "default"
    }
  }
}
```

## 配置结构

### ServerConfig

```go
type ServerConfig struct {
    Name          string
    Version       string
    Port          int
    ServerOptions []server.ServerOption
    SSEOption     []server.SSEOption
    Metadata      map[string]string
    AuthKey       string
    Mode          ServerMode
    Kubeconfigs   []KubeconfigConfig // 新增：多集群配置
}
```

### KubeconfigConfig

```go
type KubeconfigConfig struct {
    ID       string // 集群ID，用于标识集群
    Path     string // kubeconfig文件路径
    Content  string // kubeconfig内容（与Path二选一）
    IsDefault bool  // 是否为默认集群
}
```

## 环境变量支持

可以通过环境变量来配置kubeconfig目录：

```bash
export KUBECONFIG_DIR="/etc/kom/kubeconfigs"
export KOM_DEFAULT_CLUSTER="production"
```

## 目录结构示例

```
/etc/kom/
├── kubeconfigs/
│   ├── production.yaml
│   ├── staging.yaml
│   └── development.yaml
└── kom-server
```

## 错误处理

系统会处理以下错误情况：

- kubeconfig文件不存在
- kubeconfig格式错误
- 集群连接失败
- 重复的集群ID
- 无效的集群ID

## 日志和监控

系统会记录以下信息：

- 集群注册成功/失败
- 集群连接状态
- 动态集群管理操作
- 错误和警告信息

## 性能考虑

- 集群连接使用连接池
- 支持并发操作多个集群
- 自动资源清理
- 内存使用优化

## 安全考虑

- kubeconfig文件权限控制
- 敏感信息加密存储
- 访问权限验证
- 审计日志记录

## 故障排除

### 常见问题

1. **集群连接失败**
   - 检查kubeconfig文件格式
   - 验证网络连接
   - 确认认证信息

2. **集群ID冲突**
   - 使用唯一的集群ID
   - 检查现有集群列表

3. **默认集群问题**
   - 确保至少有一个集群被标记为默认
   - 检查集群状态

### 调试命令

```bash
# 查看所有已注册的集群
curl -X POST http://localhost:9096/sse \
  -H "Content-Type: application/json" \
  -d '{"method": "tools/call", "params": {"name": "list_k8s_clusters"}}'

# 检查集群状态
kubectl --kubeconfig=/path/to/kubeconfig.yaml cluster-info
```

## 升级指南

从单集群版本升级到多集群版本：

1. 更新代码到最新版本
2. 配置`Kubeconfigs`字段
3. 重启服务
4. 验证集群连接

## 示例和测试

参考 `example/multi_cluster_test.go` 文件查看完整的示例和测试用例。

## 贡献

欢迎提交Issue和Pull Request来改进多集群功能。
