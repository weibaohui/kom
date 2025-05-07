package deployment

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/weibaohui/kom/kom"
	"github.com/weibaohui/kom/mcp/tools"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/klog/v2"
)

// RestoreDeploymentTool 返回一个用于恢复 Kubernetes Deployment 副本数的工具。
// 该工具根据 Deployment 注解中的原始副本数进行恢复，若无注解则副本数默认为 1。工具参数包括目标集群（可为空表示默认集群）、命名空间和 Deployment 名称。
func RestoreDeploymentTool() mcp.Tool {
	return mcp.NewTool(
		"restore_k8s_deployment",
		mcp.WithDescription("恢复Deployment副本数，（从注解中恢复原始副本数，如果没有注解则默认为1）。对应kubectl命令: kubectl scale deployment/<name> --replicas=<original_replicas> -n <namespace> / Restore deployment replicas from annotation, default to 1 if not found. Equivalent kubectl command: kubectl scale deployment/<name> --replicas=<original_replicas> -n <namespace>"),
		mcp.WithString("cluster", mcp.Description("运行Deployment的集群 （使用空字符串表示默认集群）/ The cluster runs the deployment")),
		mcp.WithString("namespace", mcp.Description("Deployment所在的命名空间 / The namespace of the deployment")),
		mcp.WithString("name", mcp.Description("Deployment的名称 / The name of the deployment")),
	)
}

// RestoreDeploymentHandler 根据存储在注解中的副本数恢复指定 Kubernetes Deployment 的副本数量。
// 如果注解不存在，则副本数恢复为 1。
// 成功时返回操作结果文本，失败时返回错误。
func RestoreDeploymentHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 获取参数
	ctx, meta, err := tools.ParseFromRequest(ctx, request)
	if err != nil {
		return nil, err
	}

	klog.Infof("Restoring deployment %s/%s in cluster %s", meta.Namespace, meta.Name, meta.Cluster)

	err = kom.Cluster(meta.Cluster).WithContext(ctx).Resource(&appsv1.Deployment{}).Namespace(meta.Namespace).Name(meta.Name).Ctl().Scaler().Restore()
	if err != nil {
		return nil, err
	}

	return tools.TextResult("Successfully restored deployment", meta)
}
