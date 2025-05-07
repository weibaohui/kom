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

// DrainNodeTool 返回一个用于清空 Kubernetes 节点的工具。
// 该工具会驱逐节点上的所有 Pod，并阻止新的 Pod 调度到该节点，相当于执行 kubectl drain <node>。
// 需要指定节点所在的集群（可用空字符串表示默认集群）和节点名称。
func DrainNodeTool() mcp.Tool {
	return mcp.NewTool(
		"drain_k8s_node",
		mcp.WithDescription("清空节点上的Pod并防止新的Pod调度，等同于kubectl drain <node> / Drain all pods from node and prevent new scheduling, equivalent to kubectl drain <node>"),
		mcp.WithString("cluster", mcp.Required(), mcp.Description("节点所在的集群 （使用空字符串表示默认集群）/ The cluster of the node")),
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

	klog.Infof("Draining node %s in cluster %s", meta.Name, meta.Cluster)

	err = kom.Cluster(meta.Cluster).WithContext(ctx).Resource(&v1.Node{}).Name(meta.Name).Ctl().Node().Drain()
	if err != nil {
		return nil, err
	}

	return utils.TextResult("Successfully drained node", meta)
}
