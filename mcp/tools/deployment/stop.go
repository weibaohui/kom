package deployment

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/weibaohui/kom/kom"
	"github.com/weibaohui/kom/mcp/tools"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/klog/v2"
)

// StopDeploymentTool 创建一个用于停止 Kubernetes Deployment 的工具。
// 该工具通过将 Deployment 的副本数设置为 0 来停止其运行，并将原始副本数记录到注解中，便于后续恢复。
// 工具参数包括目标集群（空字符串表示默认集群）、命名空间和 Deployment 名称。
func StopDeploymentTool() mcp.Tool {
	return mcp.NewTool(
		"stop_k8s_deployment",
		mcp.WithDescription("停止Deployment。（将副本数设置为0并记录原始副本数到注解中，恢复是可使用restore_deployment方法）。对应kubectl命令: kubectl scale deployment/<name> --replicas=0 -n <namespace> / Stop deployment by setting replicas to 0 and save original replicas to annotation. Equivalent kubectl command: kubectl scale deployment/<name> --replicas=0 -n <namespace>"),
		mcp.WithString("cluster", mcp.Description("运行Deployment的集群 （使用空字符串表示默认集群）/ The cluster runs the deployment")),
		mcp.WithString("namespace", mcp.Description("Deployment所在的命名空间 / The namespace of the deployment")),
		mcp.WithString("name", mcp.Description("Deployment的名称 / The name of the deployment")),
	)
}

// StopDeploymentHandler 根据请求参数停止指定的 Kubernetes Deployment，将其副本数缩减为零，并在注解中保存原始副本数。
func StopDeploymentHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 获取参数
	ctx, meta, err := tools.ParseFromRequest(ctx, request)
	if err != nil {
		return nil, err
	}

	klog.Infof("Stopping deployment %s/%s in cluster %s", meta.Namespace, meta.Name, meta.Cluster)

	err = kom.Cluster(meta.Cluster).WithContext(ctx).Resource(&appsv1.Deployment{}).Namespace(meta.Namespace).Name(meta.Name).Ctl().Scaler().Stop()
	if err != nil {
		return nil, err
	}

	return tools.TextResult("Successfully stopped deployment", meta)
}
