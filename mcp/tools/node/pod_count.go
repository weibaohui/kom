package node

import (
	"context"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/weibaohui/kom/kom"
	"github.com/weibaohui/kom/mcp/tools"
	v1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
)

// NodePodCountTool 返回一个用于查询指定 Kubernetes 节点上运行 Pod 数量的工具。
// 工具支持指定节点所属集群（空字符串表示默认集群）、节点名称（必填），以及可选的缓存时间（默认20秒）。
func NodePodCountTool() mcp.Tool {
	return mcp.NewTool(
		"get_k8s_pod_count_running_on_node",
		mcp.WithDescription("查询某个节点上运行的Pod数量统计 (类似命令: kubectl describe node <node-name> | grep Pods) / Query node Pod count statistics"),
		mcp.WithString("cluster", mcp.Description("节点所在的集群 （使用空字符串表示默认集群）/ The cluster of the node")),
		mcp.WithString("name", mcp.Required(), mcp.Description("节点名称 / The name of the node")),
		mcp.WithNumber("cache_seconds", mcp.Description("缓存时间（默认20秒） / Cache duration in seconds,default 20 seconds")),
	)
}

// NodePodCountHandler 处理查询指定 Kubernetes 节点上 Pod 数量的请求，并返回总数、已用数和可用数。
func NodePodCountHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 获取参数
	ctx, meta, err := tools.ParseFromRequest(ctx, request)
	if err != nil {
		return nil, err
	}

	cacheSeconds := int32(20)
	if cacheSecondsVal, ok := request.Params.Arguments["cacheSeconds"].(float64); ok {
		cacheSeconds = int32(cacheSecondsVal)
	}

	klog.Infof("Querying Pod count for node %s in cluster %s with cache duration %d seconds", meta.Name, meta.Cluster, cacheSeconds)

	// 查询节点Pod数量
	total, used, available := kom.Cluster(meta.Cluster).WithContext(ctx).Resource(&v1.Node{}).WithCache(time.Duration(cacheSeconds) * time.Second).Name(meta.Name).Ctl().Node().PodCount()

	// 构建返回结果
	result := map[string]interface{}{
		"total":     total,
		"used":      used,
		"available": available,
	}

	return tools.TextResult(result, meta)
}
