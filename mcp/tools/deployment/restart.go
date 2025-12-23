package deployment

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/weibaohui/kom/kom"
	"github.com/weibaohui/kom/mcp/tools"
	appsv1 "k8s.io/api/apps/v1"
)

// RestartDeploymentTool 创建一个重启Deployment的工具
func RestartDeploymentTool() mcp.Tool {
	return mcp.NewTool(
		"restart_k8s_deployment",
		mcp.WithDescription("通过集群、命名空间和名称,重启Deployment。对应kubectl命令: kubectl rollout restart deployment/<name> -n <namespace> / Restart deployment by cluster, namespace and name. Equivalent kubectl command: kubectl rollout restart deployment/<name> -n <namespace>"),
		mcp.WithTitleAnnotation("Restart Deployment"),
		mcp.WithDestructiveHintAnnotation(true),
		mcp.WithString("cluster", mcp.Description("运行Deployment的集群 （使用空字符串表示默认集群）/ The cluster runs the deployment")),
		mcp.WithString("namespace", mcp.Description("Deployment所在的命名空间 / The namespace of the deployment")),
		mcp.WithString("name", mcp.Description("Deployment的名称 / The name of the deployment")),
	)
}

// RestartDeploymentHandler 处理重启Deployment的请求
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
