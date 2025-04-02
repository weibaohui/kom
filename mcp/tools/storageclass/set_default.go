package storageclass

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/weibaohui/kom/kom"
	"github.com/weibaohui/kom/mcp/metadata"
	"github.com/weibaohui/kom/utils"

	v1 "k8s.io/api/storage/v1"
)

func SetDefaultStorageClassTool() mcp.Tool {
	return mcp.NewTool(
		"set_default_storageclass",
		mcp.WithDescription("设置StorageClass为默认 / Set StorageClass as default"),
		mcp.WithString("cluster", mcp.Description("StorageClass所在的集群 / The cluster of the StorageClass")),
		mcp.WithString("name", mcp.Description("StorageClass的名称 / The name of the StorageClass")),
	)
}

func SetDefaultStorageClassHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 获取参数
	ctx, meta, err := metadata.ParseFromRequest(ctx, request, config)

	if err != nil {
		return nil, err
	}

	err = kom.Cluster(meta.Cluster).WithContext(ctx).Resource(&v1.StorageClass{}).Name(meta.Name).Ctl().StorageClass().SetDefault()
	if err != nil {
		return nil, err
	}

	return utils.TextResult("Successfully set StorageClass as default", meta)
}
