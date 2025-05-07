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

// StopDeploymentTool 创建一个用于停止 Kubernetes Deployment 的工具。
// 该工具通过将 Deployment 的副本数设置为 0 来停止其运行，并将原始副本数记录到注解中，便于后续恢复。
// 工具参数包括集群（必填，空字符串表示默认集群）、命名空间和 Deployment 名称。
func StopDeploymentTool() mcp.Tool {
	return mcp.NewTool(
		"stop_k8s_deployment",
		mcp.WithDescription("停止Deployment。（将副本数设置为0并记录原始副本数到注解中，恢复是可使用restore_deployment方法）。对应kubectl命令: kubectl scale deployment/<name> --replicas=0 -n <namespace> / Stop deployment by setting replicas to 0 and save original replicas to annotation. Equivalent kubectl command: kubectl scale deployment/<name> --replicas=0 -n <namespace>"),
		mcp.WithString("cluster", mcp.Required(), mcp.Description("运行Deployment的集群 （使用空字符串表示默认集群）/ The cluster runs the deployment")),
		mcp.WithString("namespace", mcp.Description("Deployment所在的命名空间 / The namespace of the deployment")),
		mcp.WithString("name", mcp.Description("Deployment的名称 / The name of the deployment")),
	)
}

// StopDeploymentHandler 处理停止Deployment的请求
func StopDeploymentHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

	klog.Infof("Stopping deployment %s/%s in cluster %s", meta.Namespace, meta.Name, meta.Cluster)

	err = kom.Cluster(meta.Cluster).WithContext(ctx).Resource(&appsv1.Deployment{}).Namespace(meta.Namespace).Name(meta.Name).Ctl().Scaler().Stop()
	if err != nil {
		return nil, err
	}

	return utils.TextResult("Successfully stopped deployment", meta)
}
