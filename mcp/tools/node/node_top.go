package node

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/weibaohui/kom/kom"
	"github.com/weibaohui/kom/mcp/tools"
	"github.com/weibaohui/kom/utils"
	v1 "k8s.io/api/core/v1"
)

func TopNode() mcp.Tool {
	return mcp.NewTool(
		"get_k8s_top_node",
		mcp.WithDescription("获取Node节点 CPU 内存 资源用量排名 列表 (类似命令 kubectl top nodes -n ns)"),
		mcp.WithTitleAnnotation("Get Node Resource Usage"),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithString("cluster", mcp.Description("运行资源的集群（使用空字符串表示默认集群）/ Cluster where the resources are running (use empty string for default cluster)")),
	)
}

func TopNodeHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

	// 获取资源元数据
	ctx, meta, err := tools.ParseFromRequest(ctx, request)
	if err != nil {
		return nil, err
	}

	kubectl := kom.Cluster(meta.Cluster).WithContext(ctx).
		Resource(&v1.Node{}).
		RemoveManagedFields()

	top, err := kubectl.Ctl().Node().Top()
	if err != nil {
		return nil, fmt.Errorf("failed to  kubectl top pod list items type of [%s%s%s]: %v", meta.Group, meta.Version, meta.Kind, err)
	}

	return tools.TextResult(utils.ToJSON(top), meta)
}
