package yaml

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/weibaohui/kom/kom"
	"github.com/weibaohui/kom/mcp/tools"
	"github.com/weibaohui/kom/mcp/tools/metadata"
)

func DeleteDynamicResource() mcp.Tool {
	return mcp.NewTool(
		"delete_yaml",
		mcp.WithDescription("Delete Kubernetes resources from YAML / 通过YAML删除Kubernetes资源"),
		mcp.WithString("yaml", mcp.Description("YAML content containing resources to delete / 需要删除的YAML内容")),
		mcp.WithString("cluster", mcp.Description("Target cluster (empty for default) / 目标集群（空值表示默认集群）")),
	)
}

func DeleteDynamicResourceHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	meta, err := metadata.ParseFromRequest(request)
	if err != nil {
		return nil, err
	}

	yamlContent, ok := request.Params.Arguments["yaml"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid yaml content")
	}

	results := kom.Cluster(meta.Cluster).WithContext(ctx).Applier().Delete(yamlContent)
	return tools.TextResult(results, meta)
}
