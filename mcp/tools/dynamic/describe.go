package dynamic

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/weibaohui/kom/kom"
	"github.com/weibaohui/kom/mcp/tools"
)

// GetDynamicResourceDescribe 返回一个用于获取指定 Kubernetes 资源详细信息的工具。
// 该工具支持通过集群、命名空间、名称、API 组、版本和类型动态检索资源描述。
func GetDynamicResourceDescribe() mcp.Tool {
	return mcp.NewTool(
		"describe_k8s_resource",
		mcp.WithDescription("通过集群、命名空间和名称获取Kubernetes资源详情 / Retrieve Kubernetes resource details by cluster, namespace, and name"),
		mcp.WithString("cluster", mcp.Description("运行资源的集群（使用空字符串表示默认集群）/ Cluster where the resources are running (use empty string for default cluster)")),
		mcp.WithString("namespace", mcp.Description("资源所在的命名空间（集群范围资源可选）/ Namespace of the resource (optional for cluster-scoped resources)")),
		mcp.WithString("name", mcp.Description("资源的名称 / Name of the resource")),
		mcp.WithString("group", mcp.Description("资源的API组 / API group of the resource")),
		mcp.WithString("version", mcp.Description("资源的API版本 / API version of the resource")),
		mcp.WithString("kind", mcp.Description("资源的类型 / Kind of the resource")),
	)
}

// GetDynamicResourceDescribeHandler 根据请求参数动态获取指定 Kubernetes 资源的详细描述信息。
// 如果资源获取或描述失败，返回相应的错误信息。
//
// 返回值：
//   - *mcp.CallToolResult：包含资源描述内容的结果。
//   - error：操作过程中发生的错误。
func GetDynamicResourceDescribeHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 获取资源元数据
	ctx, meta, err := tools.ParseFromRequest(ctx, request)
	if err != nil {
		return nil, err
	}

	var describeResult []byte
	err = kom.Cluster(meta.Cluster).WithContext(ctx).CRD(meta.Group, meta.Version, meta.Kind).Namespace(meta.Namespace).Name(meta.Name).RemoveManagedFields().Describe(&describeResult).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get item [%s/%s] type of  [%s%s%s]: %v", meta.Namespace, meta.Name, meta.Group, meta.Version, meta.Kind, err)
	}
	return tools.TextResult(describeResult, meta)

}
