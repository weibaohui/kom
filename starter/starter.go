package starter

import (
	"os"
	"path/filepath"

	"github.com/weibaohui/kom/callbacks"
	"github.com/weibaohui/kom/kom"
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

}
