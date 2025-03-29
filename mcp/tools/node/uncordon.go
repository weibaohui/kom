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

// UnCordonNodeTool 创建一个为节点取消Cordon的工具
func UnCordonNodeTool() mcp.Tool {
	return mcp.NewTool(
		"uncordon_node",
		mcp.WithDescription("设置节点为可调度状态 / Mark node as schedulable"),
		mcp.WithString("cluster", mcp.Description("节点所在的集群 / The cluster of the node")),
		mcp.WithString("name", mcp.Description("节点名称 / The name of the node")),
	)
}

// UnCordonNodeHandler 处理为节点取消Cordon的请求
func UnCordonNodeHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 获取参数
	meta, err := metadata.ParseFromRequest(request)
	if err != nil {
		return nil, err
	}

	klog.Infof("UnCordoning node %s in cluster %s", meta.Name, meta.Cluster)

	err = kom.Cluster(meta.Cluster).WithContext(ctx).Resource(&v1.Node{}).Name(meta.Name).Ctl().Node().UnCordon()
	if err != nil {
		return nil, err
	}

	return tools.TextResult("Successfully uncordoned node", meta)
}
