package pod

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/weibaohui/kom/kom"
	"github.com/weibaohui/kom/mcp/tools"
	"k8s.io/klog/v2"
)

// GetPodLinkedEnvTool 返回一个用于获取指定Pod运行时环境变量的工具。
// 该工具通过在Pod内执行Env命令，检索实际运行时的环境变量信息。适用于需要动态查看Pod当前环境变量的场景。
func GetPodLinkedEnvTool() mcp.Tool {
	return mcp.NewTool(
		"get_k8s_pod_linked_env",
		mcp.WithDescription("通过进入pod执行Env命令，获取Pod运行时的环境变量信息 (类似命令: kubectl exec -n <namespace> <pod-name> -- env) / Get pod runtime environment variables by executing Env command"),
		mcp.WithString("cluster", mcp.Description("运行Pod的集群 （使用空字符串表示默认集群） / The cluster runs the pod")),
		mcp.WithString("namespace", mcp.Description("Pod所在的命名空间 / The namespace of the pod")),
		mcp.WithString("name", mcp.Description("Pod的名称 / The name of the pod")),
	)
}

// GetPodLinkedEnvHandler 处理获取指定 Kubernetes Pod 运行时环境变量的请求。
// 根据请求参数定位集群、命名空间和 Pod 名称，通过在 Pod 内部执行 env 命令获取实际运行时环境变量。
// 返回环境变量文本结果，若获取失败则返回错误。
func GetPodLinkedEnvHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 获取参数
	ctx, meta, err := tools.ParseFromRequest(ctx, request)
	if err != nil {
		return nil, err
	}

	// 获取环境变量
	envs, err := kom.Cluster(meta.Cluster).WithContext(ctx).Namespace(meta.Namespace).Name(meta.Name).Ctl().Pod().LinkedEnv()
	if err != nil {
		klog.Errorf("get pod %s/%s env error: %v", meta.Namespace, meta.Name, err)
		return nil, err
	}

	return tools.TextResult(envs, meta)
}

// GetPodLinkedEnvFromPodYamlTool 创建一个工具，用于通过Pod的YAML定义获取环境变量信息。
func GetPodLinkedEnvFromPodYamlTool() mcp.Tool {
	return mcp.NewTool(
		"get_pod_linked_env_from_yaml",
		mcp.WithDescription("通过Pod yaml 定义 获取Pod定义中的环境变量信息 (类似命令: kubectl get pod <pod-name> -n <namespace> -o jsonpath='{.spec.containers[*].env}') / Get environment variables from pod definition"),
		mcp.WithString("cluster", mcp.Description("运行Pod的集群 （使用空字符串表示默认集群） / The cluster runs the pod")),
		mcp.WithString("namespace", mcp.Description("Pod所在的命名空间 / The namespace of the pod")),
		mcp.WithString("name", mcp.Description("Pod的名称 / The name of the pod")),
	)
}

// GetPodLinkedEnvFromPodYamlHandler 处理获取指定 Kubernetes Pod YAML 定义中环境变量的请求。
// 根据请求参数定位集群、命名空间和 Pod 名称，返回 Pod 规范中定义的环境变量列表。
func GetPodLinkedEnvFromPodYamlHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 获取参数
	ctx, meta, err := tools.ParseFromRequest(ctx, request)
	if err != nil {
		return nil, err
	}

	// 获取环境变量
	envs, err := kom.Cluster(meta.Cluster).WithContext(ctx).Namespace(meta.Namespace).Name(meta.Name).Ctl().Pod().LinkedEnvFromPod()
	if err != nil {
		klog.Errorf("get pod %s/%s env from pod error: %v", meta.Namespace, meta.Name, err)
		return nil, err
	}

	return tools.TextResult(envs, meta)
}
