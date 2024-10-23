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
	_, _ = kom.Clusters().InitInCluster()
	_, _ = kom.Clusters().InitByPathWithID(defaultKubeConfig, "default")
	callbacks.RegisterDefaultCallbacks(kom.Clusters())
	kom.Clusters().Show()
}
func InitWithConfig(path string) {
	_, _ = kom.Clusters().InitInCluster()

	// 初始化kubectl 连接
	_, _ = kom.Clusters().InitByPathWithID(path, "default")
	callbacks.RegisterDefaultCallbacks(kom.Clusters())
	kom.Clusters().Show()

}
