package mcp

import (
	"fmt"

	"github.com/mark3labs/mcp-go/server"
	"github.com/weibaohui/kom/mcp/tools"
	"github.com/weibaohui/kom/mcp/tools/cluster"
	"github.com/weibaohui/kom/mcp/tools/deployment"
	"github.com/weibaohui/kom/mcp/tools/dynamic"
	"github.com/weibaohui/kom/mcp/tools/event"
	"github.com/weibaohui/kom/mcp/tools/ingressclass"
	"github.com/weibaohui/kom/mcp/tools/node"
	"github.com/weibaohui/kom/mcp/tools/ns"
	"github.com/weibaohui/kom/mcp/tools/pod"
	"github.com/weibaohui/kom/mcp/tools/storageclass"
	"github.com/weibaohui/kom/mcp/tools/yaml"
	"k8s.io/klog/v2"
)

type ServerConfig struct {
	Name          string
	Version       string
	Port          int
	ServerOptions []server.ServerOption
	SSEOption     []server.SSEOption
	Metadata      map[string]string // 元数据
	AuthKey       string            // 认证key
	Mode          ServerMode        // 运行模式 sse,stdio
}
type ServerMode string

const (
	ServerModeSSE   ServerMode = "sse"
	ServerModeStdio ServerMode = "stdio"
)

// RunMCPServer 启动一个 MCP 服务器实例，支持通过 stdio 和 SSE 两种方式对外提供服务。
func RunMCPServer(name, version string, port int) {
	config := &ServerConfig{}
	config.Name = name
	config.Version = version
	config.Port = port
	// 创建一个新的 MCP 服务器
	s := GetMCPServerWithOption(config)
	// Start the stdio server
	if err := server.ServeStdio(s); err != nil {
		klog.Errorf("stdio server start error: %v\n", err)
	}

	// 创建 SSE 服务器
	sseServer := server.NewSSEServer(s)

	// 启动服务器
	err := sseServer.Start(fmt.Sprintf(":%d", port))
	if err != nil {
		klog.Errorf("MCP Server error: %v\n", err)
	}
}

// RunMCPServerWithOption 根据提供的配置启动 MCP 服务器，可选择 SSE 或 stdio 模式。
//
// 根据 cfg.Mode 字段决定服务器运行模式：
// - 若为 ServerModeStdio，则以 stdio 方式启动 MCP 服务器。
// - 否则，以 SSE 方式启动服务器并监听指定端口。
// 启动前会设置全局认证密钥。
func RunMCPServerWithOption(cfg *ServerConfig) {
	s := GetMCPServerWithOption(cfg)
	tools.SetAuthKey(cfg.AuthKey)
	if cfg.Mode == ServerModeStdio {
		// Start the stdio server
		if err := server.ServeStdio(s); err != nil {
			klog.Errorf("stdio server start error: %v\n", err)
		}
	} else {

		// 创建 SSE 服务器
		sseServer := server.NewSSEServer(s, cfg.SSEOption...)

		// 启动服务器
		err := sseServer.Start(fmt.Sprintf(":%d", cfg.Port))
		if err != nil {
			klog.Errorf("MCP Server error: %v\n", err)
		}
	}

}

// GetMCPSSEServerWithOption 根据提供的配置创建并返回一个包裹了 MCP 服务器的 SSE 服务器实例。
func GetMCPSSEServerWithOption(cfg *ServerConfig) *server.SSEServer {
	s := GetMCPServerWithOption(cfg)
	// 创建 SSE 服务器
	sseServer := server.NewSSEServer(s, cfg.SSEOption...)
	return sseServer
}
// GetMCPServerWithOption 根据提供的配置创建并返回一个注册了所有工具模块的 MCP 服务器实例。
// 如果配置为 nil，则返回 nil。
func GetMCPServerWithOption(cfg *ServerConfig) *server.MCPServer {
	if cfg == nil {
		klog.Errorf("MCP Server error: config is nil\n")
		return nil
	}

	// 创建一个新的 MCP 服务器
	s := server.NewMCPServer(
		cfg.Name,
		cfg.Version,
		cfg.ServerOptions...,
	)

	// 注册工具
	dynamic.RegisterTools(s)
	pod.RegisterTools(s)
	cluster.RegisterTools(s)
	event.RegisterTools(s)
	deployment.RegisterTools(s)
	node.RegisterTools(s)
	storageclass.RegisterTools(s)
	ingressclass.RegisterTools(s)
	yaml.RegisterTools(s)
	ns.RegisterTools(s)
	return s

}
