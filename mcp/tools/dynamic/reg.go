package dynamic

import (
	"github.com/mark3labs/mcp-go/server"
)

// RegisterTools 向指定的 MCPServer 实例批量注册动态资源相关的工具及其处理函数。
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
	s.AddTool(
		PatchDynamicResource(),
		PatchDynamicResourceHandler)
	 
}
