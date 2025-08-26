package aws

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"k8s.io/klog/v2"
)

// TokenManager AWS token 管理器
type TokenManager struct {
	eksConfig    *EKSAuthConfig
	executor     *ExecExecutor
	awsConfig    aws.Config
	stsClient    *sts.Client
	refreshChan  chan struct{}
	stopChan     chan struct{}
}

// NewTokenManager 创建新的 token 管理器
func NewTokenManager(eksConfig *EKSAuthConfig) (*TokenManager, error) {
	if eksConfig == nil {
		return nil, NewEKSAuthError(ErrorTypeAWSConfigMissing, "EKS config is required", nil)
	}

	tm := &TokenManager{
		eksConfig:   eksConfig,
		executor:    NewExecExecutor(),
		refreshChan: make(chan struct{}, 1),
		stopChan:    make(chan struct{}),
	}

	// 初始化 AWS 配置
	if err := tm.initAWSConfig(context.Background()); err != nil {
		return nil, err
	}

	return tm, nil
}

// initAWSConfig 初始化 AWS 配置
func (tm *TokenManager) initAWSConfig(ctx context.Context) error {
	var opts []func(*config.LoadOptions) error

	// 设置区域
	if tm.eksConfig.Region != "" {
		opts = append(opts, config.WithRegion(tm.eksConfig.Region))
	}

	// 设置 Profile
	if tm.eksConfig.Profile != "" {
		opts = append(opts, config.WithSharedConfigProfile(tm.eksConfig.Profile))
	}

	awsConfig, err := config.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return NewEKSAuthError(ErrorTypeAWSConfigMissing, "failed to load AWS config", err)
	}

	tm.awsConfig = awsConfig
	tm.stsClient = sts.NewFromConfig(awsConfig)
	
	// 保存到 EKS 配置中
	tm.eksConfig.AWSConfig = &awsConfig

	klog.V(2).Infof("Initialized AWS config with region: %s, profile: %s", 
		awsConfig.Region, tm.eksConfig.Profile)

	return nil
}

// GetValidToken 获取有效的 token
func (tm *TokenManager) GetValidToken(ctx context.Context) (string, error) {
	// 检查缓存中的 token 是否有效
	if tm.eksConfig.TokenCache.IsValid() {
		token, _ := tm.eksConfig.TokenCache.GetToken()
		klog.V(4).Infof("Using cached AWS token")
		return token, nil
	}

	klog.V(3).Infof("Cached token expired or missing, fetching new token")
	return tm.refreshToken(ctx)
}

// refreshToken 刷新 token
func (tm *TokenManager) refreshToken(ctx context.Context) (string, error) {
	// 验证 exec 配置
	if err := tm.executor.ValidateCommand(tm.eksConfig.ExecConfig); err != nil {
		return "", err
	}

	// 执行 AWS CLI 命令获取 token
	tokenResponse, err := tm.executor.GetTokenWithRetry(ctx, tm.eksConfig.ExecConfig, 2)
	if err != nil {
		return "", err
	}

	// 更新缓存
	tm.eksConfig.TokenCache.SetToken(tokenResponse.Status.Token, tokenResponse.Status.ExpirationTimestamp)

	klog.V(2).Infof("Successfully refreshed AWS token, expires at: %v", 
		tokenResponse.Status.ExpirationTimestamp)

	return tokenResponse.Status.Token, nil
}

// RefreshToken 公共方法刷新 token
func (tm *TokenManager) RefreshToken(ctx context.Context) error {
	_, err := tm.refreshToken(ctx)
	return err
}

// StartAutoRefresh 启动自动刷新机制
func (tm *TokenManager) StartAutoRefresh(ctx context.Context) {
	go tm.autoRefreshLoop(ctx)
}

// autoRefreshLoop 自动刷新循环
func (tm *TokenManager) autoRefreshLoop(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Minute) // 每5分钟检查一次
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			klog.V(2).Infof("Stopping AWS token auto-refresh due to context cancellation")
			return
		case <-tm.stopChan:
			klog.V(2).Infof("Stopping AWS token auto-refresh")
			return
		case <-ticker.C:
			tm.checkAndRefreshToken(ctx)
		case <-tm.refreshChan:
			// 手动触发刷新
			tm.checkAndRefreshToken(ctx)
		}
	}
}

// checkAndRefreshToken 检查并刷新 token
func (tm *TokenManager) checkAndRefreshToken(ctx context.Context) {
	token, expiresAt := tm.eksConfig.TokenCache.GetToken()
	
	// 如果 token 在10分钟内过期，则刷新
	refreshThreshold := time.Now().Add(10 * time.Minute)
	if token == "" || expiresAt.Before(refreshThreshold) {
		klog.V(3).Infof("Token will expire soon, refreshing...")
		if err := tm.RefreshToken(ctx); err != nil {
			klog.Errorf("Failed to refresh AWS token: %v", err)
		}
	}
}

// TriggerRefresh 触发立即刷新
func (tm *TokenManager) TriggerRefresh() {
	select {
	case tm.refreshChan <- struct{}{}:
	default:
		// 如果通道已满，忽略
	}
}

// Stop 停止自动刷新
func (tm *TokenManager) Stop() {
	close(tm.stopChan)
}

// GetTokenInfo 获取 token 信息
func (tm *TokenManager) GetTokenInfo() (token string, expiresAt time.Time, valid bool) {
	token, expiresAt = tm.eksConfig.TokenCache.GetToken()
	valid = tm.eksConfig.TokenCache.IsValid()
	return
}

// ValidateAWSCredentials 验证 AWS 凭证
func (tm *TokenManager) ValidateAWSCredentials(ctx context.Context) error {
	if tm.stsClient == nil {
		return NewEKSAuthError(ErrorTypeAWSConfigMissing, "STS client not initialized", nil)
	}

	// 调用 GetCallerIdentity 验证凭证
	_, err := tm.stsClient.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return NewEKSAuthError(ErrorTypePermissionDenied, "AWS credentials validation failed", err)
	}

	klog.V(2).Infof("AWS credentials validated successfully")
	return nil
}

// GetCallerIdentity 获取当前 AWS 身份信息
func (tm *TokenManager) GetCallerIdentity(ctx context.Context) (*sts.GetCallerIdentityOutput, error) {
	if tm.stsClient == nil {
		return nil, NewEKSAuthError(ErrorTypeAWSConfigMissing, "STS client not initialized", nil)
	}

	return tm.stsClient.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
}

// AssumeRole 承担 IAM 角色（如果配置了）
func (tm *TokenManager) AssumeRole(ctx context.Context) error {
	if tm.eksConfig.RoleARN == "" {
		// 没有配置角色，跳过
		return nil
	}

	if tm.stsClient == nil {
		return NewEKSAuthError(ErrorTypeAWSConfigMissing, "STS client not initialized", nil)
	}

	klog.V(2).Infof("Assuming role: %s", tm.eksConfig.RoleARN)

	sessionName := fmt.Sprintf("kom-eks-%d", time.Now().Unix())
	result, err := tm.stsClient.AssumeRole(ctx, &sts.AssumeRoleInput{
		RoleArn:         &tm.eksConfig.RoleARN,
		RoleSessionName: &sessionName,
	})
	if err != nil {
		return NewEKSAuthError(ErrorTypePermissionDenied, 
			fmt.Sprintf("failed to assume role %s", tm.eksConfig.RoleARN), err)
	}

	// 更新 AWS 配置使用临时凭证
	tm.awsConfig.Credentials = aws.NewCredentialsCache(aws.CredentialsProviderFunc(func(ctx context.Context) (aws.Credentials, error) {
		return aws.Credentials{
			AccessKeyID:     *result.Credentials.AccessKeyId,
			SecretAccessKey: *result.Credentials.SecretAccessKey,
			SessionToken:    *result.Credentials.SessionToken,
			Expires:         *result.Credentials.Expiration,
		}, nil
	}))

	klog.V(2).Infof("Successfully assumed role: %s", tm.eksConfig.RoleARN)
	return nil
}

// ClearCache 清理 token 缓存
func (tm *TokenManager) ClearCache() {
	tm.eksConfig.TokenCache.ClearToken()
	klog.V(2).Infof("Cleared AWS token cache")
}