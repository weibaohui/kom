package dynamic

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/weibaohui/kom/kom"
	"github.com/weibaohui/kom/mcp/metadata"
	"github.com/weibaohui/kom/utils"
)

// LabelDynamicResource 返回一个用于为Kubernetes资源动态添加或删除标签的工具。
// 该工具要求指定目标集群、资源名称、API组、版本、类型等信息，并通过label参数决定添加或删除标签。
func LabelDynamicResource() mcp.Tool {
	return mcp.NewTool(
		"label_k8s_resource",
		mcp.WithDescription("为Kubernetes资源添加或删除标签 / Add or remove labels for Kubernetes resource"),
		mcp.WithString("cluster", mcp.Required(), mcp.Description("运行资源的集群（使用空字符串表示默认集群）/ Cluster where the resources are running (use empty string for default cluster)")),
		mcp.WithString("namespace", mcp.Description("资源所在的命名空间（集群范围资源可选）/ Namespace of the resource (optional for cluster-scoped resources)")),
		mcp.WithString("name", mcp.Description("资源的名称 / Name of the resource")),
		mcp.WithString("group", mcp.Description("资源的API组 / API group of the resource")),
		mcp.WithString("version", mcp.Description("资源的API版本 / API version of the resource")),
		mcp.WithString("kind", mcp.Description("资源的类型 / Kind of the resource")),
		mcp.WithString("label", mcp.Description("要添加或删除的标签（使用key=value添加，key-删除）/ Label to add or remove (use key=value to add, key- to remove)")),
	)
}

func LabelDynamicResourceHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 获取资源元数据
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

	// 获取标签操作
	label, ok := request.Params.Arguments["label"].(string)
	if !ok || label == "" {
		return nil, fmt.Errorf("label parameter is required")
	}

	// 处理资源
	kubectl := kom.Cluster(meta.Cluster).WithContext(ctx).CRD(meta.Group, meta.Version, meta.Kind).Namespace(meta.Namespace)
	if meta.Namespace == "" {
		kubectl = kubectl.AllNamespace()
	}

	// 执行标签操作
	err = kubectl.Name(meta.Name).Ctl().Label(label)
	if err != nil {
		return nil, fmt.Errorf("failed to update label for [%s/%s] type of [%s%s%s]: %v", meta.Namespace, meta.Name, meta.Group, meta.Version, meta.Kind, err)
	}

	result := fmt.Sprintf("Successfully updated label for resource [%s/%s] of type [%s%s%s]", meta.Namespace, meta.Name, meta.Group, meta.Version, meta.Kind)
	return utils.TextResult(result, meta)
}
