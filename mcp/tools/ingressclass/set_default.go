package ingressclass

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/weibaohui/kom/kom"
	"github.com/weibaohui/kom/mcp/metadata"
	"github.com/weibaohui/kom/utils"

	v1 "k8s.io/api/networking/v1"
)

// SetDefaultIngressClassTool 返回一个用于将指定 Kubernetes IngressClass 设置为默认的工具。
// 该工具要求提供 IngressClass 所在集群（可为空表示默认集群）和 IngressClass 名称作为参数。
func SetDefaultIngressClassTool() mcp.Tool {
	return mcp.NewTool(
		"set_default_k8s_ingressclass",
		mcp.WithDescription("设置IngressClass为默认 / Set IngressClass as default (kubectl annotate ingressclass <name> ingressclass.kubernetes.io/is-default-class=true)"),
		mcp.WithString("cluster", mcp.Required(), mcp.Description("IngressClass所在的集群 （使用空字符串表示默认集群）/ The cluster of the IngressClass")),
		mcp.WithString("name", mcp.Description("IngressClass的名称 / The name of the IngressClass")),
	)
}

func SetDefaultIngressClassHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 获取参数
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

	err = kom.Cluster(meta.Cluster).WithContext(ctx).Resource(&v1.IngressClass{}).Name(meta.Name).Ctl().IngressClass().SetDefault()
	if err != nil {
		return nil, err
	}

	return utils.TextResult("Successfully set IngressClass as default", meta)
}
