package deployment

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/weibaohui/kom/kom"
	"github.com/weibaohui/kom/mcp/metadata"

	"github.com/weibaohui/kom/utils"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/klog/v2"
)

// HPAListDeploymentTool 返回一个用于查询指定 Deployment 关联的 HPA 列表的工具。该工具要求指定集群，并可选指定命名空间和 Deployment 名称。
func HPAListDeploymentTool() mcp.Tool {
	return mcp.NewTool(
		"get_k8s_deployment_hpa_list",
		mcp.WithDescription("查询Deployment的HPA列表。对应kubectl命令: kubectl get hpa -n <namespace> | grep <deployment-name> / Query deployment HPA list. Equivalent kubectl command: kubectl get hpa -n <namespace> | grep <deployment-name>"),
		mcp.WithString("cluster", mcp.Required(), mcp.Description("运行Deployment的集群 / The cluster runs the deployment")),
		mcp.WithString("namespace", mcp.Description("Deployment所在的命名空间 / The namespace of the deployment")),
		mcp.WithString("name", mcp.Description("Deployment的名称 / The name of the deployment")),
	)
}

// HPAListDeploymentHandler 处理查询Deployment的HPA列表的请求
func HPAListDeploymentHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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
	klog.Infof("Getting HPA list for deployment %s/%s in cluster %s", meta.Namespace, meta.Name, meta.Cluster)

	list, err := kom.Cluster(meta.Cluster).WithContext(ctx).Resource(&appsv1.Deployment{}).Namespace(meta.Namespace).Name(meta.Name).Ctl().Deployment().HPAList()
	if err != nil {
		return nil, err
	}

	var result string
	for _, item := range list {
		result += "HPA " + item.Name + "\n"
	}

	return utils.TextResult(result, meta)
}
