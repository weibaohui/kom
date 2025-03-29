package pod

import (
	"context"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/weibaohui/kom/kom"
	"github.com/weibaohui/kom/mcp/tools"
	"github.com/weibaohui/kom/mcp/tools/metadata"
)

// GetPodResourceUsageTool 创建获取Pod资源使用情况的工具
func GetPodResourceUsageTool() mcp.Tool {
	return mcp.NewTool(
		"get_pod_resource_usage",
		mcp.WithDescription("获取Pod的资源使用情况，包括CPU和内存的请求值、限制值、可分配值和使用比例 / Get pod resource usage including CPU and memory request, limit, allocatable and usage ratio"),
		mcp.WithString("cluster", mcp.Description("运行Pod的集群 / The cluster runs the pod")),
		mcp.WithString("namespace", mcp.Description("Pod所在的命名空间 / The namespace of the pod")),
		mcp.WithString("name", mcp.Description("Pod的名称 / The name of the pod")),
		mcp.WithNumber("cache_seconds", mcp.Description("缓存时间（默认20秒） / Cache duration in seconds,default 20 seconds")),
	)
}

// GetPodResourceUsageHandler 处理获取Pod资源使用情况的请求
func GetPodResourceUsageHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 获取参数
	meta, err := metadata.ParseFromRequest(request)
	if err != nil {
		return nil, err
	}
	cacheSeconds := int32(20)
	if cacheSecondsVal, ok := request.Params.Arguments["cacheSeconds"].(float64); ok {
		cacheSeconds = int32(cacheSecondsVal)
	}
	// 获取资源使用情况
	usage := kom.Cluster(meta.Cluster).WithContext(ctx).WithCache(time.Duration(cacheSeconds) * time.Second).Namespace(meta.Namespace).Name(meta.Name).Ctl().Pod().ResourceUsage()

	return tools.TextResult(usage, meta)
}
