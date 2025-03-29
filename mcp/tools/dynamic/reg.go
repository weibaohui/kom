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
	s.AddTool(
		DeleteDynamicResource(),
		DeleteDynamicResourceHandler)
	s.AddTool(
		ListDynamicResource(),
		ListDynamicResourceHandler)
	s.AddTool(
		AnnotateDynamicResource(),
		AnnotateDynamicResourceHandler)
	s.AddTool(
		LabelDynamicResource(),
		LabelDynamicResourceHandler)
}
