package node

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/weibaohui/kom/kom"
	"github.com/weibaohui/kom/mcp/metadata"
	"github.com/weibaohui/kom/utils"

	v1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
)

// UnTaintNodeTool 创建一个用于移除 Kubernetes 节点污点的工具。
// 该工具等同于执行 kubectl taint nodes <node> <key>=<value>:<effect>-，用于指定集群中的节点。
// 参数包括集群名称（可为空表示默认集群）、节点名称和污点表达式。
func UnTaintNodeTool() mcp.Tool {
	return mcp.NewTool(
		"untaint_k8s_node",
		mcp.WithDescription("为节点移除污点，等同于kubectl taint nodes <node> <key>=<value>:<effect>- / Remove taint from node, equivalent to kubectl taint nodes <node> <key>=<value>:<effect>-"),
		mcp.WithString("cluster", mcp.Required(), mcp.Description("节点所在的集群 （使用空字符串表示默认集群）/ The cluster of the node")),
		mcp.WithString("name", mcp.Description("节点名称 / The name of the node")),
		mcp.WithString("taint", mcp.Description("污点表达式，格式为key=value:effect / Taint expression in format key=value:effect")),
	)
}

// UnTaintNodeHandler 处理为节点移除污点的请求
func UnTaintNodeHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 获取参数
	ctx, meta, err := metadata.ParseFromRequest(ctx, request, config)

	if err != nil {
		return nil, err
	}
	// 如果只有一个集群的时候，使用空，默认集群
	// 如果大于一个集群，没有传值，那么要返回错误
	if len(kom.Clusters().AllClusters()) > 1 && meta.Cluster == "" {
		return nil, fmt.Errorf("cluster is required, 集群名称必须设置")
	}
	if len(kom.Clusters().AllClusters()) == 1 && meta.Cluster == "" {
		meta.Cluster = kom.Clusters().DefaultCluster().ID
	}
	if kom.Clusters().GetClusterById(meta.Cluster) == nil {
		return nil, fmt.Errorf("cluster %s not found 集群不存在，请检查集群名称", meta.Cluster)
	}

	taint := request.Params.Arguments["taint"].(string)

	klog.Infof("Removing taint %s from node %s in cluster %s", taint, meta.Name, meta.Cluster)

	err = kom.Cluster(meta.Cluster).WithContext(ctx).Resource(&v1.Node{}).Name(meta.Name).Ctl().Node().UnTaint(taint)
	if err != nil {
		return nil, err
	}

	return utils.TextResult("Successfully removed taint from node", meta)
}
