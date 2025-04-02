package storageclass

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/weibaohui/kom/kom"
	"github.com/weibaohui/kom/mcp/metadata"
	"github.com/weibaohui/kom/utils"

	v1 "k8s.io/api/storage/v1"
)

func GetStorageClassPVCCountTool() mcp.Tool {
	return mcp.NewTool(
		"get_storageclass_pvc_count",
		mcp.WithDescription("获取StorageClass下的PVC数量 / Get PVC count of StorageClass"),
		mcp.WithString("cluster", mcp.Description("StorageClass所在的集群 / The cluster of the StorageClass")),
		mcp.WithString("name", mcp.Description("StorageClass的名称 / The name of the StorageClass")),
	)
}

func GetStorageClassPVCCountHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 获取参数
	ctx, meta, err := metadata.ParseFromRequest(ctx, request, config)

	if err != nil {
		return nil, err
	}

	count, err := kom.Cluster(meta.Cluster).WithContext(ctx).Resource(&v1.StorageClass{}).Name(meta.Name).Ctl().StorageClass().PVCCount()
	if err != nil {
		return nil, err
	}

	return utils.TextResult(count, meta)
}
