package pod

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/weibaohui/kom/kom"
	"github.com/weibaohui/kom/mcp/tools"
	"k8s.io/klog/v2"
)

// ListPodFilesTool 创建Pod文件列表工具
func ListPodFilesTool() mcp.Tool {
	return mcp.NewTool(
		"list_files_in_k8s_pod",
		mcp.WithDescription("获取Pod中指定路径下的文件列表 (类似命令: kubectl exec <pod-name> -n <namespace> -c <container> -- ls <path>) / List files in pod path"),
		mcp.WithTitleAnnotation("List Files in Pod"),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithString("cluster", mcp.Description("集群名称（使用空字符串表示默认集群）/Cluster name")),
		mcp.WithString("namespace", mcp.Description("命名空间/Namespace")),
		mcp.WithString("name", mcp.Description("Pod名称/Pod name")),
		mcp.WithString("container", mcp.Description("容器名称/Container name")),
		mcp.WithString("path", mcp.Description("目标路径/Target path")),
	)
}

// ListPodFilesHandler 处理Pod文件列表请求
func ListPodFilesHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ctx, meta, err := tools.ParseFromRequest(ctx, request)
	if err != nil {
		return nil, err
	}

	// path, _ := request.Params.Arguments["path"].(string)
	// container, _ := request.Params.Arguments["container"].(string)
	path := request.GetString("path", "")
	container := request.GetString("container", "")

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

// ListAllPodFilesTool 创建Pod文件全量列表工具
func ListAllPodFilesTool() mcp.Tool {
	return mcp.NewTool(
		"list_pod_all_files",
		mcp.WithDescription("获取Pod中指定路径下的所有文件列表（包含子目录）(类似命令: kubectl exec <pod-name> -n <namespace> -c <container> -- find <path>) / List all files in pod path (including subdirectories)"),
		mcp.WithTitleAnnotation("List All Pod Files"),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithString("cluster", mcp.Description("集群名称（使用空字符串表示默认集群）/Cluster name")),
		mcp.WithString("namespace", mcp.Description("命名空间/Namespace")),
		mcp.WithString("name", mcp.Description("Pod名称/Pod name")),
		mcp.WithString("container", mcp.Description("容器名称/Container name")),
		mcp.WithString("path", mcp.Description("目标路径/Target path")),
	)
}

// ListAllPodFilesHandler 处理Pod文件全量列表请求
func ListAllPodFilesHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ctx, meta, err := tools.ParseFromRequest(ctx, request)
	if err != nil {
		return nil, err
	}

	// path, _ := request.Params.Arguments["path"].(string)
	// container, _ := request.Params.Arguments["container"].(string)
	path := request.GetString("path", "")
	container := request.GetString("container", "")

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

// DeletePodFileTool 创建Pod文件删除工具
func DeletePodFileTool() mcp.Tool {
	return mcp.NewTool(
		"delete_pod_file",
		mcp.WithDescription("删除Pod中的指定文件 (类似命令: kubectl exec <pod-name> -n <namespace> -c <container> -- rm <path>) / Delete file in pod"),
		mcp.WithTitleAnnotation("Delete Pod File"),
		mcp.WithDestructiveHintAnnotation(true),
		mcp.WithString("cluster", mcp.Description("集群名称（使用空字符串表示默认集群）/Cluster name")),
		mcp.WithString("namespace", mcp.Description("命名空间/Namespace")),
		mcp.WithString("name", mcp.Description("Pod名称/Pod name")),
		mcp.WithString("container", mcp.Description("容器名称/Container name")),
		mcp.WithString("path", mcp.Description("目标路径/Target path")),
	)
}

// DeletePodFileHandler 处理Pod文件删除请求
func DeletePodFileHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ctx, meta, err := tools.ParseFromRequest(ctx, request)
	if err != nil {
		return nil, err
	}

	// path, _ := request.Params.Arguments["path"].(string)
	// container, _ := request.Params.Arguments["container"].(string)
	path := request.GetString("path", "")
	container := request.GetString("container", "")

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
