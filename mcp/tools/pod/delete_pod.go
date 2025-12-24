package pod

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/weibaohui/kom/kom"
	"github.com/weibaohui/kom/mcp/tools"
	v1 "k8s.io/api/core/v1"
)

// DeletePodTool 创建删除Pod的工具
func DeletePodTool() mcp.Tool {
	return mcp.NewTool(
		"delete_k8s_pod",
		mcp.WithDescription("删除指定的Pod (类似命令: kubectl delete pod <pod-name> -n <namespace>) / Delete specified Pod"),
		mcp.WithTitleAnnotation("Delete Pod"),
		mcp.WithDestructiveHintAnnotation(true),
		mcp.WithString("cluster", mcp.Description("运行Pod的集群（使用空字符串表示默认集群）/ Cluster where the Pod is running (use empty string for default cluster)")),
		mcp.WithString("namespace", mcp.Required(), mcp.Description("Pod所在的命名空间 / Namespace of the Pod")),
		mcp.WithString("name", mcp.Required(), mcp.Description("Pod的名称 / Name of the Pod")),
		mcp.WithBoolean("force", mcp.Description("强制删除Pod / Force delete the Pod")),
	)
}

// DeletePodHandler 处理删除Pod的请求
func DeletePodHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 解析请求参数
	ctx, meta, err := tools.ParseFromRequest(ctx, request)
	if err != nil {
		return nil, err
	}

	// 构建kubectl实例
	kubectl := kom.Cluster(meta.Cluster).WithContext(ctx).Resource(&v1.Pod{}).Namespace(meta.Namespace).Name(meta.Name)

	// 执行删除操作
	force := request.GetBool("force", false)
	if force {
		err = kubectl.ForceDelete().Error
	} else {
		err = kubectl.Delete().Error
	}

	if err != nil {
		return nil, fmt.Errorf("failed to delete pod [%s/%s]: %v", meta.Namespace, meta.Name, err)
	}

	result := fmt.Sprintf("Successfully deleted pod [%s/%s]", meta.Namespace, meta.Name)
	return tools.TextResult(result, meta)
}