package ns

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/weibaohui/kom/kom"
	"github.com/weibaohui/kom/mcp/tools"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// ListNamespace 返回一个用于获取 Kubernetes 命名空间列表的工具实例。
// 工具名称为 "list_k8s_namespace"，可通过 "cluster" 参数指定目标集群，留空则使用默认集群。
func ListNamespace() mcp.Tool {
	return mcp.NewTool(
		"list_k8s_namespace",
		mcp.WithDescription("获取命名空间列表"),
		mcp.WithString("cluster", mcp.Description("运行资源的集群（使用空字符串表示默认集群）/ Cluster where the resources are running (use empty string for default cluster)")),
	)
}

// ListNamespaceHandler 处理命名空间列表请求，返回指定集群中的所有命名空间名称及其命名空间字段（如有）。
// 如果请求解析或资源获取失败，则返回相应错误。
func ListNamespaceHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

	// 获取资源元数据
	ctx, meta, err := tools.ParseFromRequest(ctx, request)
	if err != nil {
		return nil, err
	}

	// 获取资源列表
	var list []*unstructured.Unstructured

	err = kom.Cluster(meta.Cluster).
		Resource(&v1.Namespace{}).
		WithContext(ctx).
		RemoveManagedFields().List(&list).Error
	if err != nil {
		return nil, fmt.Errorf("failed to list items type of [%s%s%s]: %v", meta.Group, meta.Version, meta.Kind, err)
	}

	// 提取name和namespace信息
	var result []map[string]string
	for _, item := range list {
		ret := map[string]string{
			"name": item.GetName(),
		}
		if item.GetNamespace() != "" {
			ret["namespace"] = item.GetNamespace()
		}

		result = append(result, ret)
	}

	return tools.TextResult(result, meta)
}
