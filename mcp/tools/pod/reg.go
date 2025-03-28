package pod

import (
	"github.com/mark3labs/mcp-go/server"
)

// RegisterTools 注册Pod相关的工具到MCP服务器
func RegisterTools(s *server.MCPServer) {
	s.AddTool(
		GetPodLogsTool(),
		GetPodLogsHandler,
	)
	s.AddTool(
		FileOperationTool(),
		FileOperationHandler,
	)
}
