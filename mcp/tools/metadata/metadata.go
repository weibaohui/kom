package metadata

import (
	"fmt"

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
func ParseFromRequest(request mcp.CallToolRequest) (*ResourceMetadata, error) {
	// 验证必要参数
	// 获取cluster参数，如果不存在则使用默认值空字符串
	cluster := ""
	if clusterVal, ok := request.Params.Arguments["cluster"].(string); ok {
		cluster = clusterVal
	}

	// 获取name参数，如果不存在则返回错误
	name := ""
	if nameVal, ok := request.Params.Arguments["name"].(string); ok {
		name = nameVal
	}

	// 获取命名空间参数（可选，支持集群级资源）
	namespace := ""
	if ns, ok := request.Params.Arguments["namespace"].(string); ok {
		namespace = ns
	}

	// 获取资源类型信息
	var group, version, kind string
	if resourceType, ok := request.Params.Arguments["kind"].(string); ok && resourceType != "" {
		// 如果提供了resourceType，从type.go获取资源信息
		if info, exists := GetResourceInfo(resourceType); exists {
			// 优先使用用户指定的GVK，如果未指定则使用默认值
			group = getStringParam(request, "group", info.Group)
			version = getStringParam(request, "version", info.Version)
			kind = getStringParam(request, "kind", info.Kind)
		}
	}

	// 如果没有通过resourceType获取到信息，则使用直接指定的GVK
	if group == "" {
		group = getStringParam(request, "group", "")
	}
	if version == "" {
		version = getStringParam(request, "version", "")
	}
	if kind == "" {
		kind = getStringParam(request, "kind", "")
	}

	// 验证GVK参数
	// 某些情况下，Group和Version可能为空，但Kind不能为空
	// 例如，对于ClusterRole和ClusterRoleBinding，Group和Version可能为空，但Kind为"ClusterRole"和"ClusterRoleBinding"
	// 因此，我们需要验证Kind是否为空
	if kind == "" {
		return nil, fmt.Errorf("missing or invalid GVK parameters")
	}

	return &ResourceMetadata{
		Cluster:   cluster,
		Namespace: namespace,
		Name:      name,
		Group:     group,
		Version:   version,
		Kind:      kind,
	}, nil
}

// getStringParam 从请求参数中获取字符串值，如果不存在或无效则返回默认值
func getStringParam(request mcp.CallToolRequest, key, defaultValue string) string {
	if value, ok := request.Params.Arguments[key].(string); ok && value != "" {
		return value
	}
	return defaultValue
}
