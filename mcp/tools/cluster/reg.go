package cluster

import (
	"github.com/mark3labs/mcp-go/server"
)

func RegisterTools(s *server.MCPServer) {

	s.AddTool(
		ListClusters(),
		ListClustersHandler,
	)

	s.AddTool(
		RegisterCluster(),
		RegisterClusterHandler,
	)

	s.AddTool(
		UnregisterCluster(),
		UnregisterClusterHandler,
	)
}
