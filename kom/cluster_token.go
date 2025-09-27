package kom

import (
	"fmt"
	"k8s.io/client-go/rest"
)

// RegisterByTokenWithID 通过token注册集群（简化版本）
// 参数:
//   - token: Kubernetes 集群的访问令牌 (Bearer Token)
//   - id: 集群的唯一标识符
//
// 返回值:
//   - *Kubectl: 成功时返回 Kubectl 实例，用于操作集群
//   - error: 失败时返回错误信息
//
// 注意: 此方法需要配合其他方式设置服务器地址，推荐使用 RegisterByTokenWithServerAndID
func (c *ClusterInstances) RegisterByTokenWithID(token string, id string) (*Kubectl, error) {
	// 参数验证
	if token == "" {
		return nil, fmt.Errorf("token cannot be empty")
	}
	if id == "" {
		return nil, fmt.Errorf("cluster id cannot be empty")
	}

	config := &rest.Config{
		BearerToken: token,
	}
	return c.RegisterByConfigWithID(config, id)
}

// RegisterByTokenWithServerAndID 通过token和服务器地址注册集群
// 参数:
//   - token: Kubernetes 集群的访问令牌 (Bearer Token)
//   - server: Kubernetes API 服务器地址 (例如: https://kubernetes.example.com:6443)
//   - id: 集群的唯一标识符
//
// 返回值:
//   - *Kubectl: 成功时返回 Kubectl 实例，用于操作集群
//   - error: 失败时返回错误信息
//
// 这是推荐的token注册方式，因为它提供了完整的集群连接信息
func (c *ClusterInstances) RegisterByTokenWithServerAndID(token string, server string, id string) (*Kubectl, error) {
	// 参数验证
	if token == "" {
		return nil, fmt.Errorf("token cannot be empty")
	}
	if server == "" {
		return nil, fmt.Errorf("server address cannot be empty")
	}
	if id == "" {
		return nil, fmt.Errorf("cluster id cannot be empty")
	}

	config := &rest.Config{
		Host:        server,
		BearerToken: token,
		TLSClientConfig: rest.TLSClientConfig{
			Insecure: false, // 默认启用 TLS 验证，可根据需要调整
		},
	}
	return c.RegisterByConfigWithID(config, id)
}

// RegisterByTokenWithOptions 通过token和详细选项注册集群
// 参数:
//   - token: Kubernetes 集群的访问令牌 (Bearer Token)
//   - server: Kubernetes API 服务器地址 (例如: https://kubernetes.example.com:6443)
//   - id: 集群的唯一标识符
//   - insecure: 是否跳过 TLS 证书验证 (生产环境建议设为 false)
//
// 返回值:
//   - *Kubectl: 成功时返回 Kubectl 实例，用于操作集群
//   - error: 失败时返回错误信息
//
// 此函数提供了最灵活的token注册方式，允许自定义TLS设置
func (c *ClusterInstances) RegisterByTokenWithOptions(token string, server string, id string, insecure bool) (*Kubectl, error) {
	// 参数验证
	if token == "" {
		return nil, fmt.Errorf("token cannot be empty")
	}
	if server == "" {
		return nil, fmt.Errorf("server address cannot be empty")
	}
	if id == "" {
		return nil, fmt.Errorf("cluster id cannot be empty")
	}

	config := &rest.Config{
		Host:        server,
		BearerToken: token,
		TLSClientConfig: rest.TLSClientConfig{
			Insecure: insecure,
		},
	}
	return c.RegisterByConfigWithID(config, id)
}