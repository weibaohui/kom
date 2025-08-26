package aws

import (
	"context"
	"time"

	"k8s.io/klog/v2"
)

// AuthProvider AWS 认证提供者实现
type AuthProvider struct {
	configParser *ConfigParser
	tokenManager *TokenManager
	eksConfig    *EKSAuthConfig
}

// NewAuthProvider 创建新的 AWS 认证提供者
func NewAuthProvider() *AuthProvider {
	return &AuthProvider{
		configParser: NewConfigParser(),
	}
}

// IsEKSConfig 检测是否为 EKS 配置
func (ap *AuthProvider) IsEKSConfig(kubeconfig []byte) bool {
	return ap.configParser.IsEKSConfig(kubeconfig)
}

// InitializeFromKubeconfig 从 kubeconfig 初始化认证提供者
func (ap *AuthProvider) InitializeFromKubeconfig(kubeconfigContent []byte) error {
	// 解析 EKS 配置
	eksConfig, err := ap.configParser.ParseEKSConfig(kubeconfigContent)
	if err != nil {
		return err
	}

	// 验证配置
	if err := ap.configParser.ValidateEKSConfig(eksConfig); err != nil {
		return err
	}

	ap.eksConfig = eksConfig

	// 创建 token 管理器
	tokenManager, err := NewTokenManager(eksConfig)
	if err != nil {
		return err
	}

	ap.tokenManager = tokenManager

	klog.V(2).Infof("Initialized AWS auth provider for cluster: %s, region: %s",
		eksConfig.ClusterName, eksConfig.Region)

	return nil
}

// GetToken 获取认证 token
func (ap *AuthProvider) GetToken(ctx context.Context) (string, time.Time, error) {
	if ap.tokenManager == nil {
		return "", time.Time{}, NewEKSAuthError(ErrorTypeAWSConfigMissing, "token manager not initialized", nil)
	}

	token, err := ap.tokenManager.GetValidToken(ctx)
	if err != nil {
		return "", time.Time{}, err
	}

	_, expiresAt, _ := ap.tokenManager.GetTokenInfo()
	return token, expiresAt, nil
}

// RefreshToken 刷新 token
func (ap *AuthProvider) RefreshToken(ctx context.Context) error {
	if ap.tokenManager == nil {
		return NewEKSAuthError(ErrorTypeAWSConfigMissing, "token manager not initialized", nil)
	}

	return ap.tokenManager.RefreshToken(ctx)
}

// StartAutoRefresh 启动自动刷新
func (ap *AuthProvider) StartAutoRefresh(ctx context.Context) {
	if ap.tokenManager == nil {
		klog.Warning("Token manager not initialized, cannot start auto refresh")
		return
	}

	ap.tokenManager.StartAutoRefresh(ctx)
}

// Stop 停止认证提供者
func (ap *AuthProvider) Stop() {
	if ap.tokenManager != nil {
		ap.tokenManager.Stop()
	}
}

// GetEKSConfig 获取 EKS 配置
func (ap *AuthProvider) GetEKSConfig() *EKSAuthConfig {
	return ap.eksConfig
}

// ValidateCredentials 验证 AWS 凭证
func (ap *AuthProvider) ValidateCredentials(ctx context.Context) error {
	if ap.tokenManager == nil {
		return NewEKSAuthError(ErrorTypeAWSConfigMissing, "token manager not initialized", nil)
	}

	return ap.tokenManager.ValidateAWSCredentials(ctx)
}

// GetClusterInfo 获取集群信息
func (ap *AuthProvider) GetClusterInfo() (clusterName, region, profile string) {
	if ap.eksConfig == nil {
		return "", "", ""
	}
	return ap.eksConfig.ClusterName, ap.eksConfig.Region, ap.eksConfig.Profile
}

// IsTokenValid 检查 token 是否有效
func (ap *AuthProvider) IsTokenValid() bool {
	if ap.eksConfig == nil || ap.eksConfig.TokenCache == nil {
		return false
	}
	return ap.eksConfig.TokenCache.IsValid()
}

// GetTokenExpiry 获取 token 过期时间
func (ap *AuthProvider) GetTokenExpiry() time.Time {
	if ap.eksConfig == nil || ap.eksConfig.TokenCache == nil {
		return time.Time{}
	}
	_, expiresAt := ap.eksConfig.TokenCache.GetToken()
	return expiresAt
}

// ClearTokenCache 清理 token 缓存
func (ap *AuthProvider) ClearTokenCache() {
	if ap.tokenManager != nil {
		ap.tokenManager.ClearCache()
	}
}

// TriggerRefresh 触发立即刷新
func (ap *AuthProvider) TriggerRefresh() {
	if ap.tokenManager != nil {
		ap.tokenManager.TriggerRefresh()
	}
}

// GetCallerIdentity 获取 AWS 身份信息
func (ap *AuthProvider) GetCallerIdentity(ctx context.Context) (string, string, string, error) {
	if ap.tokenManager == nil {
		return "", "", "", NewEKSAuthError(ErrorTypeAWSConfigMissing, "token manager not initialized", nil)
	}

	identity, err := ap.tokenManager.GetCallerIdentity(ctx)
	if err != nil {
		return "", "", "", err
	}

	var account, arn, userId string
	if identity.Account != nil {
		account = *identity.Account
	}
	if identity.Arn != nil {
		arn = *identity.Arn
	}
	if identity.UserId != nil {
		userId = *identity.UserId
	}

	return account, arn, userId, nil
}

// AssumeRole 承担 IAM 角色
func (ap *AuthProvider) AssumeRole(ctx context.Context) error {
	if ap.tokenManager == nil {
		return NewEKSAuthError(ErrorTypeAWSConfigMissing, "token manager not initialized", nil)
	}

	return ap.tokenManager.AssumeRole(ctx)
}

// SetEKSConfig 设置 EKS 配置
func (ap *AuthProvider) SetEKSConfig(config *EKSAuthConfig) {
	ap.eksConfig = config
}

// SetTokenManager 设置 token 管理器
func (ap *AuthProvider) SetTokenManager(manager *TokenManager) {
	ap.tokenManager = manager
}
