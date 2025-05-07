package storageclass

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/weibaohui/kom/kom"
	"github.com/weibaohui/kom/mcp/tools"
	v1 "k8s.io/api/storage/v1"
)

// GetStorageClassPVCCountTool 返回一个用于获取指定StorageClass下PVC数量的MCP工具。
// 该工具接受“cluster”（集群名称，空字符串表示默认集群）和“name”（StorageClass名称）两个参数，用于统计对应StorageClass关联的PVC数量。
func GetStorageClassPVCCountTool() mcp.Tool {
	return mcp.NewTool(
		"get_k8s_storageclass_pvc_count",
		mcp.WithDescription("获取StorageClass下的PVC数量 / Get PVC count of StorageClass (等同于 kubectl get pvc -l \"storageclass.kubernetes.io/name=<storageclass-name>\" --no-headers | wc -l)"),
		mcp.WithString("cluster", mcp.Description("StorageClass所在的集群（使用空字符串表示默认集群） / The cluster of the StorageClass")),
		mcp.WithString("name", mcp.Description("StorageClass的名称 / The name of the StorageClass")),
	)
}

// GetStorageClassPVCCountHandler 处理获取指定 Kubernetes StorageClass 关联 PVC 数量的请求。
// 根据请求中的集群和 StorageClass 名称参数，返回该 StorageClass 下 PVC 的数量。
func GetStorageClassPVCCountHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 获取参数
	ctx, meta, err := tools.ParseFromRequest(ctx, request)
	if err != nil {
		return nil, err
	}

	count, err := kom.Cluster(meta.Cluster).WithContext(ctx).Resource(&v1.StorageClass{}).Name(meta.Name).Ctl().StorageClass().PVCCount()
	if err != nil {
		return nil, err
	}

	return tools.TextResult(count, meta)
}
