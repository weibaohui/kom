package main

import (
	"context"
	"flag"
	"net/http"

	"github.com/mark3labs/mcp-go/server"
	"github.com/weibaohui/kom/example"
	"github.com/weibaohui/kom/mcp"
	"github.com/weibaohui/kom/mcp/metadata"
	"k8s.io/klog/v2"
)

func main() {
	klog.InitFlags(nil)
	flag.Set("v", "8")
	go example.Connect()
	// example.Example()

	// 一行代码启动模式
	// mcp.RunMCPServer("kom mcp server", "0.0.1", 9096)

	// 复杂参数配置模式
	// SSE调用时，可以在MCP client 请求中，在header中注入认证信息
	// kom执行时，如果注册了callback，那么可以在callback中获取到ctx，从ctx上可以拿到注入的认证信息
	// 有了认证信息，就可以在callback中，进行权限的逻辑控制了
	authKey := "username"
	authRoleKey := "role"
	var ctxFn = func(ctx context.Context, r *http.Request) context.Context {
		authKeyVal := r.Header.Get(authKey)
		authRoleVal := r.Header.Get(authRoleKey)
		authKeyVal = "admin"
		authRoleVal = "admin-role"
		klog.Infof("%s: %s", authKey, authKeyVal)
		klog.Infof("%s: %s", authRoleKey, authRoleKey)
		ctx = context.WithValue(ctx, authKey, authKeyVal)
		ctx = context.WithValue(ctx, authRoleKey, authRoleVal)

		return ctx
	}
	cfg := metadata.ServerConfig{
		Name:    "kom mcp server",
		Version: "0.0.1",
		Port:    9096,
		ServerOptions: []server.ServerOption{
			server.WithResourceCapabilities(false, false),
			server.WithPromptCapabilities(false),
			server.WithLogging(),
		},
		SSEOption: []server.SSEOption{
			server.WithSSEContextFunc(ctxFn),
		},
		AuthKey:     authKey,
		AuthRoleKey: authRoleKey,
		Mode:        metadata.MCPServerModeSSE, // 开启STDIO 或者 SSE
	}
	mcp.RunMCPServerWithOption(&cfg)

}
