package node

import (
	"github.com/mark3labs/mcp-go/server"
)

// RegisterTools 向指定的 MCPServer 实例注册所有节点相关的工具及其处理函数。
func RegisterTools(s *server.MCPServer) {

	s.AddTool(
		TaintNodeTool(),
		TaintNodeHandler,
	)
	s.AddTool(
		UnTaintNodeTool(),
		UnTaintNodeHandler,
	)
	s.AddTool(
		CordonNodeTool(),
		CordonNodeHandler,
	)
	s.AddTool(
		UnCordonNodeTool(),
		UnCordonNodeHandler,
	)

	s.AddTool(
		DrainNodeTool(),
		DrainNodeHandler,
	)
	s.AddTool(
		NodeResourceUsageTool(),
		NodeResourceUsageHandler,
	)

	s.AddTool(
		NodeIPUsageTool(),
		NodeIPUsageHandler,
	)
	s.AddTool(
		NodePodCountTool(),
		NodePodCountHandler,
	)
	s.AddTool(
		ListNode(),
		ListNodeHandler,
	)
	s.AddTool(
		TopNode(),
		TopNodeHandler,
	)
}
