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

// UnCordonNodeTool 创建一个为节点取消Cordon的工具
func UnCordonNodeTool() mcp.Tool {
	return mcp.NewTool(
		"uncordon_node",
		mcp.WithDescription("设置节点为可调度状态，等同于kubectl uncordon <node> / Mark node as schedulable, equivalent to kubectl uncordon <node>"),
		mcp.WithString("cluster", mcp.Required(), mcp.Description("节点所在的集群 （使用空字符串表示默认集群）/ The cluster of the node")),
		mcp.WithString("name", mcp.Description("节点名称 / The name of the node")),
	)
}

// UnCordonNodeHandler 处理为节点取消Cordon的请求
func UnCordonNodeHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

	klog.Infof("UnCordoning node %s in cluster %s", meta.Name, meta.Cluster)

	err = kom.Cluster(meta.Cluster).WithContext(ctx).Resource(&v1.Node{}).Name(meta.Name).Ctl().Node().UnCordon()
	if err != nil {
		return nil, err
	}

	return utils.TextResult("Successfully uncordoned node", meta)
}
