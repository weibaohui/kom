package pod

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/weibaohui/kom/kom"
	"github.com/weibaohui/kom/mcp/metadata"
	"github.com/weibaohui/kom/utils"

	networkingv1 "k8s.io/api/networking/v1"
)

// GetPodLinkedServiceTool 返回一个用于获取指定集群、命名空间和Pod名称下关联Service的工具。
func GetPodLinkedServiceTool() mcp.Tool {
	return mcp.NewTool(
		"get_k8s_pod_linked_services",
		mcp.WithDescription("获取与Pod关联的Service，通过集群、命名空间和Pod名称 (类似命令: kubectl get svc -n <namespace> -l app=<pod-label>) / Get services linked to pod by cluster, namespace and name"),
		mcp.WithString("cluster", mcp.Required(), mcp.Description("运行Pod的集群 （使用空字符串表示默认集群） / The cluster runs the pod")),
		mcp.WithString("namespace", mcp.Description("Pod所在的命名空间 / The namespace of the pod")),
		mcp.WithString("name", mcp.Description("Pod的名称 / The name of the pod")),
	)
}

// GetPodLinkedServiceHandler 处理获取关联Service的请求
func GetPodLinkedServiceHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

	return utils.TextResult(results, meta)
}

// GetPodLinkedIngressTool 返回一个用于获取指定集群、命名空间和Pod名称下关联Ingress的工具。
// 工具要求提供集群（必填，空字符串表示默认集群）、命名空间和Pod名称作为参数。
func GetPodLinkedIngressTool() mcp.Tool {
	return mcp.NewTool(
		"get_pod_linked_ingresses",
		mcp.WithDescription("获取与Pod关联的Ingress，通过集群、命名空间和Pod名称 (类似命令: kubectl get ingress -n <namespace> -o wide | grep <service-name>) / Get ingresses linked to pod by cluster, namespace and name"),
		mcp.WithString("cluster", mcp.Required(), mcp.Description("运行Pod的集群 （使用空字符串表示默认集群） / The cluster runs the pod")),
		mcp.WithString("namespace", mcp.Description("Pod所在的命名空间 / The namespace of the pod")),
		mcp.WithString("name", mcp.Description("Pod的名称 / The name of the pod")),
	)
}

// GetPodLinkedIngressHandler 处理获取关联Ingress的请求
func GetPodLinkedIngressHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

	return utils.TextResult(results, meta)
}

// GetPodLinkedEndpointsTool 返回一个工具，用于根据集群、命名空间和Pod名称获取与指定Pod关联的Endpoints。
func GetPodLinkedEndpointsTool() mcp.Tool {
	return mcp.NewTool(
		"get_pod_linked_endpoints",
		mcp.WithDescription("获取与Pod关联的Endpoints，通过集群、命名空间和Pod名称 (类似命令: kubectl get endpoints -n <namespace> | grep <pod-ip>) / Get endpoints linked to pod by cluster, namespace and name"),
		mcp.WithString("cluster", mcp.Required(), mcp.Description("运行Pod的集群 （使用空字符串表示默认集群） / The cluster runs the pod")),
		mcp.WithString("namespace", mcp.Description("Pod所在的命名空间 / The namespace of the pod")),
		mcp.WithString("name", mcp.Description("Pod的名称 / The name of the pod")),
	)
}

// GetPodLinkedEndpointsHandler 处理获取关联Endpoints的请求
func GetPodLinkedEndpointsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

	return utils.TextResult(results, meta)
}

func getTLSSecretName(ingress *networkingv1.Ingress) string {
	if len(ingress.Spec.TLS) > 0 {
		return ingress.Spec.TLS[0].SecretName
	}
	return ""
}
