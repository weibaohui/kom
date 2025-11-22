package pod

import (
	"context"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/weibaohui/kom/kom"
	"github.com/weibaohui/kom/mcp/tools"
)

// GetPodResourceUsageTool 创建获取Pod资源使用情况的工具
func GetPodResourceUsageTool() mcp.Tool {
	return mcp.NewTool(
		"get_k8s_pod_resource_usage",
		mcp.WithDescription("获取Pod的资源使用情况，包括CPU和内存的请求值、限制值、可分配值和使用比例 (类似命令: kubectl describe pod <pod-name> -n <namespace>, kubectl top pod <pod-name> -n <namespace>) / Get pod resource usage including CPU and memory request, limit, allocatable and usage ratio"),
		mcp.WithString("cluster", mcp.Description("运行Pod的集群 （使用空字符串表示默认集群） / The cluster runs the pod")),
		mcp.WithString("namespace", mcp.Required(), mcp.Description("Pod所在的命名空间 / The namespace of the pod")),
		mcp.WithString("name", mcp.Required(), mcp.Description("Pod的名称 / The name of the pod")),
		mcp.WithNumber("cache_seconds", mcp.Description("缓存时间（默认20秒） / Cache duration in seconds,default 20 seconds")),
	)
}

// GetPodResourceUsageHandler 处理获取Pod资源使用情况的请求
func GetPodResourceUsageHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 获取参数
	ctx, meta, err := tools.ParseFromRequest(ctx, request)
	if err != nil {
		return nil, err
	}

	cacheSeconds := int32(request.GetInt("cacheSeconds", 20))
	// 获取资源使用情况
	usage, err := kom.Cluster(meta.Cluster).WithContext(ctx).WithCache(time.Duration(cacheSeconds) * time.Second).Namespace(meta.Namespace).Name(meta.Name).Ctl().Pod().ResourceUsage(kom.DenominatorLimit)
	if err != nil {
		return nil, err
	}
	return tools.TextResult(usage, meta)
}
