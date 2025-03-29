package pod

import (
	"github.com/mark3labs/mcp-go/server"
)

// RegisterTools 注册Pod相关的工具到MCP服务器
func RegisterTools(s *server.MCPServer) {
	s.AddTool(
		ListPodFilesTool(),
		ListPodFilesHandler,
	)
	s.AddTool(
		ListAllPodFilesTool(),
		ListAllPodFilesHandler,
	)
	s.AddTool(
		DeletePodFileTool(),
		DeletePodFileHandler,
	)

	s.AddTool(
		GetPodLogsTool(),
		GetPodLogsHandler,
	)

	s.AddTool(
		GetPodLinkedServiceTool(),
		GetPodLinkedServiceHandler,
	)
	s.AddTool(
		GetPodLinkedIngressTool(),
		GetPodLinkedIngressHandler,
	)
	s.AddTool(
		GetPodLinkedEndpointsTool(),
		GetPodLinkedEndpointsHandler,
	)
	s.AddTool(
		GetPodLinkedPVCTool(),
		GetPodLinkedPVCHandler,
	)
	s.AddTool(
		GetPodLinkedPVTool(),
		GetPodLinkedPVHandler,
	)
	s.AddTool(
		GetPodLinkedEnvTool(),
		GetPodLinkedEnvHandler,
	)
	s.AddTool(
		PodExecTool(),
		PodExecHandler,
	)
	s.AddTool(
		GetPodResourceUsageTool(),
		GetPodResourceUsageHandler,
	)
}
