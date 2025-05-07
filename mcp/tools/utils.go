package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/weibaohui/kom/kom"
)

var resourceMap = map[string]ResourceInfo{
	// 命名空间级别资源
	"pod":                            {Group: "", Version: "v1", Kind: "Pod", Namespaced: true},
	"deployment":                     {Group: "apps", Version: "v1", Kind: "Deployment", Namespaced: true},
	"statefulset":                    {Group: "apps", Version: "v1", Kind: "StatefulSet", Namespaced: true},
	"daemonset":                      {Group: "apps", Version: "v1", Kind: "DaemonSet", Namespaced: true},
	"replicaset":                     {Group: "apps", Version: "v1", Kind: "ReplicaSet", Namespaced: true},
	"service":                        {Group: "", Version: "v1", Kind: "Service", Namespaced: true},
	"configmap":                      {Group: "", Version: "v1", Kind: "ConfigMap", Namespaced: true},
	"secret":                         {Group: "", Version: "v1", Kind: "Secret", Namespaced: true},
	"ingress":                        {Group: "networking.k8s.io", Version: "v1", Kind: "Ingress", Namespaced: true},
	"networkpolicy":                  {Group: "networking.k8s.io", Version: "v1", Kind: "NetworkPolicy", Namespaced: true},
	"role":                           {Group: "rbac.authorization.k8s.io", Version: "v1", Kind: "Role", Namespaced: true},
	"rolebinding":                    {Group: "rbac.authorization.k8s.io", Version: "v1", Kind: "RoleBinding", Namespaced: true},
	"serviceaccount":                 {Group: "", Version: "v1", Kind: "ServiceAccount", Namespaced: true},
	"persistentvolumeclaim":          {Group: "", Version: "v1", Kind: "PersistentVolumeClaim", Namespaced: true},
	"horizontalpodautoscaler":        {Group: "autoscaling", Version: "v2", Kind: "HorizontalPodAutoscaler", Namespaced: true},
	"cronjob":                        {Group: "batch", Version: "v1", Kind: "CronJob", Namespaced: true},
	"job":                            {Group: "batch", Version: "v1", Kind: "Job", Namespaced: true},
	"node":                           {Group: "", Version: "v1", Kind: "Node", Namespaced: false},
	"namespace":                      {Group: "", Version: "v1", Kind: "Namespace", Namespaced: false},
	"persistentvolume":               {Group: "", Version: "v1", Kind: "PersistentVolume", Namespaced: false},
	"clusterrole":                    {Group: "rbac.authorization.k8s.io", Version: "v1", Kind: "ClusterRole", Namespaced: false},
	"clusterrolebinding":             {Group: "rbac.authorization.k8s.io", Version: "v1", Kind: "ClusterRoleBinding", Namespaced: false},
	"storageclass":                   {Group: "storage.k8s.io", Version: "v1", Kind: "StorageClass", Namespaced: false},
	"customresourcedefinition":       {Group: "apiextensions.k8s.io", Version: "v1", Kind: "CustomResourceDefinition", Namespaced: false},
	"mutatingwebhookconfiguration":   {Group: "admissionregistration.k8s.io", Version: "v1", Kind: "MutatingWebhookConfiguration", Namespaced: false},
	"validatingwebhookconfiguration": {Group: "admissionregistration.k8s.io", Version: "v1", Kind: "ValidatingWebhookConfiguration", Namespaced: false},
}

// GetResourceInfo 根据资源类型字符串返回资源信息
func GetResourceInfo(resourceType string) (ResourceInfo, bool) {
	resourceType = strings.ToLower(resourceType)
	if info, exists := resourceMap[resourceType]; exists {
		return info, true
	}
	return ResourceInfo{}, false
}

// IsNamespaced 判断指定资源类型是否属于命名空间作用域。
func IsNamespaced(resourceType string) bool {
	resourceType = strings.ToLower(resourceType)
	if info, exists := resourceMap[resourceType]; exists {
		return info.Namespaced
	}
	return false
}

// ParseFromRequest 从请求参数中提取并校验 Kubernetes 资源的元数据信息。
// 返回包含认证信息的新上下文、资源元数据结构体，以及错误信息（如参数缺失或集群校验失败）。
// 若存在多个集群且未指定集群，则返回错误；仅有一个集群时自动填充默认集群。
func ParseFromRequest(ctx context.Context, request mcp.CallToolRequest) (context.Context, *ResourceMetadata, error) {
	newCtx := context.Background()
	if authKey != "" {
		authVal, ok := ctx.Value(authKey).(string)
		if !ok {
			authVal = ""
		}
		newCtx = context.WithValue(newCtx, authKey, authVal)
	}

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

	meta := &ResourceMetadata{
		Cluster:   cluster,
		Namespace: namespace,
		Name:      name,
		Group:     group,
		Version:   version,
		Kind:      kind,
	}

	// 如果只有一个集群的时候，使用空，默认集群
	// 如果大于一个集群，没有传值，那么要返回错误
	if len(kom.Clusters().AllClusters()) > 1 && meta.Cluster == "" {
		return nil, nil, fmt.Errorf("cluster is required, 集群名称必须设置")
	}
	if len(kom.Clusters().AllClusters()) == 1 && meta.Cluster == "" {
		meta.Cluster = kom.Clusters().DefaultCluster().ID
		return newCtx, meta, nil
	}
	if meta.Cluster != "" && kom.Clusters().GetClusterById(meta.Cluster) == nil {
		return nil, nil, fmt.Errorf("cluster %s not found 集群不存在，请检查集群名称", meta.Cluster)
	}
	return newCtx, meta, nil
}

// getStringParam 返回请求参数中指定键的字符串值，不存在或为空时返回默认值。
func getStringParam(request mcp.CallToolRequest, key, defaultValue string) string {
	if value, ok := request.Params.Arguments[key].(string); ok && value != "" {
		return value
	}
	return defaultValue
}

// buildTextResult 返回包含指定文本内容的标准 CallToolResult 结果。
func buildTextResult(text string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: text,
			},
		},
	}
}

// TextResult 将任意类型的数据转换为标准的 mcp.CallToolResult。
// 若输入为字节切片，则作为文本内容返回；若为字符串切片，则每个字符串作为独立文本内容返回；其他类型将被序列化为 JSON 字符串后作为文本内容返回。
// 若序列化失败，返回包含资源元信息的错误。
func TextResult[T any](item T, meta *ResourceMetadata) (*mcp.CallToolResult, error) {
	switch v := any(item).(type) {
	case []byte:
		return buildTextResult(string(v)), nil
	case []string:
		var contents []mcp.Content
		for _, s := range v {
			contents = append(contents, mcp.TextContent{
				Type: "text",
				Text: s,
			})
		}
		return &mcp.CallToolResult{Content: contents}, nil
	default:
		bytes, err := json.Marshal(item)
		if err != nil {
			return nil, fmt.Errorf("failed to json marshal item [%s/%s] type of [%s%s%s]: %v",
				meta.Namespace, meta.Name, meta.Group, meta.Version, meta.Kind, err)
		}
		return buildTextResult(string(bytes)), nil
	}
}

// ErrorResult 根据提供的错误信息构建一个标记为错误的 CallToolResult，内容为错误文本。
func ErrorResult(err error) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		IsError: true,
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: err.Error(),
			},
		},
	}
}
