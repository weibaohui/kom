package pod

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/weibaohui/kom/kom"
	v1 "k8s.io/api/core/v1"
)

// RegisterPodTools 注册Pod相关的工具到MCP服务器
func RegisterPodTools(s *server.MCPServer) {
	s.AddTool(
		GetPodTool(),
		GetPodHandler,
	)
}

// GetPodTool 创建一个查询Pod的工具
func GetPodTool() mcp.Tool {
	return mcp.NewTool(
		"get-pod",
		mcp.WithDescription("Get pod information by cluster and namespace and name"),
		mcp.WithString("cluster", mcp.Description("The cluster runs the pod")),
		mcp.WithString("namespace", mcp.Description("The namespace of the pod")),
		mcp.WithString("name", mcp.Description("The name of the pod")),
	)
}

// GetPodHandler 处理查询Pod的请求
func GetPodHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 获取参数
	cluster := request.Params.Arguments["cluster"].(string)
	namespace := request.Params.Arguments["namespace"].(string)
	name := request.Params.Arguments["name"].(string)

	// 查询Pod
	var pod v1.Pod
	err := kom.Cluster(cluster).Resource(&v1.Pod{}).Namespace(namespace).Name(name).Get(&pod).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get pod: %v", err)
	}

	// 构建返回结果
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: fmt.Sprintf("Pod Information:\n"+
					"Name: %s\n"+
					"Namespace: %s\n"+
					"Status: %s\n"+
					"Node: %s\n"+
					"IP: %s\n"+
					"Start Time: %s",
					pod.Name,
					pod.Namespace,
					string(pod.Status.Phase),
					pod.Spec.NodeName,
					pod.Status.PodIP,
					pod.Status.StartTime.String()),
			},
		},
	}, nil
}
