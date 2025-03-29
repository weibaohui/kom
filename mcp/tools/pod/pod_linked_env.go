package pod

import (
	"context"
	"encoding/json"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/weibaohui/kom/kom"
	"github.com/weibaohui/kom/mcp/tools"
	"github.com/weibaohui/kom/mcp/tools/metadata"
	"k8s.io/klog/v2"
)

// GetPodLinkedEnvTool 创建获取Pod运行时环境变量的工具
func GetPodLinkedEnvTool() mcp.Tool {
	return mcp.NewTool(
		"get_pod_linked_env",
		mcp.WithDescription("获取Pod运行时的环境变量信息 / Get pod runtime environment variables"),
		mcp.WithString("cluster", mcp.Description("运行Pod的集群 / The cluster runs the pod")),
		mcp.WithString("namespace", mcp.Description("Pod所在的命名空间 / The namespace of the pod")),
		mcp.WithString("name", mcp.Description("Pod的名称 / The name of the pod")),
	)
}

// GetPodLinkedEnvHandler 处理获取Pod环境变量的请求
func GetPodLinkedEnvHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 获取参数
	meta, err := metadata.ParseFromRequest(request)
	if err != nil {
		return nil, err
	}

	// 获取环境变量
	envs, err := kom.Cluster(meta.Cluster).WithContext(ctx).Namespace(meta.Namespace).Name(meta.Name).Ctl().Pod().LinkedEnv()
	if err != nil {
		klog.Errorf("get pod %s/%s env error: %v", meta.Namespace, meta.Name, err)
		return nil, err
	}

	// 转换为JSON
	data, err := json.Marshal(envs)
	if err != nil {
		klog.Errorf("marshal env error: %v", err)
		return nil, err
	}

	return tools.TextResult(string(data), meta)
}
