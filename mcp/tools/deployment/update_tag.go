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

// UpdateTagDeploymentTool 创建一个更新Deployment镜像Tag的工具
func UpdateTagDeploymentTool() mcp.Tool {
	return mcp.NewTool(
		"update_deployment_image_tag",
		mcp.WithDescription("更新Deployment中容器的镜像Tag / Update container image tag in deployment"),
		mcp.WithString("cluster", mcp.Description("运行Deployment的集群 / The cluster runs the deployment")),
		mcp.WithString("namespace", mcp.Description("Deployment所在的命名空间 / The namespace of the deployment")),
		mcp.WithString("name", mcp.Description("Deployment的名称 / The name of the deployment")),
		mcp.WithString("container", mcp.Description("容器名称 / Container name")),
		mcp.WithString("tag", mcp.Description("新的镜像Tag / New image tag")),
	)
}

// UpdateTagDeploymentHandler 处理更新Deployment镜像Tag的请求
func UpdateTagDeploymentHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 获取参数
	meta, err := metadata.ParseFromRequest(request)
	if err != nil {
		return nil, err
	}

	container := request.Params.Arguments["container"].(string)
	tag := request.Params.Arguments["tag"].(string)

	klog.Infof("Updating deployment %s/%s container %s image tag to %s in cluster %s", meta.Namespace, meta.Name, container, tag, meta.Cluster)

	_, err = kom.Cluster(meta.Cluster).WithContext(ctx).Resource(&appsv1.Deployment{}).Namespace(meta.Namespace).Name(meta.Name).Ctl().Deployment().ReplaceImageTag(container, tag)
	if err != nil {
		return nil, err
	}

	return tools.TextResult("Successfully updated deployment image tag", meta)
}
