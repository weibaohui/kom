package pod

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/weibaohui/kom/kom"
	"github.com/weibaohui/kom/mcp/tools"
	"k8s.io/klog/v2"
)

// ListPodFilesTool 返回一个用于获取指定Kubernetes Pod中某路径下文件列表的工具。
// 工具参数包括集群名称（可为空表示默认集群）、命名空间、Pod名称、容器名称和目标路径。
func ListPodFilesTool() mcp.Tool {
	return mcp.NewTool(
		"list_files_in_k8s_pod",
		mcp.WithDescription("获取Pod中指定路径下的文件列表 (类似命令: kubectl exec <pod-name> -n <namespace> -c <container> -- ls <path>) / List files in pod path"),
		mcp.WithString("cluster", mcp.Description("集群名称（使用空字符串表示默认集群）/Cluster name")),
		mcp.WithString("namespace", mcp.Description("命名空间/Namespace")),
		mcp.WithString("name", mcp.Description("Pod名称/Pod name")),
		mcp.WithString("container", mcp.Description("容器名称/Container name")),
		mcp.WithString("path", mcp.Description("目标路径/Target path")),
	)
}

// ListPodFilesHandler 处理列出指定 Kubernetes Pod 容器内指定路径下文件的请求。
// 如果路径参数为空，则返回错误。成功时返回该路径下的文件列表文本结果。
func ListPodFilesHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ctx, meta, err := tools.ParseFromRequest(ctx, request)
	if err != nil {
		return nil, err
	}

	path, _ := request.Params.Arguments["path"].(string)
	container, _ := request.Params.Arguments["container"].(string)

	if path == "" {
		return nil, fmt.Errorf("路径参数不能为空")
	}

	podCtl := kom.Cluster(meta.Cluster).
		WithContext(ctx).
		Namespace(meta.Namespace).
		Name(meta.Name).
		Ctl().Pod().
		ContainerName(container)

	files, err := podCtl.ListFiles(path)
	if err != nil {
		klog.Errorf("List files error: %v", err)
		return nil, err
	}

	return tools.TextResult(files, meta)
}

// ListAllPodFilesTool 返回一个用于获取指定Pod容器内某路径下所有文件（包括子目录中文件）列表的工具。
func ListAllPodFilesTool() mcp.Tool {
	return mcp.NewTool(
		"list_pod_all_files",
		mcp.WithDescription("获取Pod中指定路径下的所有文件列表（包含子目录）(类似命令: kubectl exec <pod-name> -n <namespace> -c <container> -- find <path>) / List all files in pod path (including subdirectories)"),
		mcp.WithString("cluster", mcp.Description("集群名称（使用空字符串表示默认集群）/Cluster name")),
		mcp.WithString("namespace", mcp.Description("命名空间/Namespace")),
		mcp.WithString("name", mcp.Description("Pod名称/Pod name")),
		mcp.WithString("container", mcp.Description("容器名称/Container name")),
		mcp.WithString("path", mcp.Description("目标路径/Target path")),
	)
}

// ListAllPodFilesHandler 递归列出指定 Kubernetes Pod 容器内路径下的所有文件。
// 
// 处理请求，解析参数后，返回目标路径及其所有子目录下的文件完整列表。路径参数不能为空。
func ListAllPodFilesHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ctx, meta, err := tools.ParseFromRequest(ctx, request)
	if err != nil {
		return nil, err
	}

	path, _ := request.Params.Arguments["path"].(string)
	container, _ := request.Params.Arguments["container"].(string)

	if path == "" {
		return nil, fmt.Errorf("路径参数不能为空")
	}

	podCtl := kom.Cluster(meta.Cluster).
		WithContext(ctx).
		Namespace(meta.Namespace).
		Name(meta.Name).
		Ctl().Pod().
		ContainerName(container)

	files, err := podCtl.ListAllFiles(path)
	if err != nil {
		klog.Errorf("List all files error: %v", err)
		return nil, err
	}

	return tools.TextResult(files, meta)
}

// DeletePodFileTool 返回一个用于删除指定Pod容器内文件的工具。该工具支持指定集群、命名空间、Pod、容器及目标文件路径，功能等同于在命令行执行 kubectl exec ... rm <path>。
func DeletePodFileTool() mcp.Tool {
	return mcp.NewTool(
		"delete_pod_file",
		mcp.WithDescription("删除Pod中的指定文件 (类似命令: kubectl exec <pod-name> -n <namespace> -c <container> -- rm <path>) / Delete file in pod"),
		mcp.WithString("cluster", mcp.Description("集群名称（使用空字符串表示默认集群）/Cluster name")),
		mcp.WithString("namespace", mcp.Description("命名空间/Namespace")),
		mcp.WithString("name", mcp.Description("Pod名称/Pod name")),
		mcp.WithString("container", mcp.Description("容器名称/Container name")),
		mcp.WithString("path", mcp.Description("目标路径/Target path")),
	)
}

// DeletePodFileHandler 处理删除 Kubernetes Pod 内指定文件的请求。
// 如果路径参数为空，则返回错误。成功时返回删除操作的输出结果。
func DeletePodFileHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ctx, meta, err := tools.ParseFromRequest(ctx, request)
	if err != nil {
		return nil, err
	}

	path, _ := request.Params.Arguments["path"].(string)
	container, _ := request.Params.Arguments["container"].(string)

	if path == "" {
		return nil, fmt.Errorf("路径参数不能为空")
	}

	podCtl := kom.Cluster(meta.Cluster).
		WithContext(ctx).
		Namespace(meta.Namespace).
		Name(meta.Name).
		Ctl().Pod().
		ContainerName(container)

	ret, err := podCtl.DeleteFile(path)
	if err != nil {
		klog.Errorf("Delete file error: %v", err)
		return nil, err
	}

	return tools.TextResult(string(ret), meta)
}
