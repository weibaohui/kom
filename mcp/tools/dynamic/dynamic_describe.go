package dynamic

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/weibaohui/kom/kom"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/json"
)

func GetDynamicResource() mcp.Tool {
	return mcp.NewTool(
		"get_k8s_resource",
		mcp.WithDescription("Retrieve Kubernetes resource details by cluster, namespace, and name"),
		mcp.WithString("cluster", mcp.Description("Cluster where the resource is running")),
		mcp.WithString("namespace", mcp.Description("Namespace of the resource")),
		mcp.WithString("name", mcp.Description("Name of the resource")),
		mcp.WithString("group", mcp.Description("API group of the resource (e.g., apps, batch)")),
		mcp.WithString("version", mcp.Description("API version of the resource (e.g., v1, v1beta1)")),
		mcp.WithString("kind", mcp.Description("Kind of the resource (e.g., Deployment, Pod)")),
	)
}

func GetDynamicResourceHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 获取参数
	cluster := request.Params.Arguments["cluster"].(string)
	namespace := request.Params.Arguments["namespace"].(string)
	name := request.Params.Arguments["name"].(string)

	group := request.Params.Arguments["group"].(string)
	version := request.Params.Arguments["version"].(string)
	kind := request.Params.Arguments["kind"].(string)

	var item unstructured.Unstructured
	err := kom.Cluster(cluster).WithContext(ctx).CRD(group, version, kind).Namespace(namespace).Name(name).Get(&item).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get item [%s/%s] type of  [%s%s%s]: %v", namespace, name, group, version, kind, err)
	}
	bytes, err := json.Marshal(item)
	if err != nil {
		return nil, fmt.Errorf("failed to json marshal item [%s/%s] type of  [%s%s%s]: %v", namespace, name, group, version, kind, err)
	}
	// 构建返回结果
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: string(bytes),
			},
		},
	}, nil
}
