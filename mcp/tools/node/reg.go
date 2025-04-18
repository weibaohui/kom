package node

import (
	"github.com/mark3labs/mcp-go/server"
	"github.com/weibaohui/kom/mcp/metadata"
)

var config *metadata.ServerConfig

func RegisterTools(s *server.MCPServer, cfg *metadata.ServerConfig) {
	config = cfg
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
