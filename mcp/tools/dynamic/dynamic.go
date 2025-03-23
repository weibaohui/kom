package dynamic

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/weibaohui/kom/kom"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/json"
)

func GetDynamicResource() mcp.Tool {
	return mcp.NewTool(
		"get_k8s_resource",
		mcp.WithDescription("Retrieve Kubernetes resource details by cluster, namespace, and name"),
		mcp.WithString("cluster", mcp.Description("Cluster where the resource is running")),
		mcp.WithString("namespace", mcp.Description("Namespace of the resource (optional for cluster-scoped resources)")),
		mcp.WithString("name", mcp.Description("Name of the resource")),
		mcp.WithString("resourceType", mcp.Description("Type of the resource (e.g., pod, deployment)")),
		mcp.WithString("group", mcp.Description("API group of the resource (optional if resourceType is provided)")),
		mcp.WithString("version", mcp.Description("API version of the resource (optional if resourceType is provided)")),
		mcp.WithString("kind", mcp.Description("Kind of the resource (optional if resourceType is provided)")),
	)
}

func GetDynamicResourceHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 获取参数
	cluster := request.Params.Arguments["cluster"].(string)
	namespace := request.Params.Arguments["namespace"].(string)
	name := request.Params.Arguments["name"].(string)

	// 获取资源类型信息
	var group, version, kind string
	if resourceType, ok := request.Params.Arguments["resourceType"].(string); ok && resourceType != "" {
		// 如果提供了resourceType，从type.go获取资源信息
		if info, exists := GetResourceInfo(resourceType); exists {
			// 如果用户没有明确指定GVK，使用从resourceType获取的值
			if g, ok := request.Params.Arguments["group"].(string); !ok || g == "" {
				group = info.Group
			} else {
				group = g
			}
			if v, ok := request.Params.Arguments["version"].(string); !ok || v == "" {
				version = info.Version
			} else {
				version = v
			}
			if k, ok := request.Params.Arguments["kind"].(string); !ok || k == "" {
				kind = info.Kind
			} else {
				kind = k
			}
		} else {
			return nil, fmt.Errorf("unknown resource type: %s", resourceType)
		}
	} else {
		// 如果没有提供resourceType，使用明确指定的GVK
		group = request.Params.Arguments["group"].(string)
		version = request.Params.Arguments["version"].(string)
		kind = request.Params.Arguments["kind"].(string)
	}

	var item unstructured.Unstructured
	err := kom.Cluster(cluster).WithContext(ctx).CRD(group, version, kind).Namespace(namespace).Name(name).Get(&item).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get item [%s/%s] type of  [%s%s%s]: %v", namespace, name, group, version, kind, err)
	}
	bytes, err := json.Marshal(item)
	if err != nil {
		return nil, fmt.Errorf("failed to json marshal item [%s/%s] type of  [%s%s%s]: %v", namespace, name, group, version, kind, err)
	}
	// 构建返回结果
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: string(bytes),
			},
		},
	}, nil
}
