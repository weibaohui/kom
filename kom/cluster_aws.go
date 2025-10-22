package kom

import (
	"context"
	"fmt"
	"time"

	"github.com/weibaohui/kom/kom/aws"
	"k8s.io/klog/v2"
)

// NewAWSAuthProvider 创建新的AWS认证提供者
// 返回值:
//   - *aws.AuthProvider: AWS认证提供者实例
//
// 此函数用于创建AWS认证提供者，用于EKS集群的认证管理
func NewAWSAuthProvider() *aws.AuthProvider {
	return aws.NewAuthProvider()
}

// RegisterAWSCluster 注册EKS集群（自动生成集群ID）
// 参数:
//   - config: AWS EKS认证配置，包含区域、集群名称等信息
//
// 返回值:
//   - *Kubectl: 成功时返回 Kubectl 实例，用于操作集群
//   - error: 失败时返回错误信息
//
// 此函数会自动生成集群ID（格式：区域-集群名称），然后调用RegisterAWSClusterWithID
func (c *ClusterInstances) RegisterAWSCluster(config *aws.EKSAuthConfig, opts ...RegisterOption) (*Kubectl, error) {

	if config == nil {
		return nil, fmt.Errorf("RegisterAWSCluster: config is nil")
	}
	if config.Region == "" || config.ClusterName == "" {
		return nil, fmt.Errorf("RegisterAWSCluster: region or cluster_name is empty")
	}
	// 生成集群ID
    clusterID := fmt.Sprintf("%s-%s", config.Region, config.ClusterName)
    return c.RegisterAWSClusterWithID(config, clusterID, opts...)
}

// RegisterAWSClusterWithID 通过指定ID注册EKS集群
// 参数:
//   - config: AWS EKS认证配置，包含区域、集群名称、认证信息等
//   - clusterID: 集群的唯一标识符
//
// 返回值:
//   - *Kubectl: 成功时返回 Kubectl 实例，用于操作集群
//   - error: 失败时返回错误信息
//
// 此函数执行以下操作：
// 1. 检查集群是否已存在
// 2. 生成EKS集群的kubeconfig
// 3. 设置AWS token管理器和自动刷新
// 4. 启动token刷新循环
// 5. 注册集群并返回Kubectl实例
func (c *ClusterInstances) RegisterAWSClusterWithID(config *aws.EKSAuthConfig, clusterID string, opts ...RegisterOption) (*Kubectl, error) {

	var cluster *ClusterInst
	// 检查是否已存在
	if value, exists := clusterInstances.clusters.Load(clusterID); exists {
		cluster = value.(*ClusterInst)
		if cluster.Kubectl != nil {
			return cluster.Kubectl, nil
		}
	}

	var kubeconfigContent string
	var err error

	if cluster == nil {
		// 生成kubeconfig
		generator := aws.NewKubeconfigGenerator()
		kubeconfigContent, err = generator.GenerateFromAWS(config)
		if err != nil {
			return nil, fmt.Errorf("failed to generate kubeconfig for EKS cluster %s: %w", clusterID, err)
		}
		klog.V(2).Infof("Generated kubeconfig for EKS cluster: %s", clusterID)

		// 手动设置 EKS 配置
		tokenManager, err := aws.NewTokenManager(config)
		if err != nil {
			return nil, fmt.Errorf("failed to create AWS token manager: %w", err)
		}

		authProvider := NewAWSAuthProvider()
		// 设置内部状态
		authProvider.SetEKSConfig(config)
		authProvider.SetTokenManager(tokenManager)
		config.TokenCache = &aws.TokenCache{}
		tokenCtx, tokenCancel := context.WithCancel(context.Background())
		// 启动自动刷新
		authProvider.StartAutoRefresh(tokenCtx)

		// 启动 token 刷新循环
		go func() {
			ticker := time.NewTicker(5 * time.Minute) // 每5分钟检查一次
			defer ticker.Stop()

			for {
				select {
				case <-tokenCtx.Done():
					klog.V(2).Infof("Stopping token refresh for cluster %s", clusterID)
					return
				case <-ticker.C:
					// 检查 token 是否需要刷新
					if !authProvider.IsTokenValid() {
						klog.V(3).Infof("Token expired for cluster %s, refreshing...", clusterID)
						if token, _, err := authProvider.GetToken(tokenCtx); err != nil {
							klog.Errorf("Failed to refresh token for cluster %s: %v", clusterID, err)
						} else {
							// 更新 rest.Config 中的 BearerToken
							cluster.Config.BearerToken = token
							klog.V(2).Infof("Successfully refreshed token for cluster %s", clusterID)
						}
					}
				}
			}
		}()
		cluster = &ClusterInst{
			ID:                 clusterID,
			IsEKS:              true,
			AWSAuthProvider:    authProvider,
			tokenRefreshCancel: tokenCancel,
		}
	}
	klog.V(8).Infof("kubeconfigContent=\n%s", kubeconfigContent)

	clusterInstances.clusters.Store(clusterID, cluster)

	// 使用kubeconfig注册集群
    kubectl, err := c.RegisterByStringWithID(kubeconfigContent, clusterID, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to register EKS cluster %s: %w", clusterID, err)
	}

	klog.V(1).Infof("Successfully registered EKS cluster: %s", clusterID)

	return kubectl, nil
}

// RefreshEKSCredentials 刷新EKS集群凭证
// 参数:
//   - clusterID: 需要刷新凭证的集群ID
//
// 返回值:
//   - error: 失败时返回错误信息，成功时返回nil
//
// 此函数用于手动触发EKS集群的token刷新，通常在token即将过期或已过期时调用
func (c *ClusterInstances) RefreshEKSCredentials(clusterID string) error {

	// 如果集群实例存在且为EKS集群，触发token刷新
	if cluster := c.GetClusterById(clusterID); cluster != nil && cluster.IsEKS {
		if cluster.AWSAuthProvider != nil {
			cluster.AWSAuthProvider.TriggerRefresh()
		}

	}

	return nil
}

// IsEKSCluster 检查集群是否为 EKS 集群
func (ci *ClusterInst) IsEKSCluster() bool {
	return ci.IsEKS
}

// GetAWSAuthProvider 获取 AWS 认证提供者
func (ci *ClusterInst) GetAWSAuthProvider() *aws.AuthProvider {
	return ci.AWSAuthProvider
}

// StopTokenRefresh 停止 token 自动刷新
func (ci *ClusterInst) StopTokenRefresh() {
	if ci.tokenRefreshCancel != nil {
		ci.tokenRefreshCancel()
		ci.tokenRefreshCancel = nil
	}
}
