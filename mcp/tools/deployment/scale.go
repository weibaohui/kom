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

// ScaleDeploymentTool 创建一个扩缩容Deployment的工具
func ScaleDeploymentTool() mcp.Tool {
	return mcp.NewTool(
		"scale_deployment",
		mcp.WithDescription("扩缩容Deployment，通过集群、命名空间、名称和副本数 / Scale deployment by cluster, namespace, name and replicas"),
		mcp.WithString("cluster", mcp.Description("运行Deployment的集群 / The cluster runs the deployment")),
		mcp.WithString("namespace", mcp.Description("Deployment所在的命名空间 / The namespace of the deployment")),
		mcp.WithString("name", mcp.Description("Deployment的名称 / The name of the deployment")),
		mcp.WithNumber("replicas", mcp.Description("目标副本数 / Target number of replicas")),
	)
}

// ScaleDeploymentHandler 处理扩缩容Deployment的请求
func ScaleDeploymentHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 获取参数
	ctx, meta, err := metadata.ParseFromRequest(ctx, request, config)

	if err != nil {
		return nil, err
	}

	replicas := int32(1)
	if replicasVal, ok := request.Params.Arguments["replicas"].(float64); ok {
		replicas = int32(replicasVal)
	}
	klog.Infof("Scaling deployment %s/%s in cluster %s to %d replicas", meta.Namespace, meta.Name, meta.Cluster, replicas)

	err = kom.Cluster(meta.Cluster).WithContext(ctx).Resource(&appsv1.Deployment{}).Namespace(meta.Namespace).Name(meta.Name).Ctl().Deployment().Scale(replicas)
	if err != nil {
		return nil, err
	}

	return utils.TextResult("Successfully scaled deployment", meta)
}
