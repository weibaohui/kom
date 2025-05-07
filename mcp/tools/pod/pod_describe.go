package pod

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/weibaohui/kom/kom"
	"github.com/weibaohui/kom/mcp/tools"
	v1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
)

// DescribePod 返回一个用于描述 Kubernetes Pod 资源的工具，功能类似于 kubectl describe pod 命令。
func DescribePod() mcp.Tool {
	return mcp.NewTool(
		"describe_k8s_pod",
		mcp.WithDescription("描述Pod容器组，(类似命令: kubectl describe pod -n <namespace> pod_name ) "),
		mcp.WithString("cluster", mcp.Description("Pod所在集群（使用空字符串表示默认集群）")),
		mcp.WithString("namespace", mcp.Required(), mcp.Description("Pod所在的命名空间（集群范围资源可选）")),
		mcp.WithString("name", mcp.Required(), mcp.Description("Pod名称")),
	)
}

// DescribePodHandler 处理对 Kubernetes Pod 资源的描述请求，返回类似 kubectl describe pod 的文本信息。
// 如果资源获取或描述过程中发生错误，将返回详细的错误信息。
func DescribePodHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 获取资源元数据
	ctx, meta, err := tools.ParseFromRequest(ctx, request)
	if err != nil {
		return nil, err
	}

	var describeResult []byte
	err = kom.Cluster(meta.Cluster).WithContext(ctx).
		Resource(&v1.Pod{}).
		Namespace(meta.Namespace).
		Name(meta.Name).
		RemoveManagedFields().
		Describe(&describeResult).Error
	if err != nil {
		klog.Errorf("failed to get item [%s/%s] type of  [%s%s%s]: %v", meta.Namespace, meta.Name, meta.Group, meta.Version, meta.Kind, err)
		return nil, fmt.Errorf("failed to get item [%s/%s] type of  [%s%s%s]: %v", meta.Namespace, meta.Name, meta.Group, meta.Version, meta.Kind, err)
	}
	return tools.TextResult(describeResult, meta)

}
