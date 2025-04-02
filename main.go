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
	example.Connect()
	// example.Example()
	// mcp.RunMCPServer("kom mcp server", "0.0.1", 9096)

	authKey := "username"
	authRoleKey := "role"
	var ctxFn = func(ctx context.Context, r *http.Request) context.Context {
		authKeyVal := r.Header.Get(authKey)
		authRoleVal := r.Header.Get(authRoleKey)
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
	}
	mcp.RunMCPServerWithOption(&cfg)
}
