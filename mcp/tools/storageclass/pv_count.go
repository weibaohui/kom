package storageclass

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/weibaohui/kom/kom"
	"github.com/weibaohui/kom/mcp/metadata"
	"github.com/weibaohui/kom/utils"

	v1 "k8s.io/api/storage/v1"
)

// GetStorageClassPVCountTool 返回一个用于获取指定StorageClass下PV数量的MCP工具。
// 该工具要求提供StorageClass所在集群（可为空表示默认集群）和StorageClass名称作为参数。
func GetStorageClassPVCountTool() mcp.Tool {
	return mcp.NewTool(
		"get_k8s_storageclass_pv_count",
		mcp.WithDescription("获取StorageClass下的PV数量 / Get PV count of StorageClass (等同于 kubectl get pv -l \"storageclass.kubernetes.io/name=<storageclass-name>\" --no-headers | wc -l)"),
		mcp.WithString("cluster", mcp.Required(), mcp.Description("StorageClass所在的集群（使用空字符串表示默认集群） / The cluster of the StorageClass")),
		mcp.WithString("name", mcp.Description("StorageClass的名称 / The name of the StorageClass")),
	)
}

func GetStorageClassPVCountHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

	count, err := kom.Cluster(meta.Cluster).WithContext(ctx).Resource(&v1.StorageClass{}).Name(meta.Name).Ctl().StorageClass().PVCount()
	if err != nil {
		return nil, err
	}

	return utils.TextResult(count, meta)
}
