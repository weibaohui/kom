package pod

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/weibaohui/kom/kom"
	"github.com/weibaohui/kom/mcp/metadata"
	"github.com/weibaohui/kom/utils"

	"k8s.io/klog/v2"
)

// GetPodLinkedPVCTool 返回一个用于查询指定Pod关联的PersistentVolumeClaim的MCP工具。
// 工具名称为 "get_k8s_pod_linked_pvc"，支持指定集群、命名空间和Pod名称进行PVC检索。
func GetPodLinkedPVCTool() mcp.Tool {
	return mcp.NewTool(
		"get_k8s_pod_linked_pvc",
		mcp.WithDescription("获取与Pod关联的PersistentVolumeClaim (类似命令: kubectl get pvc -n <namespace> | grep <pod-name>)"),
		mcp.WithString("cluster", mcp.Required(), mcp.Description("集群名称（使用空字符串表示默认集群）")),
		mcp.WithString("namespace", mcp.Description("Pod所在命名空间")),
		mcp.WithString("name", mcp.Description("Pod名称")),
	)
}

// GetPodLinkedPVCHandler 处理PVC查询请求
func GetPodLinkedPVCHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 获取参数
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

	pvcList, err := kom.Cluster(meta.Cluster).
		Namespace(meta.Namespace).
		Name(meta.Name).
		Ctl().Pod().
		LinkedPVC()
	if err != nil {
		klog.Errorf("查询PVC失败: %v", err)
		return nil, fmt.Errorf("查询PVC失败: %v", err)
	}

	return utils.TextResult(pvcList, meta)
}
