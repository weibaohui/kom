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

// CordonNodeTool 创建一个为节点设置Cordon的工具
func CordonNodeTool() mcp.Tool {
	return mcp.NewTool(
		"cordon_node",
		mcp.WithDescription("设置节点为不可调度状态 / Mark node as unschedulable"),
		mcp.WithString("cluster", mcp.Description("节点所在的集群 / The cluster of the node")),
		mcp.WithString("name", mcp.Description("节点名称 / The name of the node")),
	)
}

// CordonNodeHandler 处理为节点设置Cordon的请求
func CordonNodeHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 获取参数
	meta, err := metadata.ParseFromRequest(request)
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
