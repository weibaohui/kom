package cluster

import (
	"github.com/mark3labs/mcp-go/server"
)

// RegisterTools 向服务器注册集群相关的工具及其处理函数。
func RegisterTools(s *server.MCPServer) {

	s.AddTool(
		ListClusters(),
		ListClustersHandler,
	)
}
