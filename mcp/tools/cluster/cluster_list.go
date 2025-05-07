package cluster

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/weibaohui/kom/kom"
	"github.com/weibaohui/kom/utils"
)

// ListClusters 返回一个用于列出所有已注册 Kubernetes 集群的工具实例。
func ListClusters() mcp.Tool {
	return mcp.NewTool(
		"list_k8s_clusters",
		mcp.WithDescription("列出所有已注册的Kubernetes集群（可用集群、可操作集群） / List all registered Kubernetes clusters"),
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

	return utils.TextResult(result, nil)
}
