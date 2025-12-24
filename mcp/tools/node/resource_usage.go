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

// NodeResourceUsageTool 创建一个查询节点资源使用情况的工具
func NodeResourceUsageTool() mcp.Tool {
	return mcp.NewTool(
		"get_k8s_node_resource_usage",
		mcp.WithDescription("查询节点资源使用情况统计，包括内存、CPU用量 (类似命令: kubectl describe node <node-name> | grep -A 5 Allocated) / Query node resource usage statistics"),
		mcp.WithTitleAnnotation("Get Node Resource Usage"),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithString("cluster", mcp.Description("节点所在的集群 （使用空字符串表示默认集群）/ The cluster of the node")),
		mcp.WithString("name", mcp.Required(), mcp.Description("节点名称 / The name of the node")),
		mcp.WithNumber("cache_seconds", mcp.Description("缓存时间（默认20秒） / Cache duration in seconds,default 20 seconds")),
	)
}

// NodeResourceUsageHandler 处理查询节点资源使用情况的请求
func NodeResourceUsageHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 获取参数
	ctx, meta, err := tools.ParseFromRequest(ctx, request)
	if err != nil {
		return nil, err
	}

	cacheSeconds := int32(request.GetInt("cacheSeconds", 20))

	klog.Infof("Querying resource usage for node %s in cluster %s with cache duration %d seconds", meta.Name, meta.Cluster, cacheSeconds)

	// 查询节点资源使用情况
	usage, err := kom.Cluster(meta.Cluster).WithContext(ctx).Resource(&v1.Node{}).WithCache(time.Duration(cacheSeconds) * time.Second).Name(meta.Name).Ctl().Node().ResourceUsage()
	if err != nil {
		return nil, err
	}
	return tools.TextResult(usage, meta)
}
