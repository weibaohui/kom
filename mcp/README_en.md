# MCP (Microservice Control Panel) User Guide

[English](README_en.md) | [中文](README.md)

## Getting Started

### Requirements
- Go 1.16 or higher
- Configured Kubernetes cluster
- Default Kubeconfig file

### Start Command
```go
mcp.RunMCPServer("kom mcp server", "0.0.1", 3619)
```

## Features

- Multi-cluster Management: Support managing multiple Kubernetes clusters simultaneously
- Dynamic Resource Operations: Support CRUD operations on various Kubernetes resources
- Event Monitoring: Real-time cluster event viewing
- Resource Description: Get detailed resource information

## API Interfaces

### Cluster Management

#### List Clusters
- API Name: `list_clusters`
- Description: List all registered Kubernetes clusters
- Parameters: None
- Returns: Cluster list containing cluster names

### Dynamic Resource Operations

#### List Resources
- API Name: `list_k8s_resource`
- Description: List Kubernetes resources by cluster and resource type
- Parameters:
  - cluster: Cluster where the resources are running (use empty string for default cluster)
  - namespace: Namespace of the resources (optional for cluster-scoped resources)
  - group: API group of the resource
  - version: API version of the resource
  - kind: Kind of the resource
  - label: Label selector to filter resources (e.g. app=k8m)

#### Get Resource
- API Name: `get_k8s_resource`
- Description: Get a specific Kubernetes resource
- Parameters:
  - cluster: Cluster name
  - namespace: Namespace
  - group: API group
  - version: API version
  - kind: Resource kind
  - name: Resource name

#### Delete Resource
- API Name: `delete_k8s_resource`
- Description: Delete a specific Kubernetes resource
- Parameters:
  - cluster: Cluster name
  - namespace: Namespace
  - group: API group
  - version: API version
  - kind: Resource kind
  - name: Resource name

### Event Monitoring

#### List Events
- API Name: `list_k8s_event`
- Description: List Kubernetes events by cluster and namespace
- Parameters:
  - cluster: Cluster where the events are running (use empty string for default cluster)
  - namespace: Namespace of the events (optional)
  - involvedObjectName: Filter events by involved object name

## Usage Examples

### List All Clusters
```json
{
  "tool": "list_clusters"
}
```

### List All Pods in Default Namespace
```json
{
  "tool": "list_k8s_resource",
  "params": {
    "cluster": "",
    "namespace": "default",
    "group": "",
    "version": "v1",
    "kind": "Pod"
  }
}
```

### View Events for a Specific Pod
```json
{
  "tool": "list_k8s_event",
  "params": {
    "cluster": "",
    "namespace": "default",
    "involvedObjectName": "my-pod"
  }
}
```

## Notes

1. When using dynamic resource operations, ensure to provide correct resource group, version, and kind information
2. For cluster-scoped resources, the namespace parameter can be omitted
3. When using label selectors, ensure to use the correct label format

## AI Tool Integration

### Claude Desktop
1. Open Claude Desktop settings panel
2. Add MCP Server address in the API configuration area
3. Enable SSE event listening function
4. Verify connection status

### Cursor
1. Enter Cursor settings interface
2. Find extension service configuration option
3. Add MCP Server URL (e.g., http://localhost:3619/sse)
4. Enable real-time event notifications

### Windsurf
1. Access configuration center
2. Set API server address
3. Enable real-time event notifications
4. Test connection

### Common Issues
1. Ensure MCP Server is running and port is accessible
2. Check if network connection is normal
3. Verify if SSE connection is established successfully
4. Check tool logs to troubleshoot connection issues

## Contributing

Issues and Pull Requests are welcome to help improve MCP.