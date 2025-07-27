package mcp

import (
	"fmt"

	"github.com/mark3labs/mcp-go/server"
	"github.com/weibaohui/kom/mcp/tools"
	"github.com/weibaohui/kom/mcp/tools/cluster"
	"github.com/weibaohui/kom/mcp/tools/daemonset"
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

// ServerConfig 定义了MCP服务器的配置参数
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

// ServerMode 定义了服务器的运行模式类型
type ServerMode string

// 定义服务器运行模式常量
const (
	ServerModeSSE   ServerMode = "sse"   // SSE模式
	ServerModeStdio ServerMode = "stdio" // 标准输入输出模式
)

// RunMCPServer 启动一个基本的MCP服务器
// 参数:
//   - name: 服务器名称
//   - version: 服务器版本
//   - port: 服务器监听端口
//
// 该函数会同时启动stdio服务器和SSE服务器
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

// RunMCPServerWithOption 使用自定义配置启动MCP服务器
// 参数:
//   - cfg: 服务器配置参数
//
// 根据配置的Mode决定启动stdio服务器还是SSE服务器
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

// GetMCPSSEServerWithOption 创建并返回一个SSE服务器实例
// 参数:
//   - cfg: 服务器配置参数
//
// 返回:
//   - *server.SSEServer: 配置完成的SSE服务器实例
func GetMCPSSEServerWithOption(cfg *ServerConfig) *server.SSEServer {
	s := GetMCPServerWithOption(cfg)
	tools.SetAuthKey(cfg.AuthKey)

	// 创建 SSE 服务器
	sseServer := server.NewSSEServer(s, cfg.SSEOption...)
	return sseServer
}

// GetMCPSSEServerWithServerAndOption 使用现有的MCP服务器实例创建SSE服务器
// 参数:
//   - s: 现有的MCP服务器实例
//   - cfg: 服务器配置参数
//
// 返回:
//   - *server.SSEServer: 配置完成的SSE服务器实例
func GetMCPSSEServerWithServerAndOption(s *server.MCPServer, cfg *ServerConfig) *server.SSEServer {
	tools.SetAuthKey(cfg.AuthKey)
	// 创建 SSE 服务器
	sseServer := server.NewSSEServer(s, cfg.SSEOption...)
	return sseServer
}

// GetMCPServerWithOption 创建并配置一个新的MCP服务器实例
// 参数:
//   - cfg: 服务器配置参数
//
// 返回:
//   - *server.MCPServer: 配置完成的MCP服务器实例，如果cfg为nil则返回nil
//
// 该函数会注册所有可用的工具到服务器实例
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
	daemonset.RegisterTools(s)
	return s

}
