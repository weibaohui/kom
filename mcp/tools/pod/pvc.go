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

// GetPodLinkedPVCTool 定义PVC查询工具
func GetPodLinkedPVCTool() mcp.Tool {
	return mcp.NewTool(
		"get_pod_linked_pvc",
		mcp.WithDescription("获取与Pod关联的PersistentVolumeClaim"),
		mcp.WithString("cluster", mcp.Description("集群名称")),
		mcp.WithString("namespace", mcp.Description("Pod所在命名空间")),
		mcp.WithString("name", mcp.Description("Pod名称")),
	)
}

// GetPodLinkedPVCHandler 处理PVC查询请求
func GetPodLinkedPVCHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ctx, meta, err := metadata.ParseFromRequest(ctx, request, config)

	if err != nil {
		klog.Errorf("解析元数据失败: %v", err)
		return nil, fmt.Errorf("解析请求失败: %v", err)
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
