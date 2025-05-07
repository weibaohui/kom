package ingressclass

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/weibaohui/kom/kom"
	"github.com/weibaohui/kom/mcp/tools"
	v1 "k8s.io/api/networking/v1"
)

func SetDefaultIngressClassTool() mcp.Tool {
	return mcp.NewTool(
		"set_default_k8s_ingressclass",
		mcp.WithDescription("设置IngressClass为默认 / Set IngressClass as default (kubectl annotate ingressclass <name> ingressclass.kubernetes.io/is-default-class=true)"),
		mcp.WithString("cluster", mcp.Description("IngressClass所在的集群 （使用空字符串表示默认集群）/ The cluster of the IngressClass")),
		mcp.WithString("name", mcp.Description("IngressClass的名称 / The name of the IngressClass")),
	)
}

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
