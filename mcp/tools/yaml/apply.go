package yaml

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/weibaohui/kom/kom"
	"github.com/weibaohui/kom/mcp/metadata"
	"github.com/weibaohui/kom/utils"
)

func ApplyDynamicResource() mcp.Tool {
	return mcp.NewTool(
		"apply_yaml",
		mcp.WithDescription("通过YAML创建或更新Kubernetes资源，等同于 'kubectl apply -f <yaml-file>' / Apply Kubernetes resources from YAML, equivalent to 'kubectl apply -f <yaml-file>'"),
		mcp.WithString("yaml", mcp.Description("需要应用的YAML内容 / YAML content containing resources to apply")),
		mcp.WithString("cluster", mcp.Description("目标集群（空值表示默认集群） / Target cluster (empty for default)")),
	)
}

func ApplyDynamicResourceHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ctx, meta, err := metadata.ParseFromRequest(ctx, request, config)

	if err != nil {
		return nil, err
	}
	// 如果只有一个集群的时候，使用空，默认集群
	// 如果大于一个集群，没有传值，那么要返回错误
	if len(kom.Clusters().AllClusters()) > 1 && meta.Cluster == "" {
		return nil, fmt.Errorf("cluster is required, 集群名称必须设置")
	}
	if len(kom.Clusters().AllClusters()) == 1 && meta.Cluster == "" {
		meta.Cluster = kom.Clusters().DefaultCluster().ID
	}
	if kom.Clusters().GetClusterById(meta.Cluster) == nil {
		return nil, fmt.Errorf("cluster %s not found 集群不存在，请检查集群名称", meta.Cluster)
	}

	yamlContent, ok := request.Params.Arguments["yaml"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid yaml content")
	}

	results := kom.Cluster(meta.Cluster).WithContext(ctx).Applier().Apply(yamlContent)
	return utils.TextResult(results, meta)
}
