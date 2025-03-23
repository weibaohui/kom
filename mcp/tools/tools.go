package tools

import (
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/weibaohui/kom/mcp/tools/metadata"
	"k8s.io/apimachinery/pkg/util/json"
)

// buildTextResult 构建标准的文本返回结果
func buildTextResult[T any](text T) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: fmt.Sprint(text),
			},
		},
	}
}

// TextResult 将任意类型转换为标准的mcp.CallToolResult
func TextResult[T any](item T, meta *metadata.ResourceMetadata) (*mcp.CallToolResult, error) {
	switch v := any(item).(type) {
	case []byte:
		return buildTextResult(v), nil
	default:
		bytes, err := json.Marshal(item)
		if err != nil {
			return nil, fmt.Errorf("failed to json marshal item [%s/%s] type of [%s%s%s]: %v",
				meta.Namespace, meta.Name, meta.Group, meta.Version, meta.Kind, err)
		}
		return buildTextResult(bytes), nil
	}
}

func ErrorResult(err error) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		IsError: true,
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: err.Error(),
			},
		},
	}
}
