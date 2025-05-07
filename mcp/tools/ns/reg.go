package ns

import (
	"github.com/mark3labs/mcp-go/server"
)

// RegisterTools 向 MCP 服务器注册命名空间相关的工具及其处理函数。
func RegisterTools(s *server.MCPServer) {

	s.AddTool(
		ListNamespace(),
		ListNamespaceHandler)

}
