package yaml

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/weibaohui/kom/kom"
	"github.com/weibaohui/kom/mcp/tools"
)

// DeleteDynamicResource 返回一个用于根据YAML内容删除Kubernetes资源的MCP工具，相当于执行 'kubectl delete -f <yaml-file>'。
func DeleteDynamicResource() mcp.Tool {
	return mcp.NewTool(
		"delete_k8s_yaml",
		mcp.WithDescription("通过YAML删除Kubernetes资源，等同于 'kubectl delete -f <yaml-file>' / Delete Kubernetes resources from YAML, equivalent to 'kubectl delete -f <yaml-file>'"),
		mcp.WithString("yaml", mcp.Description("需要删除的YAML内容 / YAML content containing resources to delete")),
		mcp.WithString("cluster", mcp.Description("目标集群（空值表示默认集群）/ Target cluster (empty for default)")),
	)
}

// DeleteDynamicResourceHandler 根据请求参数删除指定 Kubernetes 集群中的 YAML 资源，并返回操作结果。
func DeleteDynamicResourceHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ctx, meta, err := tools.ParseFromRequest(ctx, request)
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
