package deployment

import (
	"github.com/mark3labs/mcp-go/server"
)

func RegisterTools(s *server.MCPServer) {
	s.AddTool(
		ScaleDeploymentTool(),
		ScaleDeploymentHandler,
	)
	s.AddTool(
		RestartDeploymentTool(),
		RestartDeploymentHandler)
}
