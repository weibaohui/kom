package dynamic

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/weibaohui/kom/kom"
	"github.com/weibaohui/kom/mcp/tools"
	"k8s.io/apimachinery/pkg/types"
)

func PatchDynamicResource() mcp.Tool {
	return mcp.NewTool(
		"patch_k8s_resource",
		mcp.WithDescription("通过集群、命名空间和名称更新Kubernetes资源 / Patch Kubernetes resource by cluster, namespace, and name"),
		mcp.WithString("cluster", mcp.Description("运行资源的集群（使用空字符串表示默认集群）/ Cluster where the resources are running (use empty string for default cluster)")),
		mcp.WithString("namespace", mcp.Description("资源所在的命名空间（集群范围资源可选）/ Namespace of the resource (optional for cluster-scoped resources)")),
		mcp.WithString("name", mcp.Description("资源的名称 / Name of the resource")),
		mcp.WithString("group", mcp.Description("资源的API组 / API group of the resource")),
		mcp.WithString("version", mcp.Description("资源的API版本 / API version of the resource")),
		mcp.WithString("kind", mcp.Description("资源的类型 / Kind of the resource")),
		mcp.WithString("patch_data", mcp.Description("JSON补丁数据，使用application/strategic-merge-patch+json格式，用于更新资源的配置，例如：{\"spec\":{\"replicas\":5},\"metadata\":{\"labels\":{\"new-label\":\"new-value\"}}} / JSON patch data in application/strategic-merge-patch+json format for updating resource configuration")),
	)
}

// PatchDynamicResourceHandler 根据请求参数对指定的 Kubernetes 资源执行战略合并补丁操作。
// 若 patch_data 缺失或补丁失败，将返回详细错误信息。
// 成功时返回资源补丁结果的文本描述。
func PatchDynamicResourceHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 获取资源元数据
	ctx, meta, err := tools.ParseFromRequest(ctx, request)
	if err != nil {
		return nil, err
	}

	// 获取patch数据
	patchData, ok := request.Params.Arguments["patch_data"].(string)
	if !ok || patchData == "" {
		return nil, fmt.Errorf("patch data is required")
	}

	// 更新资源
	kubectl := kom.Cluster(meta.Cluster).WithContext(ctx).CRD(meta.Group, meta.Version, meta.Kind).Namespace(meta.Namespace)
	if meta.Namespace == "" {
		kubectl = kubectl.AllNamespace()
	}
	var item interface{}
	err = kubectl.Name(meta.Name).Patch(&item, types.StrategicMergePatchType, patchData).Error
	if err != nil {
		return nil, fmt.Errorf("failed to patch item [%s/%s] type of [%s%s%s]: %v", meta.Namespace, meta.Name, meta.Group, meta.Version, meta.Kind, err)
	}

	result := fmt.Sprintf("Successfully patched resource [%s/%s] of type [%s%s%s]", meta.Namespace, meta.Name, meta.Group, meta.Version, meta.Kind)
	return tools.TextResult(result, meta)
}
