package pod

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/weibaohui/kom/kom"
	"github.com/weibaohui/kom/mcp/tools"
	v1 "k8s.io/api/events/v1"
)

// ListPodEventResource 返回一个用于列出指定 Pod 相关事件的工具实例。该工具支持指定集群、命名空间和 Pod 名称参数。
func ListPodEventResource() mcp.Tool {
	return mcp.NewTool(
		"list_k8s_pod_event",
		mcp.WithDescription("列出Pod相关的事件。 kubectl get events -n <namespace> ) "),
		mcp.WithString("cluster", mcp.Description("运行事件的集群（使用空字符串表示默认集群）")),
		mcp.WithString("namespace", mcp.Required(), mcp.Description("Pod所在的命名空间")),
		mcp.WithString("name", mcp.Required(), mcp.Description("Pod名称")),
	)
}

// ListPodEventResourceHandler 处理 Pod 事件查询请求，返回指定集群、命名空间和 Pod 名称相关的事件列表。
// 如果未指定命名空间，则查询所有命名空间下与该 Pod 相关的事件。
// 返回格式化后的事件列表结果，如查询或参数解析失败则返回错误。
func ListPodEventResourceHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 获取资源元数据
	ctx, meta, err := tools.ParseFromRequest(ctx, request)
	if err != nil {
		return nil, err
	}

	// 获取标签选择器和涉及对象名称
	involvedObjectName, _ := request.Params.Arguments["name"].(string)

	// 获取事件列表
	var list []*v1.Event
	kubectl := kom.Cluster(meta.Cluster).WithContext(ctx).CRD("events.k8s.io", "v1", "Event").Namespace(meta.Namespace).RemoveManagedFields()
	if meta.Namespace == "" {
		kubectl = kubectl.AllNamespace()
	}

	if involvedObjectName != "" {
		// kubectl = kubectl.WithFieldSelector("involvedObject.name=" + involvedObjectName)
		kubectl = kubectl.WithFieldSelector("regarding.name=" + involvedObjectName)
	}
	err = kubectl.List(&list).Error
	if err != nil {
		return nil, fmt.Errorf("failed to list events: %v", err)
	}

	return tools.TextResult(list, meta)
}
