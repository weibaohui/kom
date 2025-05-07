package pod

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/weibaohui/kom/kom"
	"github.com/weibaohui/kom/mcp/tools"
	"k8s.io/klog/v2"
)

// ExecTool 返回一个用于在 Kubernetes Pod 容器内执行命令的工具。
// 工具支持指定目标集群、命名空间、Pod 名称、容器名称、命令及其参数，功能类似于 kubectl exec。
func ExecTool() mcp.Tool {
	return mcp.NewTool(
		"run_command_in_k8s_pod",
		mcp.WithDescription("在Pod内执行命令，需指定容器名称 (类似命令: kubectl exec -n <namespace> <pod-name> -c <container-name> -- <command> [args...]) / Execute command in pod with container name"),
		mcp.WithString("cluster", mcp.Description("集群名称 （使用空字符串表示默认集群）/ Cluster name")),
		mcp.WithString("namespace", mcp.Required(), mcp.Description("命名空间 / Namespace")),
		mcp.WithString("name", mcp.Required(), mcp.Description("Pod名称 / Pod name")),
		mcp.WithString("container", mcp.Description("容器名称（必填） / Container name (required)")),
		mcp.WithString("command", mcp.Required(), mcp.Description("要执行的命令 / Command to execute")),
		mcp.WithArray("args", mcp.Description("命令参数列表 / Command arguments")),
	)
}

// ExecHandler 在指定的 Kubernetes Pod 容器内执行命令，并返回命令输出结果。
// 如果命令执行失败，则返回错误。
func ExecHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ctx, meta, err := tools.ParseFromRequest(ctx, request)
	if err != nil {
		return nil, err
	}

	// 容器名称必填校验
	containerName := request.Params.Arguments["container"].(string)

	// 解析命令参数
	var argsVal []string
	args, ok := request.Params.Arguments["args"].([]interface{})
	if ok {
		// 将 []interface{} 转换为 []string
		for _, arg := range args {
			if str, ok := arg.(string); ok {
				argsVal = append(argsVal, str)
			}
		}
	}
	command := request.Params.Arguments["command"].(string)

	klog.V(6).Infof("Executing command in pod %s/%s container %s: %v %v", meta.Namespace, meta.Name, containerName, command, argsVal)

	// 执行命令
	var execResult string
	err = kom.Cluster(meta.Cluster).WithContext(ctx).
		Namespace(meta.Namespace).
		Name(meta.Name).
		Ctl().Pod().
		ContainerName(containerName).
		Command(command, argsVal...).
		Execute(&execResult).Error

	if err != nil {
		return nil, fmt.Errorf("command execution failed: %v", err)
	}

	return tools.TextResult(execResult, meta)
}
