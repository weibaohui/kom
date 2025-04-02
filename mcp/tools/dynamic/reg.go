package dynamic

import (
	"github.com/mark3labs/mcp-go/server"
	"github.com/weibaohui/kom/mcp/metadata"
)

var config *metadata.ServerConfig

func RegisterTools(s *server.MCPServer, cfg *metadata.ServerConfig) {
	config = cfg
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
	s.AddTool(
		PatchDynamicResource(),
		PatchDynamicResourceHandler)
}
