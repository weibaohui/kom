package storageclass

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/weibaohui/kom/kom"
	"github.com/weibaohui/kom/mcp/tools"
	v1 "k8s.io/api/storage/v1"
)

// GetStorageClassPVCountTool 返回一个用于获取指定 StorageClass 下 PV 数量的工具。
// 工具支持指定集群（通过 "cluster" 参数，空字符串表示默认集群）和 StorageClass 名称（"name" 参数）。
func GetStorageClassPVCountTool() mcp.Tool {
	return mcp.NewTool(
		"get_k8s_storageclass_pv_count",
		mcp.WithDescription("获取StorageClass下的PV数量 / Get PV count of StorageClass (等同于 kubectl get pv -l \"storageclass.kubernetes.io/name=<storageclass-name>\" --no-headers | wc -l)"),
		mcp.WithString("cluster", mcp.Description("StorageClass所在的集群（使用空字符串表示默认集群） / The cluster of the StorageClass")),
		mcp.WithString("name", mcp.Description("StorageClass的名称 / The name of the StorageClass")),
	)
}

// GetStorageClassPVCountHandler 根据请求参数获取指定 Kubernetes StorageClass 关联的持久卷（PV）数量。
// 如果请求解析或查询过程中发生错误，则返回相应的错误。
//
// 返回值：
//   - *mcp.CallToolResult：包含 PV 数量的结果。
//   - error：如有错误则返回，否则为 nil。
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
