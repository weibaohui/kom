package deployment

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/weibaohui/kom/kom"
	"github.com/weibaohui/kom/mcp/metadata"
	"github.com/weibaohui/kom/utils"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/klog/v2"
)

// RestoreDeploymentTool 创建一个恢复Deployment的工具
func RestoreDeploymentTool() mcp.Tool {
	return mcp.NewTool(
		"restore_deployment",
		mcp.WithDescription("恢复Deployment，从注解中恢复原始副本数，如果没有注解则默认为1 / Restore deployment replicas from annotation, default to 1 if not found"),
		mcp.WithString("cluster", mcp.Description("运行Deployment的集群 / The cluster runs the deployment")),
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

	klog.Infof("Restoring deployment %s/%s in cluster %s", meta.Namespace, meta.Name, meta.Cluster)

	err = kom.Cluster(meta.Cluster).WithContext(ctx).Resource(&appsv1.Deployment{}).Namespace(meta.Namespace).Name(meta.Name).Ctl().Scaler().Restore()
	if err != nil {
		return nil, err
	}

	return utils.TextResult("Successfully restored deployment", meta)
}
