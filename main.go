package main

import (
	"flag"

	"github.com/weibaohui/kom/example"
	"github.com/weibaohui/kom/mcp"
	"k8s.io/klog/v2"
)

func main() {
	klog.InitFlags(nil)
	flag.Set("v", "8")
	example.Connect()
	example.Example()
	mcp.RunMCPServer(9096)

}
