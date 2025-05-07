package event

import (
	"github.com/mark3labs/mcp-go/server"
)

// RegisterTools 向指定的 MCPServer 实例注册事件资源及其处理器。
func RegisterTools(s *server.MCPServer) {

	s.AddTool(
		ListEventResource(),
		ListEventResourceHandler,
	)
}
