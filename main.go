package main

import (
	"flag"

	"github.com/weibaohui/kom/example"
	"github.com/weibaohui/kom/kom_starter"
	"k8s.io/klog/v2"
)

func main() {
	klog.InitFlags(nil)
	_ = flag.Set("v", "6")
	kom_starter.Init()
	example.Example()
}
