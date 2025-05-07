package pod

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/weibaohui/kom/kom"
	"github.com/weibaohui/kom/mcp/tools"
	"k8s.io/klog/v2"
)

// GetPodLinkedPVCTool 定义PVC查询工具
func GetPodLinkedPVCTool() mcp.Tool {
	return mcp.NewTool(
		"get_k8s_pod_linked_pvc",
		mcp.WithDescription("获取与Pod关联的PersistentVolumeClaim (类似命令: kubectl get pvc -n <namespace> | grep <pod-name>)"),
		mcp.WithString("cluster", mcp.Description("集群名称（使用空字符串表示默认集群）")),
		mcp.WithString("namespace", mcp.Description("Pod所在命名空间")),
		mcp.WithString("name", mcp.Description("Pod名称")),
	)
}

// GetPodLinkedPVCHandler 处理PVC查询请求
func GetPodLinkedPVCHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 获取参数
	ctx, meta, err := tools.ParseFromRequest(ctx, request)
	if err != nil {
		return nil, err
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

	return tools.TextResult(pvcList, meta)
}
