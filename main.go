package main

import (
	"flag"

	"github.com/weibaohui/kom/example"
	"k8s.io/klog/v2"
)

func main() {
	klog.InitFlags(nil)
	flag.Set("v", "8")
	example.Connect()
	example.Example()
}
