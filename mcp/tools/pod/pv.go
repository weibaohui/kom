package pod

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/weibaohui/kom/kom"
	"github.com/weibaohui/kom/mcp/tools"
	"k8s.io/klog/v2"
)

// GetPodLinkedPVTool 返回一个用于查询与指定Pod关联的PersistentVolume的工具定义。
func GetPodLinkedPVTool() mcp.Tool {
	return mcp.NewTool(
		"get_k8s_pod_linked_pv",
		mcp.WithDescription("获取与Pod关联的PersistentVolume (类似命令: kubectl get pv | grep <pod-name>)"),
		mcp.WithString("cluster", mcp.Description("集群名称（使用空字符串表示默认集群）")),
		mcp.WithString("namespace", mcp.Required(), mcp.Description("Pod所在命名空间")),
		mcp.WithString("name", mcp.Required(), mcp.Description("Pod名称")),
	)
}

// GetPodLinkedPVHandler 根据请求参数查询并返回指定 Pod 关联的 PersistentVolume 列表。
func GetPodLinkedPVHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 获取参数
	ctx, meta, err := tools.ParseFromRequest(ctx, request)
	if err != nil {
		return nil, err
	}

	pvList, err := kom.Cluster(meta.Cluster).
		Namespace(meta.Namespace).
		Name(meta.Name).
		Ctl().Pod().
		LinkedPV()
	if err != nil {
		klog.Errorf("查询PV失败: %v", err)
		return nil, fmt.Errorf("查询PV失败: %v", err)
	}

	return tools.TextResult(pvList, meta)
}
