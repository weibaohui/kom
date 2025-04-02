package node

import (
	"context"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/weibaohui/kom/kom"
	"github.com/weibaohui/kom/mcp/metadata"
	"github.com/weibaohui/kom/utils"

	v1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
)

// NodeIPUsageTool 创建一个查询节点IP资源使用情况的工具
func NodeIPUsageTool() mcp.Tool {
	return mcp.NewTool(
		"get_node_ip_usage",
		mcp.WithDescription("查询节点IP资源使用情况 / Query node IP resource usage"),
		mcp.WithString("cluster", mcp.Description("节点所在的集群 / The cluster of the node")),
		mcp.WithString("name", mcp.Description("节点名称 / The name of the node")),
		mcp.WithNumber("cache_seconds", mcp.Description("缓存时间（默认20秒） / Cache duration in seconds,default 20 seconds")),
	)
}

// NodeIPUsageHandler 处理查询节点IP资源使用情况的请求
func NodeIPUsageHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 获取参数
	ctx, meta, err := metadata.ParseFromRequest(ctx, request, config)

	if err != nil {
		return nil, err
	}
	cacheSeconds := int32(20)
	if cacheSecondsVal, ok := request.Params.Arguments["cacheSeconds"].(float64); ok {
		cacheSeconds = int32(cacheSecondsVal)
	}

	klog.Infof("Querying IP usage for node %s in cluster %s with cache duration %d seconds", meta.Name, meta.Cluster, cacheSeconds)

	// 查询节点IP资源使用情况
	total, used, available := kom.Cluster(meta.Cluster).WithContext(ctx).Resource(&v1.Node{}).WithCache(time.Duration(cacheSeconds) * time.Second).Name(meta.Name).Ctl().Node().IPUsage()

	// 构建返回结果
	result := map[string]interface{}{
		"total":     total,
		"used":      used,
		"available": available,
	}

	return utils.TextResult(result, meta)
}
