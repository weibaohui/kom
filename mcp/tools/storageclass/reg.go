package storageclass

import (
	"github.com/mark3labs/mcp-go/server"
)

// RegisterTools 向指定的 MCPServer 实例注册存储类相关的工具及其处理器。
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
