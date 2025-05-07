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

// UpdateTagDeploymentTool 创建一个更新Deployment镜像Tag的工具
func UpdateTagDeploymentTool() mcp.Tool {
	return mcp.NewTool(
		"update_k8s_deployment_image_tag",
		mcp.WithDescription("更新Deployment中容器的镜像Tag。对应kubectl命令: kubectl set image deployment/<name> <container>=<image>:<tag> -n <namespace> / Update container image tag in deployment. Equivalent kubectl command: kubectl set image deployment/<name> <container>=<image>:<tag> -n <namespace>"),
		mcp.WithString("cluster", mcp.Required(), mcp.Description("运行Deployment的集群 （使用空字符串表示默认集群）/ The cluster runs the deployment")),
		mcp.WithString("namespace", mcp.Description("Deployment所在的命名空间 / The namespace of the deployment")),
		mcp.WithString("name", mcp.Description("Deployment的名称 / The name of the deployment")),
		mcp.WithString("container", mcp.Description("容器名称 / Container name")),
		mcp.WithString("tag", mcp.Description("新的镜像Tag / New image tag")),
	)
}

// UpdateTagDeploymentHandler 处理更新Deployment镜像Tag的请求
func UpdateTagDeploymentHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

	container := request.Params.Arguments["container"].(string)
	tag := request.Params.Arguments["tag"].(string)

	klog.Infof("Updating deployment %s/%s container %s image tag to %s in cluster %s", meta.Namespace, meta.Name, container, tag, meta.Cluster)

	_, err = kom.Cluster(meta.Cluster).WithContext(ctx).Resource(&appsv1.Deployment{}).Namespace(meta.Namespace).Name(meta.Name).Ctl().Deployment().ReplaceImageTag(container, tag)
	if err != nil {
		return nil, err
	}

	return utils.TextResult("Successfully updated deployment image tag", meta)
}
