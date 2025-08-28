package example

import (
	"os"
	"path/filepath"

	"github.com/weibaohui/kom/callbacks"
	"github.com/weibaohui/kom/kom"
	"k8s.io/client-go/util/homedir"
)

func Connect() {
	callbacks.RegisterInit()

	defaultKubeConfig := os.Getenv("KUBECONFIG")
	if defaultKubeConfig == "" {
		defaultKubeConfig = filepath.Join(homedir.HomeDir(), ".kube", "config")
	}

	// 配置 EKS 集群信息
	// config := aws.EKSAuthConfig{
	// 	AccessKey:       "XXX",                     // AWS Access Key ID
	// 	SecretAccessKey: "yyy", // AWS Secret Access Key
	// 	Region:          "us-east-1",                                // AWS 区域
	// 	ClusterName:     "k8m",                                      // EKS 集群名称
	// }
	//
	// _, _ = kom.Clusters().RegisterAWSCluster(config)

	_, _ = kom.Clusters().RegisterByPathWithID(defaultKubeConfig, "default")
	kom.Clusters().Show()

}
