package storageclass

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/weibaohui/kom/kom"
	"github.com/weibaohui/kom/mcp/tools"
	v1 "k8s.io/api/storage/v1"
)

func GetStorageClassPVCountTool() mcp.Tool {
	return mcp.NewTool(
		"get_k8s_storageclass_pv_count",
		mcp.WithDescription("获取StorageClass下的PV数量 / Get PV count of StorageClass (等同于 kubectl get pv -l \"storageclass.kubernetes.io/name=<storageclass-name>\" --no-headers | wc -l)"),
		mcp.WithTitleAnnotation("Get StorageClass PV Count"),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithString("cluster", mcp.Description("StorageClass所在的集群（使用空字符串表示默认集群） / The cluster of the StorageClass")),
		mcp.WithString("name", mcp.Description("StorageClass的名称 / The name of the StorageClass")),
	)
}

func GetStorageClassPVCountHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 获取参数
	ctx, meta, err := tools.ParseFromRequest(ctx, request)
	if err != nil {
		return nil, err
	}

	count, err := kom.Cluster(meta.Cluster).WithContext(ctx).Resource(&v1.StorageClass{}).Name(meta.Name).Ctl().StorageClass().PVCount()
	if err != nil {
		return nil, err
	}

	return tools.TextResult(count, meta)
}
