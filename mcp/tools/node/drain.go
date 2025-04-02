package node

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/weibaohui/kom/kom"
	"github.com/weibaohui/kom/mcp/metadata"
	"github.com/weibaohui/kom/utils"

	v1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
)

// DrainNodeTool 创建一个为节点执行Drain的工具
func DrainNodeTool() mcp.Tool {
	return mcp.NewTool(
		"drain_node",
		mcp.WithDescription("清空节点上的Pod并防止新的Pod调度 / Drain all pods from node and prevent new scheduling"),
		mcp.WithString("cluster", mcp.Description("节点所在的集群 / The cluster of the node")),
		mcp.WithString("name", mcp.Description("节点名称 / The name of the node")),
	)
}

// DrainNodeHandler 处理为节点执行Drain的请求
func DrainNodeHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 获取参数
	ctx, meta, err := metadata.ParseFromRequest(ctx, request, config)

	if err != nil {
		return nil, err
	}

	klog.Infof("Draining node %s in cluster %s", meta.Name, meta.Cluster)

	err = kom.Cluster(meta.Cluster).WithContext(ctx).Resource(&v1.Node{}).Name(meta.Name).Ctl().Node().Drain()
	if err != nil {
		return nil, err
	}

	return utils.TextResult("Successfully drained node", meta)
}
