package dynamic

import (
	"github.com/mark3labs/mcp-go/server"
)

func RegisterTools(s *server.MCPServer) {
	s.AddTool(
		GetDynamicResource(),
		GetDynamicResourceHandler,
	)
	s.AddTool(
		GetDynamicResourceDescribe(),
		GetDynamicResourceDescribeHandler)
}
