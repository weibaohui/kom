package node

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/weibaohui/kom/kom"
	"github.com/weibaohui/kom/mcp/tools"
	"github.com/weibaohui/kom/mcp/tools/metadata"
	v1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
)

// TaintNodeTool 创建一个为节点添加污点的工具
func TaintNodeTool() mcp.Tool {
	return mcp.NewTool(
		"taint_node",
		mcp.WithDescription("为节点添加污点 / Add taint to node"),
		mcp.WithString("cluster", mcp.Description("节点所在的集群 / The cluster of the node")),
		mcp.WithString("name", mcp.Description("节点名称 / The name of the node")),
		mcp.WithString("taint", mcp.Description("污点表达式，格式为key=value:effect / Taint expression in format key=value:effect")),
	)
}

// TaintNodeHandler 处理为节点添加污点的请求
func TaintNodeHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 获取参数
	meta, err := metadata.ParseFromRequest(request)
	if err != nil {
		return nil, err
	}

	taint := request.Params.Arguments["taint"].(string)

	klog.Infof("Adding taint %s to node %s in cluster %s", taint, meta.Name, meta.Cluster)

	err = kom.Cluster(meta.Cluster).WithContext(ctx).Resource(&v1.Node{}).Name(meta.Name).Ctl().Node().Taint(taint)
	if err != nil {
		return nil, err
	}

	return tools.TextResult("Successfully added taint to node", meta)
}
