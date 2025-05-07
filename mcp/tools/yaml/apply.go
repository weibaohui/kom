package yaml

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/weibaohui/kom/kom"
	"github.com/weibaohui/kom/mcp/tools"
)

// ApplyDynamicResource 返回一个用于通过YAML内容在指定Kubernetes集群中创建或更新资源的MCP工具。
func ApplyDynamicResource() mcp.Tool {
	return mcp.NewTool(
		"apply_k8s_yaml",
		mcp.WithDescription("通过YAML创建或更新Kubernetes资源，等同于 'kubectl apply -f <yaml-file>' / Apply Kubernetes resources from YAML, equivalent to 'kubectl apply -f <yaml-file>'"),
		mcp.WithString("yaml", mcp.Description("需要应用的YAML内容 / YAML content containing resources to apply")),
		mcp.WithString("cluster", mcp.Description("目标集群（空值表示默认集群） / Target cluster (empty for default)")),
	)
}

// ApplyDynamicResourceHandler 处理“apply_k8s_yaml”工具的请求，将 YAML 内容中的 Kubernetes 资源应用到指定集群。
// 如果请求参数无效或 YAML 内容缺失，则返回错误。
// 返回应用结果的文本格式。
func ApplyDynamicResourceHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ctx, meta, err := tools.ParseFromRequest(ctx, request)

	if err != nil {
		return nil, err
	}

	yamlContent, ok := request.Params.Arguments["yaml"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid yaml content")
	}

	results := kom.Cluster(meta.Cluster).WithContext(ctx).Applier().Apply(yamlContent)
	return tools.TextResult(results, meta)
}
