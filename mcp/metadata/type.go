package metadata

import (
	"github.com/mark3labs/mcp-go/server"
)

// ResourceMetadata 封装资源的元数据信息
type ResourceMetadata struct {
	Cluster   string
	Namespace string
	Name      string
	Group     string
	Version   string
	Kind      string
}
type ServerConfig struct {
	Name          string
	Version       string
	Port          int
	ServerOptions []server.ServerOption
	SSEOption     []server.SSEOption
	Metadata      map[string]string // 元数据
	AuthKey       string            // 认证key
	AuthRoleKey   string            // 认证key
	Mode          MCPServerMode     // 运行模式 sse,stdio
}
type MCPServerMode string

const (
	MCPServerModeSSE   MCPServerMode = "sse"
	MCPServerModeStdio MCPServerMode = "stdio"
)

type ResourceInfo struct {
	Group      string
	Version    string
	Kind       string
	Namespaced bool
}
