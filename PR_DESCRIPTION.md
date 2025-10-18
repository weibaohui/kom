# Multi-Cluster Support for SSE Mode

## Overview

This PR implements multi-cluster support for kom's SSE mode, enabling a single kom deployment to serve multiple Kubernetes clusters simultaneously. Addresses issue #43.

## Changes

### Core Implementation
- Added `KubeconfigConfig` struct for cluster configuration
- Added `LoadKubeconfigsFromDirectory()` helper function
- Added `initializeMultiCluster()` function for startup cluster registration
- Updated `ServerConfig` to include `Kubeconfigs` field
- Added cluster management tools (register/unregister)

### Files Modified
- `mcp/server.go` - Multi-cluster configuration and initialization
- `mcp/tools/cluster/cluster_list.go` - Enhanced cluster listing
- `mcp/tools/cluster/reg.go` - Dynamic cluster management tools
- `main.go` - Updated with multi-cluster example
- `doc/multi_cluster_sse.md` - Documentation
- `example/simple_test.go` - Basic tests

## Usage

```go
kubeconfigs := []mcp.KubeconfigConfig{
    {ID: "prod", Path: "/path/to/prod.yaml", IsDefault: true},
    {ID: "staging", Path: "/path/to/staging.yaml"},
}

cfg := mcp.ServerConfig{
    Mode:        mcp.ServerModeSSE,
    Kubeconfigs: kubeconfigs,
}
```

## Testing

- Multi-cluster registration: ✅ PASS
- Directory loading: ✅ PASS
- Dynamic cluster management: ✅ PASS

## Backward Compatibility

Fully backward compatible - existing single-cluster configurations continue to work.

Closes #43