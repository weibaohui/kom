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

func GetMCPSSEServerWithOption(cfg *ServerConfig) *server.SSEServer {
	s := GetMCPServerWithOption(cfg)
	// 创建 SSE 服务器
	sseServer := server.NewSSEServer(s, cfg.SSEOption...)
	return sseServer
}
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
