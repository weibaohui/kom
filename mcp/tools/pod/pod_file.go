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

// UploadPodFileTool 创建上传文件到Pod的工具
func UploadPodFileTool() mcp.Tool {
	return mcp.NewTool(
		"upload_file_to_pod",
		mcp.WithDescription("上传文件到Pod容器内 / Upload file to pod container"),
		mcp.WithString("cluster", mcp.Description("集群名称 / Cluster name")),
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
