package kom_starter

import (
	"os"
	"path/filepath"

	"github.com/weibaohui/kom/callbacks"
	"github.com/weibaohui/kom/kom"
	"github.com/weibaohui/kom/kom/doc"
	"k8s.io/client-go/util/homedir"
)

func Init() {

	defaultKubeConfig := os.Getenv("KUBECONFIG")
	if defaultKubeConfig == "" {
		defaultKubeConfig = filepath.Join(homedir.HomeDir(), ".kube", "config")
	}
	// 初始化kubectl 连接
	kom.InitConnection(defaultKubeConfig)
	callbacks.RegisterDefaultCallbacks()
	doc.Instance()
}
func InitWithConfig(path string) {

	// 初始化kubectl 连接
	kom.InitConnection(path)
	callbacks.RegisterDefaultCallbacks()
	doc.Instance()
}
