package node

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/weibaohui/kom/kom"
	"github.com/weibaohui/kom/mcp/tools"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func ListNode() mcp.Tool {
	return mcp.NewTool(
		"list_k8s_node",
		mcp.WithDescription("获取Node列表 (类似命令 kubectl get node)"),
		mcp.WithString("cluster", mcp.Description("Node所在集群（使用空字符串表示默认集群）")),
	)
}

func ListNodeHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

	// 获取资源元数据
	ctx, meta, err := tools.ParseFromRequest(ctx, request)
	if err != nil {
		return nil, err
	}

	// 获取资源列表
	var list []*unstructured.Unstructured
	kubectl := kom.Cluster(meta.Cluster).WithContext(ctx).
		Resource(&v1.Node{}).RemoveManagedFields()

	err = kubectl.List(&list).Error
	if err != nil {
		return nil, fmt.Errorf("failed to list items type of [%s%s%s]: %v", meta.Group, meta.Version, meta.Kind, err)
	}

	// 提取name和namespace信息
	var result []map[string]string
	for _, item := range list {
		ret := map[string]string{
			"name": item.GetName(),
		}

		result = append(result, ret)
	}

	return tools.TextResult(result, meta)
}
