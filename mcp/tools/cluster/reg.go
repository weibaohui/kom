package cluster

import (
	"github.com/mark3labs/mcp-go/server"
	"github.com/weibaohui/kom/mcp/metadata"
)

var config *metadata.ServerConfig

func RegisterTools(s *server.MCPServer, cfg *metadata.ServerConfig) {
	config = cfg
	s.AddTool(
		ListClusters(),
		ListClustersHandler,
	)
}
