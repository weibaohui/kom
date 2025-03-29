package deployment

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/weibaohui/kom/kom"
	"github.com/weibaohui/kom/mcp/tools"
	"github.com/weibaohui/kom/mcp/tools/metadata"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/klog/v2"
)

// StopDeploymentTool 创建一个停止Deployment的工具
func StopDeploymentTool() mcp.Tool {
	return mcp.NewTool(
		"stop_deployment",
		mcp.WithDescription("停止Deployment，将副本数设置为0并记录原始副本数到注解中 / Stop deployment by setting replicas to 0 and save original replicas to annotation"),
		mcp.WithString("cluster", mcp.Description("运行Deployment的集群 / The cluster runs the deployment")),
		mcp.WithString("namespace", mcp.Description("Deployment所在的命名空间 / The namespace of the deployment")),
		mcp.WithString("name", mcp.Description("Deployment的名称 / The name of the deployment")),
	)
}

// StopDeploymentHandler 处理停止Deployment的请求
func StopDeploymentHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 获取参数
	meta, err := metadata.ParseFromRequest(request)
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
