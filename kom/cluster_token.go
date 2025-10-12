package kom

import (
    "fmt"

    "k8s.io/client-go/rest"
)

// RegisterByTokenWithServerAndID 通过token和服务器地址注册集群
// 参数:
//   - token: Kubernetes 集群的访问令牌 (Bearer Token)
//   - server: Kubernetes API 服务器地址 (例如: https://kubernetes.example.com:6443)
//   - id: 集群的唯一标识符
//   - caData: 可选的 CA 证书数据，用于启用 TLS 验证
//
// 返回值:
//   - *Kubectl: 成功时返回 Kubectl 实例，用于操作集群
//   - error: 失败时返回错误信息
//
// 这是推荐的token注册方式，因为它提供了完整的集群连接信息
func (c *ClusterInstances) RegisterByTokenWithServerAndID(token string, server string, id string, opts ...RegisterOption) (*Kubectl, error) {
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
    return c.RegisterByConfigWithID(config, id, opts...)
}

// RegisterByTokenWithServerAndIDLegacy 兼容旧签名：支持 caData 可变参数
// Deprecated: 请使用带 opts 的 RegisterByTokenWithServerAndID，并通过 RegisterCACert 指定 CA。
func (c *ClusterInstances) RegisterByTokenWithServerAndIDLegacy(token string, server string, id string, caData ...string) (*Kubectl, error) {
    var opts []RegisterOption
    if len(caData) > 0 {
        opts = append(opts, RegisterCACert([]byte(caData[0])))
    }
    return c.RegisterByTokenWithServerAndID(token, server, id, opts...)
}
