package dynamic

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/weibaohui/kom/kom"
	"github.com/weibaohui/kom/mcp/metadata"
	"github.com/weibaohui/kom/utils"

	"k8s.io/apimachinery/pkg/types"
)

// PatchDynamicResource 返回一个用于通过集群、命名空间和名称动态更新 Kubernetes 资源的 MCP 工具。该工具要求指定集群，并支持通过 JSON 补丁数据以 strategic-merge-patch 方式修改资源配置。
func PatchDynamicResource() mcp.Tool {
	return mcp.NewTool(
		"patch_k8s_resource",
		mcp.WithDescription("通过集群、命名空间和名称更新Kubernetes资源 / Patch Kubernetes resource by cluster, namespace, and name"),
		mcp.WithString("cluster", mcp.Required(), mcp.Description("运行资源的集群（使用空字符串表示默认集群）/ Cluster where the resources are running (use empty string for default cluster)")),
		mcp.WithString("namespace", mcp.Description("资源所在的命名空间（集群范围资源可选）/ Namespace of the resource (optional for cluster-scoped resources)")),
		mcp.WithString("name", mcp.Description("资源的名称 / Name of the resource")),
		mcp.WithString("group", mcp.Description("资源的API组 / API group of the resource")),
		mcp.WithString("version", mcp.Description("资源的API版本 / API version of the resource")),
		mcp.WithString("kind", mcp.Description("资源的类型 / Kind of the resource")),
		mcp.WithString("patch_data", mcp.Description("JSON补丁数据，使用application/strategic-merge-patch+json格式，用于更新资源的配置，例如：{\"spec\":{\"replicas\":5},\"metadata\":{\"labels\":{\"new-label\":\"new-value\"}}} / JSON patch data in application/strategic-merge-patch+json format for updating resource configuration")),
	)
}

func PatchDynamicResourceHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 获取资源元数据
	ctx, meta, err := metadata.ParseFromRequest(ctx, request, config)

	if err != nil {
		return nil, err
	}
	// 如果只有一个集群的时候，使用空，默认集群
	// 如果大于一个集群，没有传值，那么要返回错误
	if len(kom.Clusters().AllClusters()) > 1 && meta.Cluster == "" {
		return nil, fmt.Errorf("cluster is required, 集群名称必须设置")
	}
	if len(kom.Clusters().AllClusters()) == 1 && meta.Cluster == "" {
		meta.Cluster = kom.Clusters().DefaultCluster().ID
	}
	if kom.Clusters().GetClusterById(meta.Cluster) == nil {
		return nil, fmt.Errorf("cluster %s not found 集群不存在，请检查集群名称", meta.Cluster)
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
	return utils.TextResult(result, meta)
}
