package yaml

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/weibaohui/kom/kom"
	"github.com/weibaohui/kom/mcp/tools"
)

func ApplyDynamicResource() mcp.Tool {
	return mcp.NewTool(
		"apply_k8s_yaml",
		mcp.WithDescription("通过YAML创建或更新Kubernetes资源，等同于 'kubectl apply -f <yaml-file>' / Apply Kubernetes resources from YAML, equivalent to 'kubectl apply -f <yaml-file>'"),
		mcp.WithTitleAnnotation("Apply YAML"),
		mcp.WithDestructiveHintAnnotation(true),
		mcp.WithString("yaml", mcp.Description("需要应用的YAML内容 / YAML content containing resources to apply")),
		mcp.WithString("cluster", mcp.Description("目标集群（空值表示默认集群） / Target cluster (empty for default)")),
	)
}

func ApplyDynamicResourceHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ctx, meta, err := tools.ParseFromRequest(ctx, request)

	if err != nil {
		return nil, err
	}

	yamlContent := request.GetString("yaml", "")
	if yamlContent == "" {
		return nil, fmt.Errorf("invalid yaml content")
	}

	results := kom.Cluster(meta.Cluster).WithContext(ctx).Applier().Apply(yamlContent)
	return tools.TextResult(results, meta)
}
