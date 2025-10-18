# Multi-Cluster Support in SSE Mode

## Overview

kom now supports managing multiple Kubernetes clusters in SSE mode, allowing configuration and connection to multiple clusters through kubeconfig files or content.

## Features

- Support for kubeconfig file path registration
- Support for kubeconfig content registration  
- Batch loading of kubeconfig files from directory
- Dynamic cluster registration and unregistration
- Default cluster designation
- Full cluster information display
- Complete compatibility with existing MCP tools

## Configuration

### Method 1: Manual kubeconfig file specification

```go
kubeconfigs := []mcp.KubeconfigConfig{
    {
        ID:        "production",
        Path:      "/path/to/production-kubeconfig.yaml",
        IsDefault: true,
    },
    {
        ID:   "staging",
        Path: "/path/to/staging-kubeconfig.yaml",
    },
}

cfg := mcp.ServerConfig{
    Name:        "kom mcp server",
    Version:     "0.0.1",
    Port:        9096,
    Mode:        mcp.ServerModeSSE,
    Kubeconfigs: kubeconfigs,
}
```

### Method 2: Directory auto-discovery

```go
kubeconfigs, err := mcp.LoadKubeconfigsFromDirectory("/path/to/kubeconfigs")
if err != nil {
    log.Fatal(err)
}

cfg := mcp.ServerConfig{
    Mode:        mcp.ServerModeSSE,
    Kubeconfigs: kubeconfigs,
}
```

### Method 3: Kubeconfig content

```go
kubeconfigs := []mcp.KubeconfigConfig{
    {
        ID:      "cluster1",
        Content: `apiVersion: v1
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
    token: your-token-here`,
        IsDefault: true,
    },
}
```

## Dynamic Cluster Management

### Register cluster at runtime

```bash
register_k8s_cluster(
    cluster_id="new-cluster",
    kubeconfig_path="/path/to/kubeconfig.yaml"
)
```

### Unregister cluster

```bash
unregister_k8s_cluster(cluster_id="old-cluster")
```

### List clusters

```bash
list_k8s_clusters()
```

## Usage Examples

### Basic multi-cluster setup

```go
package main

import (
    "github.com/weibaohui/kom/mcp"
)

func main() {
    kubeconfigs := []mcp.KubeconfigConfig{
        {ID: "prod", Path: "/kubeconfigs/prod.yaml", IsDefault: true},
        {ID: "staging", Path: "/kubeconfigs/staging.yaml"},
    }

    cfg := mcp.ServerConfig{
        Name:        "kom multi-cluster",
        Version:     "1.0.0",
        Port:        9096,
        Mode:        mcp.ServerModeSSE,
        Kubeconfigs: kubeconfigs,
    }

    mcp.RunMCPServerWithOption(&cfg)
}
```

## Testing

Run the test suite to verify multi-cluster functionality:

```bash
go test ./example -v -run TestSimpleMultiClusterSSE
```

## Notes

- All existing MCP tools work with multi-cluster setup
- Clusters are identified by their ID
- Default cluster is used when no specific cluster is specified
- Network errors are handled gracefully