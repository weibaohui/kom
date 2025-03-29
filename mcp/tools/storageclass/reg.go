package storageclass

import (
	"github.com/mark3labs/mcp-go/server"
)

func RegisterTools(s *server.MCPServer) {
	s.AddTool(
		SetDefaultStorageClassTool(),
		SetDefaultStorageClassHandler,
	)
	s.AddTool(
		GetStorageClassPVCCountTool(),
		GetStorageClassPVCCountHandler,
	)
	s.AddTool(
		GetStorageClassPVCountTool(),
		GetStorageClassPVCountHandler,
	)
}
