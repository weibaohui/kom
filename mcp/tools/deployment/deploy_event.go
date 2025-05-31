package deployment

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/weibaohui/kom/kom"
	"github.com/weibaohui/kom/mcp/tools"
	v1 "k8s.io/api/events/v1"
)

func ListDeployEventResource() mcp.Tool {
	return mcp.NewTool(
		"list_k8s_deploy_event",
		mcp.WithDescription("列出Deployment相关的事件。 kubectl get events -n <namespace> ) "),
		mcp.WithString("cluster", mcp.Description("运行事件的集群（使用空字符串表示默认集群）")),
		mcp.WithString("namespace", mcp.Required(), mcp.Description("Deploy所在的命名空间")),
		mcp.WithString("name", mcp.Required(), mcp.Description("Deploy名称")),
	)
}

func ListDeployEventResourceHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 获取资源元数据
	ctx, meta, err := tools.ParseFromRequest(ctx, request)
	if err != nil {
		return nil, err
	}

	// 获取标签选择器和涉及对象名称
	// involvedObjectName := request.Params.Arguments["name"].(string)
	involvedObjectName := request.GetString("name", "")
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
