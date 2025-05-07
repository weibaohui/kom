package pod

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/weibaohui/kom/kom"
	"github.com/weibaohui/kom/mcp/tools"
	"github.com/weibaohui/kom/utils"
	v1 "k8s.io/api/core/v1"
)

// TopPod 返回一个用于获取 Kubernetes Pod CPU 和内存资源用量排名的工具，类似于 "kubectl top pods" 命令。
func TopPod() mcp.Tool {
	return mcp.NewTool(
		"get_k8s_top_pod",
		mcp.WithDescription("获取Pod CPU 内存 资源用量排名 列表 (类似命令 kubectl top pods -n ns)"),
		mcp.WithString("cluster", mcp.Description("运行资源的集群（使用空字符串表示默认集群）/ Cluster where the resources are running (use empty string for default cluster)")),
		mcp.WithString("namespace", mcp.Description("资源所在的命名空间（集群范围资源可选）/ Namespace of the resources (optional for cluster-scoped resources)")),
	)
}

// TopPodHandler 处理获取指定集群和命名空间下 Pod 资源使用排行的请求，返回 Pod 的 CPU 和内存使用情况排名数据。
func TopPodHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

	// 获取资源元数据
	ctx, meta, err := tools.ParseFromRequest(ctx, request)
	if err != nil {
		return nil, err
	}

	kubectl := kom.Cluster(meta.Cluster).WithContext(ctx).
		Resource(&v1.Pod{}).
		RemoveManagedFields()
	if meta.Namespace == "" {
		kubectl = kubectl.AllNamespace()
	} else {
		kubectl = kubectl.Namespace(meta.Namespace)
	}

	top, err := kubectl.Ctl().Pod().Top()
	if err != nil {
		return nil, fmt.Errorf("failed to  kubectl top pod list items type of [%s%s%s]: %v", meta.Group, meta.Version, meta.Kind, err)
	}

	return tools.TextResult(utils.ToJSON(top), meta)
}
