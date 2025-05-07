package pod

import (
	"context"
	"fmt"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/weibaohui/kom/kom"
	"github.com/weibaohui/kom/mcp/metadata"
	"github.com/weibaohui/kom/utils"
)

// GetPodResourceUsageTool 创建获取Pod资源使用情况的工具
func GetPodResourceUsageTool() mcp.Tool {
	return mcp.NewTool(
		"get_pod_resource_usage",
		mcp.WithDescription("获取Pod的资源使用情况，包括CPU和内存的请求值、限制值、可分配值和使用比例 (类似命令: kubectl describe pod <pod-name> -n <namespace>, kubectl top pod <pod-name> -n <namespace>) / Get pod resource usage including CPU and memory request, limit, allocatable and usage ratio"),
		mcp.WithString("cluster", mcp.Required(), mcp.Description("运行Pod的集群 （使用空字符串表示默认集群） / The cluster runs the pod")),
		mcp.WithString("namespace", mcp.Description("Pod所在的命名空间 / The namespace of the pod")),
		mcp.WithString("name", mcp.Description("Pod的名称 / The name of the pod")),
		mcp.WithNumber("cache_seconds", mcp.Description("缓存时间（默认20秒） / Cache duration in seconds,default 20 seconds")),
	)
}

// GetPodResourceUsageHandler 处理获取Pod资源使用情况的请求
func GetPodResourceUsageHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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
	// 获取资源使用情况
	usage, err := kom.Cluster(meta.Cluster).WithContext(ctx).WithCache(time.Duration(cacheSeconds) * time.Second).Namespace(meta.Namespace).Name(meta.Name).Ctl().Pod().ResourceUsage()
	if err != nil {
		return nil, err
	}
	return utils.TextResult(usage, meta)
}
