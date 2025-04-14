package pod

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/weibaohui/kom/kom"
	"github.com/weibaohui/kom/mcp/metadata"
	"github.com/weibaohui/kom/utils"

	"k8s.io/klog/v2"
)

// ListPodFilesTool 创建Pod文件列表工具
func ListPodFilesTool() mcp.Tool {
	return mcp.NewTool(
		"list_pod_files",
		mcp.WithDescription("获取Pod中指定路径下的文件列表/List files in pod path"),
		mcp.WithString("cluster", mcp.Description("集群名称/Cluster name")),
		mcp.WithString("namespace", mcp.Description("命名空间/Namespace")),
		mcp.WithString("name", mcp.Description("Pod名称/Pod name")),
		mcp.WithString("container", mcp.Description("容器名称/Container name")),
		mcp.WithString("path", mcp.Description("目标路径/Target path")),
	)
}

// ListPodFilesHandler 处理Pod文件列表请求
func ListPodFilesHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ctx, meta, err := metadata.ParseFromRequest(ctx, request, config)

	if err != nil {
		return nil, err
	}
	// 如果只有一个集群的时候，使用空，默认集群
	// 如果大于一个集群，没有传值，那么要返回错误
	if len(kom.Clusters().AllClusters()) > 1 && meta.Cluster == "" {
		return nil, fmt.Errorf("cluster is required, 集群名称必须设置")
	}
	if kom.Clusters().GetClusterById(meta.Cluster) == nil {
		return nil, fmt.Errorf("cluster %s not found 集群不存在，请检查集群名称", meta.Cluster)
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

	return utils.TextResult(files, meta)
}

// ListAllPodFilesTool 创建Pod文件全量列表工具
func ListAllPodFilesTool() mcp.Tool {
	return mcp.NewTool(
		"list_pod_all_files",
		mcp.WithDescription("获取Pod中指定路径下的所有文件列表（包含子目录）/List all files in pod path (including subdirectories)"),
		mcp.WithString("cluster", mcp.Description("集群名称/Cluster name")),
		mcp.WithString("namespace", mcp.Description("命名空间/Namespace")),
		mcp.WithString("name", mcp.Description("Pod名称/Pod name")),
		mcp.WithString("container", mcp.Description("容器名称/Container name")),
		mcp.WithString("path", mcp.Description("目标路径/Target path")),
	)
}

// ListAllPodFilesHandler 处理Pod文件全量列表请求
func ListAllPodFilesHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ctx, meta, err := metadata.ParseFromRequest(ctx, request, config)

	if err != nil {
		return nil, err
	}
	// 如果只有一个集群的时候，使用空，默认集群
	// 如果大于一个集群，没有传值，那么要返回错误
	if len(kom.Clusters().AllClusters()) > 1 && meta.Cluster == "" {
		return nil, fmt.Errorf("cluster is required, 集群名称必须设置")
	}
	if kom.Clusters().GetClusterById(meta.Cluster) == nil {
		return nil, fmt.Errorf("cluster %s not found 集群不存在，请检查集群名称", meta.Cluster)
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

	return utils.TextResult(files, meta)
}

// DeletePodFileTool 创建Pod文件删除工具
func DeletePodFileTool() mcp.Tool {
	return mcp.NewTool(
		"delete_pod_file",
		mcp.WithDescription("删除Pod中的指定文件/Delete file in pod"),
		mcp.WithString("cluster", mcp.Description("集群名称/Cluster name")),
		mcp.WithString("namespace", mcp.Description("命名空间/Namespace")),
		mcp.WithString("name", mcp.Description("Pod名称/Pod name")),
		mcp.WithString("container", mcp.Description("容器名称/Container name")),
		mcp.WithString("path", mcp.Description("目标路径/Target path")),
	)
}

// DeletePodFileHandler 处理Pod文件删除请求
func DeletePodFileHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ctx, meta, err := metadata.ParseFromRequest(ctx, request, config)

	if err != nil {
		return nil, err
	}
	// 如果只有一个集群的时候，使用空，默认集群
	// 如果大于一个集群，没有传值，那么要返回错误
	if len(kom.Clusters().AllClusters()) > 1 && meta.Cluster == "" {
		return nil, fmt.Errorf("cluster is required, 集群名称必须设置")
	}
	if kom.Clusters().GetClusterById(meta.Cluster) == nil {
		return nil, fmt.Errorf("cluster %s not found 集群不存在，请检查集群名称", meta.Cluster)
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

	return utils.TextResult(string(ret), meta)
}
