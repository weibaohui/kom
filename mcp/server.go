package mcp

import (
	"fmt"

	"github.com/mark3labs/mcp-go/server"
	"github.com/weibaohui/kom/mcp/metadata"
	"github.com/weibaohui/kom/mcp/tools/cluster"
	"github.com/weibaohui/kom/mcp/tools/deployment"
	"github.com/weibaohui/kom/mcp/tools/dynamic"
	"github.com/weibaohui/kom/mcp/tools/event"
	"github.com/weibaohui/kom/mcp/tools/ingressclass"
	"github.com/weibaohui/kom/mcp/tools/node"
	"github.com/weibaohui/kom/mcp/tools/pod"
	"github.com/weibaohui/kom/mcp/tools/storageclass"
	"github.com/weibaohui/kom/mcp/tools/yaml"
	"k8s.io/klog/v2"
)

func RunMCPServer(name, version string, port int) {
	config.Name = name
	config.Version = version
	config.Port = port
	// 创建一个新的 MCP 服务器
	s := server.NewMCPServer(
		name,
		version,
		server.WithResourceCapabilities(false, false),
		server.WithPromptCapabilities(false),
		server.WithLogging(),
	)

	// 注册工具
	dynamic.RegisterTools(s, config)
	pod.RegisterTools(s, config)
	cluster.RegisterTools(s, config)
	event.RegisterTools(s, config)
	deployment.RegisterTools(s, config)
	node.RegisterTools(s, config)
	storageclass.RegisterTools(s, config)
	ingressclass.RegisterTools(s, config)
	yaml.RegisterTools(s, config)

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

var config *metadata.ServerConfig

func GetServerConfig() *metadata.ServerConfig {
	return config
}
func RunMCPServerWithOption(cfg *metadata.ServerConfig) {
	s := GetMCPServerWithOption(cfg)
	if cfg.Mode == metadata.MCPServerModeStdio || cfg.Mode == metadata.MCPServerModeBoth {
		// Start the stdio server
		if err := server.ServeStdio(s); err != nil {
			klog.Errorf("stdio server start error: %v\n", err)
		}
	}

	// 创建 SSE 服务器
	sseServer := server.NewSSEServer(s, config.SSEOption...)

	// 启动服务器
	err := sseServer.Start(fmt.Sprintf(":%d", config.Port))
	if err != nil {
		klog.Errorf("MCP Server error: %v\n", err)
	}

}
func GetMCPServerWithOption(cfg *metadata.ServerConfig) *server.MCPServer {
	if cfg == nil {
		klog.Errorf("MCP Server error: config is nil\n")
		return nil
	}
	config = cfg
	// 创建一个新的 MCP 服务器
	s := server.NewMCPServer(
		config.Name,
		config.Version,
		config.ServerOptions...,
	)

	// 注册工具
	dynamic.RegisterTools(s, config)
	pod.RegisterTools(s, config)
	cluster.RegisterTools(s, config)
	event.RegisterTools(s, config)
	deployment.RegisterTools(s, config)
	node.RegisterTools(s, config)
	storageclass.RegisterTools(s, config)
	ingressclass.RegisterTools(s, config)
	yaml.RegisterTools(s, config)
	return s

}
