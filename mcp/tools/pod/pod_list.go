package pod

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/weibaohui/kom/kom"
	"github.com/weibaohui/kom/mcp/tools"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// ListPod 返回一个用于获取 Kubernetes Pod 列表的工具实例，支持按集群、命名空间和字段选择器筛选。
func ListPod() mcp.Tool {
	return mcp.NewTool(
		"list_k8s_pod",
		mcp.WithDescription("获取Pod列表 (类似命令 kubectl get pods -n ns)"),
		mcp.WithString("cluster", mcp.Description("运行资源的集群（使用空字符串表示默认集群）/ Cluster where the resources are running (use empty string for default cluster)")),
		mcp.WithString("namespace", mcp.Required(), mcp.Description("资源所在的命名空间（集群范围资源可选）/ Namespace of the resources (optional for cluster-scoped resources)")),
		mcp.WithString("fieldSelector", mcp.Description("用于过滤资源的字段选择器（例如：metadata.name=test-deploy）/ Field selector to filter resources (e.g. metadata.name=test-deploy)")),
	)
}

// ListPodHandler 处理列出 Kubernetes Pod 的请求，根据请求参数返回指定集群和命名空间下的 Pod 名称及命名空间信息。
// 如果未指定命名空间，则查询所有命名空间；支持通过 fieldSelector 过滤 Pod。
// 返回包含 Pod 名称和命名空间的文本结果，或在出错时返回错误。
func ListPodHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

	// 获取资源元数据
	ctx, meta, err := tools.ParseFromRequest(ctx, request)
	if err != nil {
		return nil, err
	}
	fieldSelector, _ := request.Params.Arguments["fieldSelector"].(string)

	// 获取资源列表
	var list []*unstructured.Unstructured
	kubectl := kom.Cluster(meta.Cluster).WithContext(ctx).
		Resource(&v1.Pod{}).Namespace(meta.Namespace).RemoveManagedFields()
	if meta.Namespace == "" {
		kubectl = kubectl.AllNamespace()
	}
	if fieldSelector != "" {
		kubectl = kubectl.WithFieldSelector(fieldSelector)
	}
	err = kubectl.List(&list).Error
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
