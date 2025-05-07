package pod

import (
	"context"
	"io"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/weibaohui/kom/kom"
	"github.com/weibaohui/kom/mcp/tools"
	utils2 "github.com/weibaohui/kom/utils"

	v1 "k8s.io/api/core/v1"
)

// GetPodLogsTool 创建一个获取Pod日志的工具
func GetPodLogsTool() mcp.Tool {
	return mcp.NewTool(
		"get_k8s_pod_logs",
		mcp.WithDescription("获取Pod日志，通过集群、命名空间和名称，可限制返回行数 (类似命令: kubectl logs [-f] [-p] [-c container] [-n namespace] <pod-name> [--tail=N]) / Get pod logs by cluster, namespace and name with tail lines limit"),
		mcp.WithString("cluster", mcp.Description("运行Pod的集群 （使用空字符串表示默认集群） （使用空字符串表示默认集群）/ The cluster runs the pod")),
		mcp.WithString("namespace", mcp.Required(), mcp.Description("Pod所在的命名空间 / The namespace of the pod")),
		mcp.WithString("name", mcp.Required(), mcp.Description("Pod的名称 / The name of the pod")),
		mcp.WithString("container", mcp.Description("Pod中容器的名称(如果Pod中有多个容器则必须指定,只有一个容器时可以为空) / Name of the container in the pod (must be specified if there are more than one container in Pod, only one container could use empty string)")),
		mcp.WithNumber("tail", mcp.Description("显示日志末尾的行数(默认100行) / Number of lines from the end of the logs to show (default 100)")),
		mcp.WithBoolean("previous", mcp.Description("是否获取上一个容器的日志(默认false) / Whether to get logs from the previous container (default false)")),
	)
}

// GetPodLogsHandler 处理获取Pod日志的请求
func GetPodLogsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 获取参数
	ctx, meta, err := tools.ParseFromRequest(ctx, request)
	if err != nil {
		return nil, err
	}

	tailLines := int64(100)
	if tailLinesVal, ok := request.Params.Arguments["tail"].(float64); ok {
		tailLines = int64(tailLinesVal)
	}
	containerName := ""
	if containerNameVal, ok := request.Params.Arguments["container"].(string); ok {
		containerName = containerNameVal
	}
	var stream io.ReadCloser
	opt := &v1.PodLogOptions{}
	opt.TailLines = utils2.Ptr(tailLines)
	// 设置是否获取上一个容器的日志
	if previous, ok := request.Params.Arguments["previous"].(bool); ok && previous {
		opt.Previous = true
	}
	err = kom.Cluster(meta.Cluster).WithContext(ctx).Namespace(meta.Namespace).Name(meta.Name).Ctl().Pod().ContainerName(containerName).GetLogs(&stream, opt).Error
	if err != nil {
		return nil, err
	}
	// 读取所有日志内容
	var logs []byte
	logs, err = io.ReadAll(stream)
	if err != nil {
		return nil, err
	}
	return tools.TextResult(string(logs), meta)
}
