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

// TopNode 返回一个用于获取 Kubernetes 集群中节点 CPU 和内存资源用量排名的工具。
// 工具名称为 "get_k8s_top_node"，支持指定目标集群，空字符串表示默认集群。
func TopNode() mcp.Tool {
	return mcp.NewTool(
		"get_k8s_top_node",
		mcp.WithDescription("获取Node节点 CPU 内存 资源用量排名 列表 (类似命令 kubectl top nodes -n ns)"),
		mcp.WithString("cluster", mcp.Description("运行资源的集群（使用空字符串表示默认集群）/ Cluster where the resources are running (use empty string for default cluster)")),
	)
}

// TopNodeHandler 处理获取 Kubernetes 节点 CPU 和内存资源使用排行的请求。
// 解析请求中的资源元数据，调用 Kubernetes API 获取节点的资源使用情况，并以 JSON 格式返回结果。
// 如获取过程中发生错误，则返回相应的错误信息。
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
