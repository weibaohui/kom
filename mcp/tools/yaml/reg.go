package yaml

import (
	"github.com/mark3labs/mcp-go/server"
)

// RegisterTools 向指定的 MCPServer 注册动态资源的应用和删除工具。
func RegisterTools(s *server.MCPServer) {

	s.AddTool(
		ApplyDynamicResource(),
		ApplyDynamicResourceHandler,
	)
	s.AddTool(
		DeleteDynamicResource(),
		DeleteDynamicResourceHandler,
	)
}
