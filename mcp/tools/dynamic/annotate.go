package dynamic

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/weibaohui/kom/kom"
	"github.com/weibaohui/kom/mcp/tools"
)

func AnnotateDynamicResource() mcp.Tool {
	return mcp.NewTool(
		"annotate_k8s_resource",
		mcp.WithDescription("为Kubernetes资源添加或删除注解 / Add or remove annotations for Kubernetes resource"),
		mcp.WithTitleAnnotation("Annotate Resource"),
		mcp.WithDestructiveHintAnnotation(true),
		mcp.WithString("cluster", mcp.Description("运行资源的集群（使用空字符串表示默认集群）/ Cluster where the resources are running (use empty string for default cluster)")),
		mcp.WithString("namespace", mcp.Description("资源所在的命名空间（集群范围资源可选）/ Namespace of the resource (optional for cluster-scoped resources)")),
		mcp.WithString("name", mcp.Description("资源的名称 / Name of the resource")),
		mcp.WithString("group", mcp.Description("资源的API组 / API group of the resource")),
		mcp.WithString("version", mcp.Description("资源的API版本 / API version of the resource")),
		mcp.WithString("kind", mcp.Description("资源的类型 / Kind of the resource")),
		mcp.WithString("annotation", mcp.Description("要添加或删除的注解（使用key=value添加，key-删除）/ Annotation to add or remove (use key=value to add, key- to remove)")),
	)
}

func AnnotateDynamicResourceHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 获取资源元数据
	ctx, meta, err := tools.ParseFromRequest(ctx, request)
	if err != nil {
		return nil, err
	}

	// 获取注解操作
	// annotation, ok := request.Params.Arguments["annotation"].(string)
	annotation := request.GetString("annotation", "")
	if annotation == "" {
		return nil, fmt.Errorf("annotation parameter is required")
	}

	// 处理资源
	kubectl := kom.Cluster(meta.Cluster).WithContext(ctx).CRD(meta.Group, meta.Version, meta.Kind).Namespace(meta.Namespace)
	if meta.Namespace == "" {
		kubectl = kubectl.AllNamespace()
	}

	// 执行注解操作
	err = kubectl.Name(meta.Name).Ctl().Annotate(annotation)
	if err != nil {
		return nil, fmt.Errorf("failed to update annotation for [%s/%s] type of [%s%s%s]: %v", meta.Namespace, meta.Name, meta.Group, meta.Version, meta.Kind, err)
	}

	result := fmt.Sprintf("Successfully updated annotation for resource [%s/%s] of type [%s%s%s]", meta.Namespace, meta.Name, meta.Group, meta.Version, meta.Kind)
	return tools.TextResult(result, meta)
}
