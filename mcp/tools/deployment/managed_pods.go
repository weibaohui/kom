package deployment

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/weibaohui/kom/kom"
	"github.com/weibaohui/kom/mcp/tools"
	"github.com/weibaohui/kom/mcp/tools/metadata"
)

// ManagedPodsDeploymentTool 创建一个获取Deployment管理的Pod列表的工具
func ManagedPodsDeploymentTool() mcp.Tool {
	return mcp.NewTool(
		"list_deployment_pods",
		mcp.WithDescription("获取Deployment管理的Pod列表，通过集群、命名空间和名称 / Get managed pods of deployment by cluster, namespace and name"),
		mcp.WithString("cluster", mcp.Description("运行Deployment的集群 / The cluster runs the deployment")),
		mcp.WithString("namespace", mcp.Description("Deployment所在的命名空间 / The namespace of the deployment")),
		mcp.WithString("name", mcp.Description("Deployment的名称 / The name of the deployment")),
	)
}

// ManagedPodsDeploymentHandler 处理获取Deployment管理的Pod列表的请求
func ManagedPodsDeploymentHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 获取参数
	meta, err := metadata.ParseFromRequest(request)
	if err != nil {
		return nil, err
	}

	list, err := kom.Cluster(meta.Cluster).WithContext(ctx).Namespace(meta.Namespace).Name(meta.Name).Ctl().Deployment().ManagedPods()
	if err != nil {
		return nil, err
	}

	// 构建Pod列表信息
	var podNames []string
	for _, pod := range list {
		podNames = append(podNames, pod.Name)
	}

	return tools.TextResult(podNames, meta)
}
