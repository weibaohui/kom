package daemonset

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/weibaohui/kom/kom"
	"github.com/weibaohui/kom/mcp/tools"
	appsv1 "k8s.io/api/apps/v1"
)

// RestartDaemonSetTool 创建一个重启DaemonSet的工具
func RestartDaemonSetTool() mcp.Tool {
	return mcp.NewTool(
		"restart_k8s_daemonset",
		mcp.WithDescription("通过集群、命名空间和名称,重启DaemonSet。对应kubectl命令: kubectl rollout restart daemonset/<name> -n <namespace> / Restart daemonset by cluster, namespace and name. Equivalent kubectl command: kubectl rollout restart daemonset/<name> -n <namespace>"),
		mcp.WithString("cluster", mcp.Description("运行DaemonSet的集群 （使用空字符串表示默认集群）/ The cluster runs the daemonset")),
		mcp.WithString("namespace", mcp.Description("DaemonSet所在的命名空间 / The namespace of the daemonset")),
		mcp.WithString("name", mcp.Description("DaemonSet的名称 / The name of the daemonset")),
	)
}

// RestartDaemonSetHandler 处理重启DaemonSet的请求
func RestartDaemonSetHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 获取参数
	ctx, meta, err := tools.ParseFromRequest(ctx, request)
	if err != nil {
		return nil, err
	}

	err = kom.Cluster(meta.Cluster).WithContext(ctx).Resource(&appsv1.DaemonSet{}).Namespace(meta.Namespace).Name(meta.Name).Ctl().DaemonSet().Restart()
	if err != nil {
		return nil, err
	}

	return tools.TextResult("Successfully restarted daemonset", meta)
}
