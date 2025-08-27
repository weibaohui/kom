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

// SetEKSConfig 设置 EKS 配置
func (ap *AuthProvider) SetEKSConfig(config *EKSAuthConfig) {
	ap.eksConfig = config
}

// SetTokenManager 设置 token 管理器
func (ap *AuthProvider) SetTokenManager(manager *TokenManager) {
	ap.tokenManager = manager
}
