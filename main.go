package main

import (
	"flag"

	"github.com/weibaohui/kom/example"
	"github.com/weibaohui/kom/starter"
	"k8s.io/klog/v2"
)

func main() {
	klog.InitFlags(nil)
	flag.Set("v", "2")
	starter.Init()
	example.Example()
}
