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

// UploadPodFileTool 返回一个用于将文件上传到指定Kubernetes Pod容器内的MCP工具。
// 工具支持指定目标集群、命名空间、Pod名称、容器名称、容器内目标路径及文件内容，功能类似于 kubectl cp 命令。
func UploadPodFileTool() mcp.Tool {
	return mcp.NewTool(
		"upload_file_to_k8s_pod",
		mcp.WithDescription("上传文件到Pod容器内 (类似命令: kubectl cp <local-file> <namespace>/<pod-name>:<container-path>) / Upload file to pod container"),
		mcp.WithString("cluster", mcp.Required(), mcp.Description("集群名称 （使用空字符串表示默认集群）/ Cluster name")),
		mcp.WithString("namespace", mcp.Description("命名空间 / Namespace")),
		mcp.WithString("name", mcp.Description("Pod名称 / Pod name")),
		mcp.WithString("container", mcp.Description("容器名称（必填） / Container name (required)")),
		mcp.WithString("file_path", mcp.Description("目标文件路径 / Target file path")),
		mcp.WithString("content", mcp.Description("文件内容 / File content")),
	)
}

// UploadPodFileHandler  处理文件上传到Pod
func UploadPodFileHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ctx, meta, err := metadata.ParseFromRequest(ctx, request, config)

	if err != nil {
		return nil, err
	}
	// 如果只有一个集群的时候，使用空，默认集群
	// 如果大于一个集群，没有传值，那么要返回错误
	if len(kom.Clusters().AllClusters()) > 1 && meta.Cluster == "" {
		return nil, fmt.Errorf("cluster is required, 集群名称必须设置")
	}
	if len(kom.Clusters().AllClusters()) == 1 && meta.Cluster == "" {
		meta.Cluster = kom.Clusters().DefaultCluster().ID
	}
	if kom.Clusters().GetClusterById(meta.Cluster) == nil {
		return nil, fmt.Errorf("cluster %s not found 集群不存在，请检查集群名称", meta.Cluster)
	}

	// 容器名称必填校验
	containerName := request.Params.Arguments["container"].(string)

	// 获取文件路径和内容
	filePath := request.Params.Arguments["file_path"].(string)
	content := request.Params.Arguments["content"].(string)

	klog.V(6).Infof("Uploading file to pod %s/%s container %s: path %s", meta.Namespace, meta.Name, containerName, filePath)

	// 上传文件
	err = kom.Cluster(meta.Cluster).WithContext(ctx).
		Namespace(meta.Namespace).
		Name(meta.Name).
		Ctl().Pod().
		ContainerName(containerName).
		SaveFile(filePath, content)

	if err != nil {
		return nil, fmt.Errorf("file upload failed: %v", err)
	}

	return utils.TextResult("File uploaded successfully", meta)
}
