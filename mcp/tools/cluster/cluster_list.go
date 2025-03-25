package cluster

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/weibaohui/kom/kom"
	"github.com/weibaohui/kom/mcp/tools"
)

func ListClusters() mcp.Tool {
	return mcp.NewTool(
		"list_clusters",
		mcp.WithDescription("List all registered Kubernetes clusters / 列出所有已注册的Kubernetes集群"),
	)
}

func ListClustersHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 获取所有已注册的集群名称
	clusters := kom.Clusters().AllClusters()

	// 提取集群名称
	var result []map[string]string
	for clusterName, _ := range clusters {
		result = append(result, map[string]string{
			"name": clusterName,
		})
	}

	return tools.TextResult(result, nil)
}
