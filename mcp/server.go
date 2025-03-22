package mcp

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/weibaohui/kom/mcp/tools/pod"
	"k8s.io/klog/v2"
)

func RunMCPServer(port int) {
	// 创建一个新的 MCP 服务器
	s := server.NewMCPServer(
		"kom mcp server",
		"0.0.1",
		server.WithResourceCapabilities(true, true),
		server.WithPromptCapabilities(true),
		server.WithLogging(),
	)

	// 添加一个加法计算工具
	s.AddTool(
		mcp.NewTool(
			"add-numbers",
			mcp.WithDescription("Add two numbers together"),
			mcp.WithNumber("number1", mcp.Description("First number to add")),
			mcp.WithNumber("number2", mcp.Description("Second number to add")),
		),
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			num1 := request.Params.Arguments["number1"].(float64)
			num2 := request.Params.Arguments["number2"].(float64)
			result := num1 + num2

			return &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.TextContent{
						Type: "text",
						Text: fmt.Sprintf("The sum of %.2f and %.2f is %.2f", num1, num2, result),
					},
				},
			}, nil
		},
	)

	// 注册Pod相关工具
	pod.RegisterPodTools(s)

	// 创建 SSE 服务器
	sseServer := server.NewSSEServer(s)

	// 启动服务器
	err := sseServer.Start(fmt.Sprintf(":%d", port))
	if err != nil {
		klog.Errorf("MCP Server error: %v\n", err)
	}
}
