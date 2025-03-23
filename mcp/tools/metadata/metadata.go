package metadata

import (
	"github.com/mark3labs/mcp-go/mcp"
)

// ResourceMetadata 封装资源的元数据信息
type ResourceMetadata struct {
	Cluster   string
	Namespace string
	Name      string
	Group     string
	Version   string
	Kind      string
}

// ParseFromRequest 从请求中解析资源元数据
func ParseFromRequest(request mcp.CallToolRequest) *ResourceMetadata {
	// 获取基本参数
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
		}
	}

	// 如果没有提供kind，使用明确指定的GVK
	if len(group) == 0 {
		group = request.Params.Arguments["group"].(string)
	}
	if len(version) == 0 {
		version = request.Params.Arguments["version"].(string)
	}
	if len(kind) == 0 {
		kind = request.Params.Arguments["kind"].(string)
	}

	return &ResourceMetadata{
		Cluster:   cluster,
		Namespace: namespace,
		Name:      name,
		Group:     group,
		Version:   version,
		Kind:      kind,
	}
}
