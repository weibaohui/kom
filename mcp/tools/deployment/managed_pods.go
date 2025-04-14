package deployment

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/weibaohui/kom/kom"
	"github.com/weibaohui/kom/mcp/metadata"
	"github.com/weibaohui/kom/utils"
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
	ctx, meta, err := metadata.ParseFromRequest(ctx, request, config)

	if err != nil {
		return nil, err
	}
	// 如果只有一个集群的时候，使用空，默认集群
	// 如果大于一个集群，没有传值，那么要返回错误
	if len(kom.Clusters().AllClusters()) > 1 && meta.Cluster == "" {
		return nil, fmt.Errorf("cluster is required, 集群名称必须设置")
	}
	if kom.Clusters().GetClusterById(meta.Cluster) == nil {
		return nil, fmt.Errorf("cluster %s not found 集群不存在，请检查集群名称", meta.Cluster)
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

	return utils.TextResult(podNames, meta)
}
