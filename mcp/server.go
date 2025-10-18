package mcp

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mark3labs/mcp-go/server"
	"github.com/weibaohui/kom/kom"
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
	Kubeconfigs   []KubeconfigConfig // 多集群kubeconfig配置
}

// KubeconfigConfig 定义了单个集群的kubeconfig配置
type KubeconfigConfig struct {
	ID       string // 集群ID，用于标识集群
	Path     string // kubeconfig文件路径
	Content  string // kubeconfig内容（与Path二选一）
	IsDefault bool  // 是否为默认集群
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

// LoadKubeconfigsFromDirectory 从目录中加载所有kubeconfig文件
// 参数:
//   - dir: kubeconfig文件所在目录
//   - pattern: 文件匹配模式，默认为"*.yaml"和"*.yml"
//
// 返回:
//   - []KubeconfigConfig: kubeconfig配置列表
//   - error: 加载失败时返回错误信息
//
// 该函数会扫描指定目录，为每个kubeconfig文件创建配置
func LoadKubeconfigsFromDirectory(dir string, pattern ...string) ([]KubeconfigConfig, error) {
	if dir == "" {
		return nil, fmt.Errorf("directory path cannot be empty")
	}

	// 默认匹配模式
	patterns := []string{"*.yaml", "*.yml"}
	if len(pattern) > 0 {
		patterns = pattern
	}

	var configs []KubeconfigConfig

	for _, p := range patterns {
		matches, err := filepath.Glob(filepath.Join(dir, p))
		if err != nil {
			return nil, fmt.Errorf("failed to glob pattern %s: %w", p, err)
		}

		for _, match := range matches {
			// 使用文件名（不含扩展名）作为集群ID
			clusterID := filepath.Base(match)
			ext := filepath.Ext(clusterID)
			if ext != "" {
				clusterID = clusterID[:len(clusterID)-len(ext)]
			}

			configs = append(configs, KubeconfigConfig{
				ID:   clusterID,
				Path: match,
			})
		}
	}

	return configs, nil
}

// initializeMultiCluster 初始化多集群配置
// 参数:
//   - cfg: 服务器配置参数
//
// 返回:
//   - error: 初始化失败时返回错误信息
//
// 该函数会根据配置的Kubeconfigs初始化多个Kubernetes集群连接
func initializeMultiCluster(cfg *ServerConfig) error {
	if cfg == nil || len(cfg.Kubeconfigs) == 0 {
		klog.V(2).Infof("No kubeconfig configurations provided, skipping multi-cluster initialization")
		return nil
	}

	klog.Infof("Initializing %d clusters from kubeconfig configurations", len(cfg.Kubeconfigs))

	for _, kubeconfig := range cfg.Kubeconfigs {
		if kubeconfig.ID == "" {
			return fmt.Errorf("cluster ID cannot be empty")
		}

		var err error
		if kubeconfig.Content != "" {
			// 使用kubeconfig内容
			_, err = kom.Clusters().RegisterByStringWithID(kubeconfig.Content, kubeconfig.ID)
			klog.V(2).Infof("Registered cluster %s from kubeconfig content", kubeconfig.ID)
		} else if kubeconfig.Path != "" {
			// 使用kubeconfig文件路径
			// 检查文件是否存在
			if _, err := os.Stat(kubeconfig.Path); os.IsNotExist(err) {
				return fmt.Errorf("kubeconfig file not found: %s", kubeconfig.Path)
			}
			_, err = kom.Clusters().RegisterByPathWithID(kubeconfig.Path, kubeconfig.ID)
			klog.V(2).Infof("Registered cluster %s from kubeconfig file: %s", kubeconfig.ID, kubeconfig.Path)
		} else {
			return fmt.Errorf("either kubeconfig content or path must be provided for cluster %s", kubeconfig.ID)
		}

		if err != nil {
			return fmt.Errorf("failed to register cluster %s: %w", kubeconfig.ID, err)
		}

		// 如果指定为默认集群，将其设置为默认集群
		if kubeconfig.IsDefault {
			// 这里可以通过修改DefaultCluster方法来实现，或者通过其他方式标记默认集群
			klog.V(2).Infof("Cluster %s marked as default", kubeconfig.ID)
		}
	}

	// 显示所有已注册的集群
	kom.Clusters().Show()
	return nil
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

	// 初始化多集群配置
	if err := initializeMultiCluster(cfg); err != nil {
		klog.Errorf("Failed to initialize multi-cluster configuration: %v", err)
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
