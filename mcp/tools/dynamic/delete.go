package dynamic

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/weibaohui/kom/kom"
	"github.com/weibaohui/kom/mcp/tools"
)

// DeleteDynamicResource 返回一个用于根据集群、命名空间和资源元数据动态删除 Kubernetes 资源的工具定义。
func DeleteDynamicResource() mcp.Tool {
	return mcp.NewTool(
		"delete_k8s_resource",
		mcp.WithDescription("通过集群、命名空间和名称删除Kubernetes资源 / Delete Kubernetes resource by cluster, namespace, and name"),
		mcp.WithString("cluster", mcp.Description("运行资源的集群（使用空字符串表示默认集群）/ Cluster where the resources are running (use empty string for default cluster)")),
		mcp.WithString("namespace", mcp.Description("资源所在的命名空间（集群范围资源可选）/ Namespace of the resource (optional for cluster-scoped resources)")),
		mcp.WithString("name", mcp.Description("资源的名称 / Name of the resource")),
		mcp.WithString("group", mcp.Description("资源的API组 / API group of the resource")),
		mcp.WithString("version", mcp.Description("资源的API版本 / API version of the resource")),
		mcp.WithString("kind", mcp.Description("资源的类型 / Kind of the resource")),
		mcp.WithBoolean("force", mcp.Description("强制删除资源 / Force delete the resource")),
	)
}

// DeleteDynamicResourceHandler 根据请求参数删除指定的 Kubernetes 资源，支持普通删除和强制删除。
// 若删除失败，返回详细的错误信息；删除成功则返回操作结果描述。
func DeleteDynamicResourceHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 获取资源元数据
	ctx, meta, err := tools.ParseFromRequest(ctx, request)
	if err != nil {
		return nil, err
	}

	// 删除资源
	kubectl := kom.Cluster(meta.Cluster).WithContext(ctx).CRD(meta.Group, meta.Version, meta.Kind).Namespace(meta.Namespace)
	if meta.Namespace == "" {
		kubectl = kubectl.AllNamespace()
	}
	if force, ok := request.Params.Arguments["force"].(bool); ok && force {
		err = kubectl.Name(meta.Name).ForceDelete().Error
	} else {
		err = kubectl.Name(meta.Name).Delete().Error
	}
	if err != nil {
		return nil, fmt.Errorf("failed to delete item [%s/%s] type of [%s%s%s]: %v", meta.Namespace, meta.Name, meta.Group, meta.Version, meta.Kind, err)
	}
	result := fmt.Sprintf("Successfully deleted resource [%s/%s] of type [%s%s%s]", meta.Namespace, meta.Name, meta.Group, meta.Version, meta.Kind)
	return tools.TextResult(result, meta)

}
