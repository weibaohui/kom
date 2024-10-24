package kom_starter

import (
	"os"
	"path/filepath"

	_ "github.com/weibaohui/kom/callbacks"
	"github.com/weibaohui/kom/kom"
	"k8s.io/client-go/util/homedir"
)

func Init() {

	defaultKubeConfig := os.Getenv("KUBECONFIG")
	if defaultKubeConfig == "" {
		defaultKubeConfig = filepath.Join(homedir.HomeDir(), ".kube", "config")
	}
	_, _ = kom.Clusters().RegisterInCluster()
	_, _ = kom.Clusters().RegisterByPathWithID(defaultKubeConfig, "default")
	kom.Clusters().Show()
}
func InitWithConfig(path string) {
	_, _ = kom.Clusters().RegisterInCluster()

	// 初始化kubectl 连接
	_, _ = kom.Clusters().RegisterByPathWithID(path, "default")
	kom.Clusters().Show()

}
