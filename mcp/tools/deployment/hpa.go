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

// HPAListDeploymentTool 创建一个查询Deployment的HPA列表的工具
func HPAListDeploymentTool() mcp.Tool {
	return mcp.NewTool(
		"get_deployment_hpa_list",
		mcp.WithDescription("查询Deployment的HPA列表 / Query deployment HPA list"),
		mcp.WithString("cluster", mcp.Description("运行Deployment的集群 / The cluster runs the deployment")),
		mcp.WithString("namespace", mcp.Description("Deployment所在的命名空间 / The namespace of the deployment")),
		mcp.WithString("name", mcp.Description("Deployment的名称 / The name of the deployment")),
	)
}

// HPAListDeploymentHandler 处理查询Deployment的HPA列表的请求
func HPAListDeploymentHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 获取参数
	meta, err := metadata.ParseFromRequest(request)
	if err != nil {
		return nil, err
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

	return tools.TextResult(result, meta)
}
