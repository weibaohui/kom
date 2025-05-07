package ingressclass

import (
	"github.com/mark3labs/mcp-go/server"
)

// RegisterTools 向 MCP 服务器注册默认的 IngressClass 工具及其处理器。
func RegisterTools(s *server.MCPServer) {

	s.AddTool(
		SetDefaultIngressClassTool(),
		SetDefaultIngressClassHandler,
	)
}
