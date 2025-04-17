package deployment

import (
	"github.com/mark3labs/mcp-go/server"
	"github.com/weibaohui/kom/mcp/metadata"
)

var config *metadata.ServerConfig

func RegisterTools(s *server.MCPServer, cfg *metadata.ServerConfig) {
	config = cfg

	s.AddTool(
		ScaleDeploymentTool(),
		ScaleDeploymentHandler,
	)
	s.AddTool(
		RestartDeploymentTool(),
		RestartDeploymentHandler,
	)
	s.AddTool(
		StopDeploymentTool(),
		StopDeploymentHandler,
	)
	s.AddTool(
		RestoreDeploymentTool(),
		RestoreDeploymentHandler,
	)
	s.AddTool(
		UpdateTagDeploymentTool(),
		UpdateTagDeploymentHandler,
	)
	s.AddTool(
		RolloutHistoryDeploymentTool(),
		RolloutHistoryDeploymentHandler,
	)
	s.AddTool(
		RolloutUndoDeploymentTool(),
		RolloutUndoDeploymentHandler,
	)
	s.AddTool(
		RolloutPauseDeploymentTool(),
		RolloutPauseDeploymentHandler,
	)
	s.AddTool(
		RolloutResumeDeploymentTool(),
		RolloutResumeDeploymentHandler,
	)
	s.AddTool(
		RolloutStatusDeploymentTool(),
		RolloutStatusDeploymentHandler,
	)
	s.AddTool(
		HPAListDeploymentTool(),
		HPAListDeploymentHandler,
	)
}
