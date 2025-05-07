package node

import (
	"github.com/mark3labs/mcp-go/server"
)

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
}
