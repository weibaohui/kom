package yaml

import (
	"github.com/mark3labs/mcp-go/server"
)

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
