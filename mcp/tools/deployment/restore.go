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

// RestoreDeploymentTool 创建一个用于根据注解恢复 Kubernetes Deployment 副本数的工具。
// 如果未找到原始副本数注解，则副本数恢复为 1。该工具要求指定集群（cluster），可选命名空间（namespace）和 Deployment 名称（name）。
// 相当于执行 kubectl scale deployment/<name> --replicas=<original_replicas> -n <namespace>。
func RestoreDeploymentTool() mcp.Tool {
	return mcp.NewTool(
		"restore_k8s_deployment",
		mcp.WithDescription("恢复Deployment副本数，（从注解中恢复原始副本数，如果没有注解则默认为1）。对应kubectl命令: kubectl scale deployment/<name> --replicas=<original_replicas> -n <namespace> / Restore deployment replicas from annotation, default to 1 if not found. Equivalent kubectl command: kubectl scale deployment/<name> --replicas=<original_replicas> -n <namespace>"),
		mcp.WithString("cluster", mcp.Required(), mcp.Description("运行Deployment的集群 （使用空字符串表示默认集群）/ The cluster runs the deployment")),
		mcp.WithString("namespace", mcp.Description("Deployment所在的命名空间 / The namespace of the deployment")),
		mcp.WithString("name", mcp.Description("Deployment的名称 / The name of the deployment")),
	)
}

// RestoreDeploymentHandler 处理恢复Deployment的请求
func RestoreDeploymentHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

	klog.Infof("Restoring deployment %s/%s in cluster %s", meta.Namespace, meta.Name, meta.Cluster)

	err = kom.Cluster(meta.Cluster).WithContext(ctx).Resource(&appsv1.Deployment{}).Namespace(meta.Namespace).Name(meta.Name).Ctl().Scaler().Restore()
	if err != nil {
		return nil, err
	}

	return utils.TextResult("Successfully restored deployment", meta)
}
