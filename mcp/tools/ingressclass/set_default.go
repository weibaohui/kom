package ingressclass

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/weibaohui/kom/kom"
	"github.com/weibaohui/kom/mcp/tools"
	v1 "k8s.io/api/networking/v1"
)

// SetDefaultIngressClassTool 返回一个用于将指定 Kubernetes 集群中的 IngressClass 设置为默认的 MCP 工具。
func SetDefaultIngressClassTool() mcp.Tool {
	return mcp.NewTool(
		"set_default_k8s_ingressclass",
		mcp.WithDescription("设置IngressClass为默认 / Set IngressClass as default (kubectl annotate ingressclass <name> ingressclass.kubernetes.io/is-default-class=true)"),
		mcp.WithString("cluster", mcp.Description("IngressClass所在的集群 （使用空字符串表示默认集群）/ The cluster of the IngressClass")),
		mcp.WithString("name", mcp.Description("IngressClass的名称 / The name of the IngressClass")),
	)
}

// SetDefaultIngressClassHandler 处理设置指定集群中 IngressClass 为默认的请求。
// 解析请求参数后，将指定名称的 IngressClass 标记为默认，并返回操作结果。
// 若参数解析或设置过程中发生错误，则返回相应错误。
func SetDefaultIngressClassHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 获取参数
	ctx, meta, err := tools.ParseFromRequest(ctx, request)
	if err != nil {
		return nil, err
	}

	err = kom.Cluster(meta.Cluster).WithContext(ctx).Resource(&v1.IngressClass{}).Name(meta.Name).Ctl().IngressClass().SetDefault()
	if err != nil {
		return nil, err
	}

	return tools.TextResult("Successfully set IngressClass as default", meta)
}
