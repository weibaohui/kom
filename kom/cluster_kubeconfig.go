package kom

import (
	"fmt"
	"strings"

	"k8s.io/client-go/tools/clientcmd"
)

// RegisterByPath 通过kubeconfig文件路径注册集群
// 参数:
//   - path: kubeconfig文件的路径
//
// 返回值:
//   - *Kubectl: 成功时返回 Kubectl 实例，用于操作集群
//   - error: 失败时返回错误信息
//
// 此函数会读取指定路径的kubeconfig文件，并使用服务器地址作为集群ID
func (c *ClusterInstances) RegisterByPath(path string) (*Kubectl, error) {
	if strings.TrimSpace(path) == "" {
		return nil, fmt.Errorf("RegisterByPath: empty path")
	}
	config, err := clientcmd.BuildConfigFromFlags("", path)
	if err != nil {
		return nil, fmt.Errorf("RegisterByPath Error %s %v", path, err)
	}
	return c.RegisterByConfig(config)
}

// RegisterByString 通过kubeconfig文件的string内容进行注册
// 参数:
//   - str: kubeconfig文件的字符串内容
//
// 返回值:
//   - *Kubectl: 成功时返回 Kubectl 实例，用于操作集群
//   - error: 失败时返回错误信息
//
// 此函数会解析kubeconfig字符串内容，并使用服务器地址作为集群ID
func (c *ClusterInstances) RegisterByString(str string) (*Kubectl, error) {
	config, err := clientcmd.Load([]byte(str))
	if err != nil {
		return nil, fmt.Errorf("RegisterByString Error,content=:\n%s\n,err:%v", str, err)
	}

	clientConfig := clientcmd.NewDefaultClientConfig(*config, &clientcmd.ConfigOverrides{})
	restConfig, err := clientConfig.ClientConfig()
	if err != nil {
		return nil, err
	}
	return c.RegisterByConfig(restConfig)

}

// RegisterByStringWithID 通过kubeconfig文件的string内容和指定ID进行注册
// 参数:
//   - str: kubeconfig文件的字符串内容
//   - id: 集群的唯一标识符
//
// 返回值:
//   - *Kubectl: 成功时返回 Kubectl 实例，用于操作集群
//   - error: 失败时返回错误信息
//
// 此函数允许指定自定义的集群ID，而不是使用默认的服务器地址
func (c *ClusterInstances) RegisterByStringWithID(str string, id string) (*Kubectl, error) {
	config, err := clientcmd.Load([]byte(str))
	if err != nil {
		return nil, fmt.Errorf("RegisterByStringWithID Error content=\n%s\n,id:%s,err:%v", str, id, err)
	}

	clientConfig := clientcmd.NewDefaultClientConfig(*config, &clientcmd.ConfigOverrides{})
	restConfig, err := clientConfig.ClientConfig()
	if err != nil {
		return nil, err
	}

	return c.RegisterByConfigWithID(restConfig, id)
}

// RegisterByPathWithID 通过kubeconfig文件路径和指定ID注册集群
// 参数:
//   - path: kubeconfig文件的路径
//   - id: 集群的唯一标识符
//
// 返回值:
//   - *Kubectl: 成功时返回 Kubectl 实例，用于操作集群
//   - error: 失败时返回错误信息
//
// 此函数允许指定自定义的集群ID，而不是使用默认的服务器地址
func (c *ClusterInstances) RegisterByPathWithID(path string, id string) (*Kubectl, error) {
	config, err := clientcmd.BuildConfigFromFlags("", path)
	if err != nil {
		return nil, fmt.Errorf("RegisterByPathWithID Error path:%s,id:%s,err:%v", path, id, err)
	}
	return c.RegisterByConfigWithID(config, id)
}
