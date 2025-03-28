package pod

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/weibaohui/kom/kom"
	"github.com/weibaohui/kom/mcp/tools"
	"github.com/weibaohui/kom/mcp/tools/metadata"
	"k8s.io/klog/v2"
)

// FileOperationTool 创建文件操作工具
func FileOperationTool() mcp.Tool {
	return mcp.NewTool(
		"pod_file_operations",
		mcp.WithDescription("Pod文件操作(列表/删除)/Pod file operations (list/delete)"),
		mcp.WithString("cluster", mcp.Description("集群名称/Cluster name")),
		mcp.WithString("namespace", mcp.Description("命名空间/Namespace")),
		mcp.WithString("name", mcp.Description("Pod名称/Pod name")),
		mcp.WithString("container", mcp.Description("容器名称/Container name")),
		mcp.WithString("path", mcp.Description("目标路径/Target path")),
		mcp.WithString("operation", mcp.Description("操作类型(list/listAll/delete)/Operation type")),
	)
}

// FileOperationHandler 处理文件操作请求
func FileOperationHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	meta, err := metadata.ParseFromRequest(request)
	if err != nil {
		return nil, err
	}

	// 获取操作参数
	operation, _ := request.Params.Arguments["operation"].(string)
	path, _ := request.Params.Arguments["path"].(string)
	container, _ := request.Params.Arguments["container"].(string)

	// 路径校验
	if path == "" {
		return nil, fmt.Errorf("路径参数不能为空")
	}

	podCtl := kom.Cluster(meta.Cluster).
		WithContext(ctx).
		Namespace(meta.Namespace).
		Name(meta.Name).
		Ctl().Pod().
		ContainerName(container)

	var result interface{}
	switch strings.ToLower(operation) {
	case "list":
		files, err := podCtl.ListFiles(path)
		if err != nil {
			klog.Errorf("List files error: %v", err)
			return nil, err
		}
		result = files
	case "listAll":
		files, err := podCtl.ListAllFiles(path)
		if err != nil {
			klog.Errorf("List all files error: %v", err)
			return nil, err
		}
		result = files
	case "delete":
		ret, err := podCtl.DeleteFile(path)
		if err != nil {
			klog.Errorf("Delete file error: %v", err)
			return nil, err
		}
		result = string(ret)
	default:
		return nil, fmt.Errorf("不支持的操作类型: %s", operation)
	}

	return tools.TextResult(result, meta)
}
