package cluster

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/weibaohui/kom/kom"
	"github.com/weibaohui/kom/mcp/tools"
)

func ListClusters() mcp.Tool {
	return mcp.NewTool(
		"list_k8s_clusters",
		mcp.WithDescription("列出所有已注册的Kubernetes集群（可用集群、可操作集群） / List all registered Kubernetes clusters"),
		mcp.WithTitleAnnotation("List Clusters"),
		mcp.WithReadOnlyHintAnnotation(true),
	)
}

func ListClustersHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

	// 获取所有已注册的集群名称
	clusters := kom.Clusters().AllClusters()

	// 提取集群名称和详细信息
	var result []map[string]interface{}
	for clusterName, cluster := range clusters {
		clusterInfo := map[string]interface{}{
			"name":    clusterName,
			"host":    cluster.Config.Host,
			"version": "unknown",
		}
		
		if cluster.GetServerVersion() != nil {
			clusterInfo["version"] = cluster.GetServerVersion().GitVersion
		}
		
		result = append(result, clusterInfo)
	}

	return tools.TextResult(result, nil)
}

func RegisterCluster() mcp.Tool {
	return mcp.NewTool(
		"register_k8s_cluster",
		mcp.WithDescription("动态注册新的Kubernetes集群 / Dynamically register a new Kubernetes cluster"),
		mcp.WithTitleAnnotation("Register Cluster"),
		mcp.WithDestructiveHintAnnotation(true),
		mcp.WithString("cluster_id", mcp.Description("集群的唯一标识符 / Unique identifier for the cluster")),
		mcp.WithString("kubeconfig_path", mcp.Description("kubeconfig文件路径 / Path to the kubeconfig file")),
		mcp.WithString("kubeconfig_content", mcp.Description("kubeconfig内容（与kubeconfig_path二选一）/ Kubeconfig content (alternative to kubeconfig_path)")),
		mcp.WithBoolean("is_default", mcp.Description("是否设置为默认集群 / Whether to set as default cluster")),
	)
}

func RegisterClusterHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	clusterID := request.GetString("cluster_id", "")
	if clusterID == "" {
		return nil, fmt.Errorf("cluster_id is required")
	}

	kubeconfigPath := request.GetString("kubeconfig_path", "")
	kubeconfigContent := request.GetString("kubeconfig_content", "")
	isDefault := request.GetBool("is_default", false)

	var err error
	if kubeconfigContent != "" {
		// 使用kubeconfig内容
		_, err = kom.Clusters().RegisterByStringWithID(kubeconfigContent, clusterID)
	} else if kubeconfigPath != "" {
		// 使用kubeconfig文件路径
		_, err = kom.Clusters().RegisterByPathWithID(kubeconfigPath, clusterID)
	} else {
		return nil, fmt.Errorf("either kubeconfig_path or kubeconfig_content must be provided")
	}

	if err != nil {
		return nil, fmt.Errorf("failed to register cluster %s: %w", clusterID, err)
	}

	result := map[string]interface{}{
		"message":    fmt.Sprintf("Successfully registered cluster %s", clusterID),
		"cluster_id": clusterID,
		"is_default": isDefault,
	}

	return tools.TextResult(result, nil)
}

func UnregisterCluster() mcp.Tool {
	return mcp.NewTool(
		"unregister_k8s_cluster",
		mcp.WithDescription("注销Kubernetes集群 / Unregister a Kubernetes cluster"),
		mcp.WithTitleAnnotation("Unregister Cluster"),
		mcp.WithDestructiveHintAnnotation(true),
		mcp.WithString("cluster_id", mcp.Description("要注销的集群标识符 / Cluster identifier to unregister")),
	)
}

func UnregisterClusterHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	clusterID := request.GetString("cluster_id", "")
	if clusterID == "" {
		return nil, fmt.Errorf("cluster_id is required")
	}

	// 检查集群是否存在
	if kom.Clusters().GetClusterById(clusterID) == nil {
		return nil, fmt.Errorf("cluster %s not found", clusterID)
	}

	// 注销集群
	kom.Clusters().RemoveClusterById(clusterID)

	result := map[string]interface{}{
		"message":    fmt.Sprintf("Successfully unregistered cluster %s", clusterID),
		"cluster_id": clusterID,
	}

	return tools.TextResult(result, nil)
}
