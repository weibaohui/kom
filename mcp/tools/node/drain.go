package node

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/weibaohui/kom/kom"
	"github.com/weibaohui/kom/mcp/tools"
	v1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
)

// DrainNodeTool 返回一个用于清空指定 Kubernetes 节点并阻止新 Pod 调度的工具。
// 工具支持指定节点名称和所属集群（空字符串表示默认集群），功能等同于 kubectl drain <node>。
func DrainNodeTool() mcp.Tool {
	return mcp.NewTool(
		"drain_k8s_node",
		mcp.WithDescription("清空节点上的Pod并防止新的Pod调度，等同于kubectl drain <node> / Drain all pods from node and prevent new scheduling, equivalent to kubectl drain <node>"),
		mcp.WithString("cluster", mcp.Description("节点所在的集群 （使用空字符串表示默认集群）/ The cluster of the node")),
		mcp.WithString("name", mcp.Required(), mcp.Description("节点名称 / The name of the node")),
	)
}

// DrainNodeHandler 处理节点的驱逐请求，将指定 Kubernetes 节点上的所有 Pod 清空并阻止新 Pod 调度。
// 成功时返回操作结果文本，失败时返回错误。
func DrainNodeHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 获取参数
	ctx, meta, err := tools.ParseFromRequest(ctx, request)
	if err != nil {
		return nil, err
	}

	klog.Infof("Draining node %s in cluster %s", meta.Name, meta.Cluster)

	err = kom.Cluster(meta.Cluster).WithContext(ctx).Resource(&v1.Node{}).Name(meta.Name).Ctl().Node().Drain()
	if err != nil {
		return nil, err
	}

	return tools.TextResult("Successfully drained node", meta)
}
