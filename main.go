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

	authKey := "Authorization"
	var ctxFn = func(ctx context.Context, r *http.Request) context.Context {
		authVal := r.Header.Get(authKey)
		klog.Infof("%s: %s", authKey, authVal)
		return context.WithValue(ctx, authKey, authVal)
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
		AuthKey: authKey,
	}
	mcp.RunMCPServerWithOption(&cfg)
}
