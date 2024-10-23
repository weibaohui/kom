package kom_starter

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
	kom.Clusters().SetCallbackRegisterFunc(callbacks.RegisterDefaultCallbacks)
	_, _ = kom.Clusters().InitInCluster()
	_, _ = kom.Clusters().InitByPathWithID(defaultKubeConfig, "default")
	kom.Clusters().Show()
}
func InitWithConfig(path string) {
	_, _ = kom.Clusters().InitInCluster()

	kom.Clusters().SetCallbackRegisterFunc(callbacks.RegisterDefaultCallbacks)
	// 初始化kubectl 连接
	_, _ = kom.Clusters().InitByPathWithID(path, "default")
	kom.Clusters().Show()

}
