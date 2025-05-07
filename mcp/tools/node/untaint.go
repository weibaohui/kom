package node

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/weibaohui/kom/kom"
	"github.com/weibaohui/kom/mcp/tools"
	v1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
)

// UnTaintNodeTool 创建一个为节点移除污点的工具
func UnTaintNodeTool() mcp.Tool {
	return mcp.NewTool(
		"untaint_k8s_node",
		mcp.WithDescription("为节点移除污点，等同于kubectl taint nodes <node> <key>=<value>:<effect>- / Remove taint from node, equivalent to kubectl taint nodes <node> <key>=<value>:<effect>-"),
		mcp.WithString("cluster", mcp.Description("节点所在的集群 （使用空字符串表示默认集群）/ The cluster of the node")),
		mcp.WithString("name", mcp.Description("节点名称 / The name of the node")),
		mcp.WithString("taint", mcp.Description("污点表达式，格式为key=value:effect / Taint expression in format key=value:effect")),
	)
}

// UnTaintNodeHandler 处理为节点移除污点的请求
func UnTaintNodeHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 获取参数
	ctx, meta, err := tools.ParseFromRequest(ctx, request)
	if err != nil {
		return nil, err
	}

	taint := request.Params.Arguments["taint"].(string)

	klog.Infof("Removing taint %s from node %s in cluster %s", taint, meta.Name, meta.Cluster)

	err = kom.Cluster(meta.Cluster).WithContext(ctx).Resource(&v1.Node{}).Name(meta.Name).Ctl().Node().UnTaint(taint)
	if err != nil {
		return nil, err
	}

	return tools.TextResult("Successfully removed taint from node", meta)
}
