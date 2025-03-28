package pod

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/weibaohui/kom/kom"
	"github.com/weibaohui/kom/mcp/tools"
	"github.com/weibaohui/kom/mcp/tools/metadata"
	"k8s.io/klog/v2"
)

// GetPodLinkedPVTool 定义PV查询工具
func GetPodLinkedPVTool() mcp.Tool {
	return mcp.NewTool(
		"get-pod-linked-pv",
		mcp.WithDescription("获取与Pod关联的PersistentVolume"),
		mcp.WithString("cluster", mcp.Description("集群名称")),
		mcp.WithString("namespace", mcp.Description("Pod所在命名空间")),
		mcp.WithString("name", mcp.Description("Pod名称")),
	)
}

// GetPodLinkedPVHandler 处理PV查询请求
func GetPodLinkedPVHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	meta, err := metadata.ParseFromRequest(req)
	if err != nil {
		klog.Errorf("解析元数据失败: %v", err)
		return nil, fmt.Errorf("解析请求失败: %v", err)
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
