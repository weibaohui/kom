package main

import (
	"context"
	"flag"
	"net/http"

	mcp2 "github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/weibaohui/kom/example"
	"github.com/weibaohui/kom/mcp"
	"github.com/weibaohui/kom/utils"
	"k8s.io/klog/v2"
)

// main 初始化并启动一个带有认证和自定义钩子的MCP服务器。
// 
// 该函数设置日志级别，启动示例连接，定义认证信息注入与日志钩子，
// 并以SSE模式运行MCP服务器。支持通过HTTP请求头注入认证信息，
// 并在工具调用及错误发生时进行详细日志记录。
func main() {
	klog.InitFlags(nil)
	flag.Set("v", "6")
	go example.Connect()
	// example.Example()

	// 一行代码启动模式
	// mcp.RunMCPServer("kom mcp server", "0.0.1", 9096)

	// 复杂参数配置模式
	// SSE调用时，可以在MCP client 请求中，在header中注入认证信息
	// kom执行时，如果注册了callback，那么可以在callback中获取到ctx，从ctx上可以拿到注入的认证信息
	// 有了认证信息，就可以在callback中，进行权限的逻辑控制了
	authKey := "username"
	var ctxFn = func(ctx context.Context, r *http.Request) context.Context {
		authKeyVal := r.Header.Get(authKey)
		klog.Infof("%s: %s", authKey, authKeyVal)
		ctx = context.WithValue(ctx, authKey, authKeyVal)

		return ctx
	}

	var actFn = func(ctx context.Context, id any, request *mcp2.CallToolRequest, result *mcp2.CallToolResult) {
		// 记录工具调用请求
		klog.V(6).Infof("CallToolRequest: %v", utils.ToJSON(request))
		klog.V(6).Infof("CallToolResult: %v", utils.ToJSON(result))
	}

	var errFn = func(ctx context.Context, id any, method mcp2.MCPMethod, message any, err error) {
		if request, ok := message.(*mcp2.CallToolRequest); ok {
			klog.V(6).Infof("CallToolRequest: %v", utils.ToJSON(request))
			klog.V(6).Infof("CallTool message: %v", utils.ToJSON(message))
		}
	}
	hooks := &server.Hooks{
		OnError:         []server.OnErrorHookFunc{errFn},
		OnAfterCallTool: []server.OnAfterCallToolFunc{actFn},
	}

	cfg := mcp.ServerConfig{
		Name:    "kom mcp server",
		Version: "0.0.1",
		Port:    9096,
		ServerOptions: []server.ServerOption{
			server.WithResourceCapabilities(false, false),
			server.WithPromptCapabilities(false),
			server.WithLogging(),
			server.WithHooks(hooks),
		},
		SSEOption: []server.SSEOption{
			server.WithSSEContextFunc(ctxFn),
		},
		AuthKey: authKey,
		Mode:    mcp.ServerModeSSE, // 开启STDIO 或者 SSE
	}
	mcp.RunMCPServerWithOption(&cfg)

}
