package metadata

import (
	"context"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
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

// IsNamespaced 判断资源是否为命名空间级别
func IsNamespaced(resourceType string) bool {
	resourceType = strings.ToLower(resourceType)
	if info, exists := resourceMap[resourceType]; exists {
		return info.Namespaced
	}
	return false
}

// ParseFromRequest 从请求中解析资源元数据
func ParseFromRequest(ctx context.Context, request mcp.CallToolRequest, serverConfig *ServerConfig) (context.Context, *ResourceMetadata, error) {
	newCtx := context.Background()
	newCtx := context.Background()
	if serverConfig != nil {
		authVal, ok := ctx.Value(serverConfig.AuthKey).(string)
		if !ok {
			authVal = ""
		}

		authRoleVal, ok := ctx.Value(serverConfig.AuthRoleKey).(string)
		if !ok {
			authRoleVal = ""
		}

		newCtx = context.WithValue(ctx, serverConfig.AuthKey, authVal)
		newCtx = context.WithValue(newCtx, serverConfig.AuthRoleKey, authRoleVal)

		// 用 newCtx 传递下去
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

	return newCtx, &ResourceMetadata{
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
