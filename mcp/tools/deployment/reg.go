package deployment

import (
	"github.com/mark3labs/mcp-go/server"
)

// RegisterTools 向指定的 MCPServer 注册所有部署相关的工具及其处理函数。
func RegisterTools(s *server.MCPServer) {

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

	s.AddTool(
		ListDeployEventResource(),
		ListDeployEventResourceHandler,
	)
}
