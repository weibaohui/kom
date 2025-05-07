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

// TaintNodeTool 返回一个用于为 Kubernetes 节点添加污点的工具。
// 该工具等同于 kubectl taint nodes 命令，支持指定集群、节点名称和污点表达式。
func TaintNodeTool() mcp.Tool {
	return mcp.NewTool(
		"taint_k8s_node",
		mcp.WithDescription("为节点添加污点，等同于kubectl taint nodes <node> <key>=<value>:<effect> / Add taint to node, equivalent to kubectl taint nodes <node> <key>=<value>:<effect>"),
		mcp.WithString("cluster", mcp.Required(), mcp.Description("节点所在的集群 （使用空字符串表示默认集群）/ The cluster of the node")),
		mcp.WithString("name", mcp.Description("节点名称 / The name of the node")),
		mcp.WithString("taint", mcp.Description("污点表达式，格式为key=value:effect / Taint expression in format key=value:effect")),
	)
}

// TaintNodeHandler 处理为节点添加污点的请求
func TaintNodeHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

	klog.Infof("Adding taint %s to node %s in cluster %s", taint, meta.Name, meta.Cluster)

	err = kom.Cluster(meta.Cluster).WithContext(ctx).Resource(&v1.Node{}).Name(meta.Name).Ctl().Node().Taint(taint)
	if err != nil {
		return nil, err
	}

	return utils.TextResult("Successfully added taint to node", meta)
}
