package storageclass

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/weibaohui/kom/kom"
	"github.com/weibaohui/kom/mcp/tools"
	v1 "k8s.io/api/storage/v1"
)

// SetDefaultStorageClassTool 返回一个用于将指定 StorageClass 设置为默认的 mcp.Tool 实例。
// 该工具允许用户指定集群和 StorageClass 名称，实现等同于为 StorageClass 添加默认注解的效果。
func SetDefaultStorageClassTool() mcp.Tool {
	return mcp.NewTool(
		"set_k8s_default_storageclass",
		mcp.WithDescription("设置StorageClass为默认 ，等同于执行kubectl annotate storageclass <name> storageclass.kubernetes.io/is-default-class=true / Set StorageClass as default"),
		mcp.WithString("cluster", mcp.Description("StorageClass所在的集群 （使用空字符串表示默认集群）/ The cluster of the StorageClass")),
		mcp.WithString("name", mcp.Description("StorageClass的名称 / The name of the StorageClass")),
	)
}

// SetDefaultStorageClassHandler 处理设置指定 Kubernetes 集群中 StorageClass 为默认存储类的请求。
// 根据请求参数定位目标集群和 StorageClass，并将其标记为默认存储类。
// 成功时返回操作结果文本，失败时返回相应错误。
func SetDefaultStorageClassHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 获取参数
	ctx, meta, err := tools.ParseFromRequest(ctx, request)
	if err != nil {
		return nil, err
	}

	err = kom.Cluster(meta.Cluster).WithContext(ctx).Resource(&v1.StorageClass{}).Name(meta.Name).Ctl().StorageClass().SetDefault()
	if err != nil {
		return nil, err
	}

	return tools.TextResult("Successfully set StorageClass as default", meta)
}
