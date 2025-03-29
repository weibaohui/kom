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

// PodExecTool 创建执行Pod命令的工具
func PodExecTool() mcp.Tool {
	return mcp.NewTool(
		"pod_exec",
		mcp.WithDescription("在Pod内执行命令，需指定容器名称 / Execute command in pod with container name"),
		mcp.WithString("cluster", mcp.Description("集群名称 / Cluster name")),
		mcp.WithString("namespace", mcp.Description("命名空间 / Namespace")),
		mcp.WithString("name", mcp.Description("Pod名称 / Pod name")),
		mcp.WithString("container", mcp.Description("容器名称（必填） / Container name (required)")),
		mcp.WithString("command", mcp.Description("要执行的命令 / Command to execute")),
		mcp.WithArray("args", mcp.Description("命令参数列表 / Command arguments")),
	)
}

// PodExecHandler 处理Pod命令执行
func PodExecHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	meta, err := metadata.ParseFromRequest(request)
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
