package node

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/weibaohui/kom/kom"
	"github.com/weibaohui/kom/mcp/tools"
	v1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
)

// CordonNodeTool 创建一个用于将 Kubernetes 节点设置为不可调度状态（cordon）的工具，相当于执行 kubectl cordon <node>。
func CordonNodeTool() mcp.Tool {
	return mcp.NewTool(
		"cordon_k8s_node",
		mcp.WithDescription("设置节点为不可调度状态，等同于kubectl cordon <node> / Mark node as unschedulable, equivalent to kubectl cordon <node>"),
		mcp.WithString("cluster", mcp.Description("节点所在的集群 （使用空字符串表示默认集群）/ The cluster of the node")),
		mcp.WithString("name", mcp.Required(), mcp.Description("节点名称 / The name of the node")),
	)
}

// CordonNodeHandler 处理将指定 Kubernetes 节点设置为不可调度（Cordon）状态的请求。
// 根据请求参数确定目标集群和节点，并执行 cordon 操作。
// 成功时返回操作结果，失败时返回错误。
func CordonNodeHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 获取参数
	ctx, meta, err := tools.ParseFromRequest(ctx, request)
	if err != nil {
		return nil, err
	}

	klog.Infof("Cordoning node %s in cluster %s", meta.Name, meta.Cluster)

	err = kom.Cluster(meta.Cluster).WithContext(ctx).Resource(&v1.Node{}).Name(meta.Name).Ctl().Node().Cordon()
	if err != nil {
		return nil, err
	}

	return tools.TextResult("Successfully cordoned node", meta)
}
