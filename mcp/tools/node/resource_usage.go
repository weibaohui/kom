package node

import (
	"context"
	"fmt"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/weibaohui/kom/kom"
	"github.com/weibaohui/kom/mcp/metadata"
	"github.com/weibaohui/kom/utils"

	v1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
)

// NodeResourceUsageTool 创建一个用于查询 Kubernetes 集群中节点资源使用情况的工具。
// 工具支持指定集群、节点名称和缓存时间，返回类似 kubectl describe node <node-name> | grep -A 5 Allocated 的资源统计信息。
func NodeResourceUsageTool() mcp.Tool {
	return mcp.NewTool(
		"get_node_k8s_resource_usage",
		mcp.WithDescription("查询节点资源使用情况统计 (类似命令: kubectl describe node <node-name> | grep -A 5 Allocated) / Query node resource usage statistics"),
		mcp.WithString("cluster", mcp.Required(), mcp.Description("节点所在的集群 （使用空字符串表示默认集群）/ The cluster of the node")),
		mcp.WithString("name", mcp.Description("节点名称 / The name of the node")),
		mcp.WithNumber("cache_seconds", mcp.Description("缓存时间（默认20秒） / Cache duration in seconds,default 20 seconds")),
	)
}

// NodeResourceUsageHandler 处理查询节点资源使用情况的请求
func NodeResourceUsageHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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
	cacheSeconds := int32(20)
	if cacheSecondsVal, ok := request.Params.Arguments["cacheSeconds"].(float64); ok {
		cacheSeconds = int32(cacheSecondsVal)
	}

	klog.Infof("Querying resource usage for node %s in cluster %s with cache duration %d seconds", meta.Name, meta.Cluster, cacheSeconds)

	// 查询节点资源使用情况
	usage, err := kom.Cluster(meta.Cluster).WithContext(ctx).Resource(&v1.Node{}).WithCache(time.Duration(cacheSeconds) * time.Second).Name(meta.Name).Ctl().Node().ResourceUsage()
	if err != nil {
		return nil, err
	}
	return utils.TextResult(usage, meta)
}
