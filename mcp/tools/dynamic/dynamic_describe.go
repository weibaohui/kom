package dynamic

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/weibaohui/kom/kom"
)

func GetDynamicResourceDescribe() mcp.Tool {
	return mcp.NewTool(
		"describe_k8s_resource",
		mcp.WithDescription("Retrieve Kubernetes resource details by cluster, namespace, and name"),
		mcp.WithString("cluster", mcp.Description("Cluster where the resource is running")),
		mcp.WithString("namespace", mcp.Description("Namespace of the resource (optional for cluster-scoped resources)")),
		mcp.WithString("name", mcp.Description("Name of the resource")),
		mcp.WithString("group", mcp.Description("API group of the resource (optional if resourceType is provided)")),
		mcp.WithString("version", mcp.Description("API version of the resource (optional if resourceType is provided)")),
		mcp.WithString("kind", mcp.Description("Kind of the resource (optional if resourceType is provided)")),
	)
}

func GetDynamicResourceDescribeHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 获取参数
	cluster := request.Params.Arguments["cluster"].(string)
	namespace := request.Params.Arguments["namespace"].(string)
	name := request.Params.Arguments["name"].(string)

	// 获取资源类型信息
	var group, version, kind string
	if resourceType, ok := request.Params.Arguments["kind"].(string); ok && resourceType != "" {
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
		// 如果没有提供kind，使用明确指定的GVK
		group = request.Params.Arguments["group"].(string)
		version = request.Params.Arguments["version"].(string)
		kind = request.Params.Arguments["kind"].(string)
	}

	var describeResult []byte
	err := kom.Cluster(cluster).WithContext(ctx).CRD(group, version, kind).Namespace(namespace).Name(name).Describe(&describeResult).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get item [%s/%s] type of  [%s%s%s]: %v", namespace, name, group, version, kind, err)
	}

	// 构建返回结果
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: string(describeResult),
			},
		},
	}, nil
}
