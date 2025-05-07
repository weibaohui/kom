package pod

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/weibaohui/kom/kom"
	"github.com/weibaohui/kom/mcp/tools"
	networkingv1 "k8s.io/api/networking/v1"
)

// GetPodLinkedServiceTool 返回一个用于获取指定集群、命名空间和Pod名称下关联Service的工具。
func GetPodLinkedServiceTool() mcp.Tool {
	return mcp.NewTool(
		"get_k8s_pod_linked_services",
		mcp.WithDescription("获取与Pod关联的Service，通过集群、命名空间和Pod名称 (类似命令: kubectl get svc -n <namespace> -l app=<pod-label>) / Get services linked to pod by cluster, namespace and name"),
		mcp.WithString("cluster", mcp.Description("运行Pod的集群 （使用空字符串表示默认集群） / The cluster runs the pod")),
		mcp.WithString("namespace", mcp.Description("Pod所在的命名空间 / The namespace of the pod")),
		mcp.WithString("name", mcp.Description("Pod的名称 / The name of the pod")),
	)
}

// GetPodLinkedServiceHandler 处理获取与指定 Pod 关联的 Service 的请求，并返回相关 Service 的基本信息。
func GetPodLinkedServiceHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ctx, meta, err := tools.ParseFromRequest(ctx, request)
	if err != nil {
		return nil, err
	}

	services, err := kom.Cluster(meta.Cluster).WithContext(ctx).Namespace(meta.Namespace).Name(meta.Name).Ctl().Pod().LinkedService()
	if err != nil {
		return nil, fmt.Errorf("获取关联Service失败: %v", err)
	}

	var results []map[string]interface{}
	for _, svc := range services {
		results = append(results, map[string]interface{}{
			"name":      svc.Name,
			"namespace": svc.Namespace,
			"type":      svc.Spec.Type,
			"clusterIP": svc.Spec.ClusterIP,
		})
	}

	return tools.TextResult(results, meta)
}

// GetPodLinkedIngressTool 返回一个用于获取与指定Pod关联的Ingress资源的工具。
// 该工具通过集群、命名空间和Pod名称定位Pod，并检索与其相关联的Ingress列表。
func GetPodLinkedIngressTool() mcp.Tool {
	return mcp.NewTool(
		"get_pod_linked_ingresses",
		mcp.WithDescription("获取与Pod关联的Ingress，通过集群、命名空间和Pod名称 (类似命令: kubectl get ingress -n <namespace> -o wide | grep <service-name>) / Get ingresses linked to pod by cluster, namespace and name"),
		mcp.WithString("cluster", mcp.Description("运行Pod的集群 （使用空字符串表示默认集群） / The cluster runs the pod")),
		mcp.WithString("namespace", mcp.Description("Pod所在的命名空间 / The namespace of the pod")),
		mcp.WithString("name", mcp.Description("Pod的名称 / The name of the pod")),
	)
}

// GetPodLinkedIngressHandler 根据请求获取与指定 Pod 关联的 Ingress 资源，并返回其关键信息。
func GetPodLinkedIngressHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ctx, meta, err := tools.ParseFromRequest(ctx, request)
	if err != nil {
		return nil, err
	}

	ingresses, err := kom.Cluster(meta.Cluster).WithContext(ctx).Namespace(meta.Namespace).Name(meta.Name).Ctl().Pod().LinkedIngress()
	if err != nil {
		return nil, fmt.Errorf("获取关联Ingress失败: %v", err)
	}

	var results []map[string]interface{}
	for _, ingress := range ingresses {
		results = append(results, map[string]interface{}{
			"name":      ingress.Name,
			"namespace": ingress.Namespace,
			"hosts":     ingress.Spec.Rules[0].Host,
			"tlsSecret": getTLSSecretName(ingress),
		})
	}

	return tools.TextResult(results, meta)
}

// GetPodLinkedEndpointsTool 返回一个用于获取与指定Pod关联的Endpoints的工具。
// 该工具通过集群、命名空间和Pod名称定位Pod，并检索与其关联的所有Endpoints。
func GetPodLinkedEndpointsTool() mcp.Tool {
	return mcp.NewTool(
		"get_pod_linked_endpoints",
		mcp.WithDescription("获取与Pod关联的Endpoints，通过集群、命名空间和Pod名称 (类似命令: kubectl get endpoints -n <namespace> | grep <pod-ip>) / Get endpoints linked to pod by cluster, namespace and name"),
		mcp.WithString("cluster", mcp.Description("运行Pod的集群 （使用空字符串表示默认集群） / The cluster runs the pod")),
		mcp.WithString("namespace", mcp.Description("Pod所在的命名空间 / The namespace of the pod")),
		mcp.WithString("name", mcp.Description("Pod的名称 / The name of the pod")),
	)
}

// GetPodLinkedEndpointsHandler 处理请求并返回与指定 Pod 关联的 Endpoints 信息。
func GetPodLinkedEndpointsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ctx, meta, err := tools.ParseFromRequest(ctx, request)
	if err != nil {
		return nil, err
	}

	endpoints, err := kom.Cluster(meta.Cluster).WithContext(ctx).Namespace(meta.Namespace).Name(meta.Name).Ctl().Pod().LinkedEndpoints()
	if err != nil {
		return nil, fmt.Errorf("获取关联Endpoints失败: %v", err)
	}

	var results []map[string]interface{}
	for _, ep := range endpoints {
		results = append(results, map[string]interface{}{
			"name":      ep.Name,
			"namespace": ep.Namespace,
			"subsets":   ep.Subsets,
		})
	}

	return tools.TextResult(results, meta)
}

// getTLSSecretName 返回 Ingress 资源中第一个 TLS 配置的 Secret 名称，如无则返回空字符串。
func getTLSSecretName(ingress *networkingv1.Ingress) string {
	if len(ingress.Spec.TLS) > 0 {
		return ingress.Spec.TLS[0].SecretName
	}
	return ""
}
