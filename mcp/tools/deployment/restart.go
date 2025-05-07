package deployment

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/weibaohui/kom/kom"
	"github.com/weibaohui/kom/mcp/tools"
	appsv1 "k8s.io/api/apps/v1"
)

// RestartDeploymentTool 创建一个用于重启 Kubernetes Deployment 的工具。
// 工具通过指定集群、命名空间和 Deployment 名称，实现 Deployment 的重启操作，等效于 kubectl rollout restart deployment/<name> -n <namespace> 命令。
func RestartDeploymentTool() mcp.Tool {
	return mcp.NewTool(
		"restart_k8s_deployment",
		mcp.WithDescription("通过集群、命名空间和名称,重启Deployment。对应kubectl命令: kubectl rollout restart deployment/<name> -n <namespace> / Restart deployment by cluster, namespace and name. Equivalent kubectl command: kubectl rollout restart deployment/<name> -n <namespace>"),
		mcp.WithString("cluster", mcp.Description("运行Deployment的集群 （使用空字符串表示默认集群）/ The cluster runs the deployment")),
		mcp.WithString("namespace", mcp.Description("Deployment所在的命名空间 / The namespace of the deployment")),
		mcp.WithString("name", mcp.Description("Deployment的名称 / The name of the deployment")),
	)
}

// RestartDeploymentHandler 处理重启指定 Kubernetes Deployment 的请求，并返回操作结果。
func RestartDeploymentHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 获取参数
	ctx, meta, err := tools.ParseFromRequest(ctx, request)
	if err != nil {
		return nil, err
	}

	err = kom.Cluster(meta.Cluster).WithContext(ctx).Resource(&appsv1.Deployment{}).Namespace(meta.Namespace).Name(meta.Name).Ctl().Deployment().Restart()
	if err != nil {
		return nil, err
	}

	return tools.TextResult("Successfully restarted deployment", meta)
}
