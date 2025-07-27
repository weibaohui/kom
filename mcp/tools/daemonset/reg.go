package daemonset

import (
	"github.com/mark3labs/mcp-go/server"
)

func RegisterTools(s *server.MCPServer) {
	s.AddTool(
		RestartDaemonSetTool(),
		RestartDaemonSetHandler,
	)
}
