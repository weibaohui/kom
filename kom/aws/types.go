package aws

import (
	"context"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"k8s.io/client-go/tools/clientcmd/api"
)

// AWSAuthProvider AWS 认证提供者接口
type AWSAuthProvider interface {
	// GetToken 获取认证 token
	GetToken(ctx context.Context) (string, time.Time, error)
	// RefreshToken 刷新 token
	RefreshToken(ctx context.Context) error
	// IsEKSConfig 检测是否为 EKS 配置
	IsEKSConfig(kubeconfig []byte) bool
}

// EKSAuthConfig AWS EKS 认证配置
type EKSAuthConfig struct {
	ClusterName   string            `json:"cluster_name"`   // EKS 集群名称
	Region        string            `json:"region"`         // AWS 区域
	Profile       string            `json:"profile"`        // AWS Profile (可选)
	RoleARN       string            `json:"role_arn"`       // 要承担的 IAM 角色 ARN (可选)
	ExecConfig    *ExecConfig       `json:"exec_config"`    // exec 命令配置
	TokenCache    *TokenCache       `json:"token_cache"`    // token 缓存
	AWSConfig     *aws.Config       `json:"-"`              // AWS 配置，不序列化
}

// ExecConfig 执行命令配置
type ExecConfig struct {
	Command string            `json:"command"` // 命令 (如 aws)
	Args    []string          `json:"args"`    // 参数列表
	Env     map[string]string `json:"env"`     // 环境变量
}

// TokenCache token 缓存
type TokenCache struct {
	Token     string        `json:"token"`      // Bearer token
	ExpiresAt time.Time     `json:"expires_at"` // 过期时间
	mutex     sync.RWMutex  `json:"-"`          // 读写锁，不序列化
}

// IsValid 检查 token 是否有效
func (tc *TokenCache) IsValid() bool {
	tc.mutex.RLock()
	defer tc.mutex.RUnlock()
	return tc.Token != "" && time.Now().Before(tc.ExpiresAt)
}

// GetToken 安全获取 token
func (tc *TokenCache) GetToken() (string, time.Time) {
	tc.mutex.RLock()
	defer tc.mutex.RUnlock()
	return tc.Token, tc.ExpiresAt
}

// SetToken 安全设置 token
func (tc *TokenCache) SetToken(token string, expiresAt time.Time) {
	tc.mutex.Lock()
	defer tc.mutex.Unlock()
	tc.Token = token
	tc.ExpiresAt = expiresAt
}

// ClearToken 清理 token
func (tc *TokenCache) ClearToken() {
	tc.mutex.Lock()
	defer tc.mutex.Unlock()
	tc.Token = ""
	tc.ExpiresAt = time.Time{}
}

// EKSAuthError EKS 认证错误
type EKSAuthError struct {
	Type    string `json:"type"`    // TokenExpired, AWSConfigMissing, ExecFailed 等
	Message string `json:"message"` // 错误消息
	Cause   error  `json:"-"`       // 原始错误，不序列化
}

func (e *EKSAuthError) Error() string {
	if e.Cause != nil {
		return e.Message + ": " + e.Cause.Error()
	}
	return e.Message
}

// NewEKSAuthError 创建 EKS 认证错误
func NewEKSAuthError(errorType, message string, cause error) *EKSAuthError {
	return &EKSAuthError{
		Type:    errorType,
		Message: message,
		Cause:   cause,
	}
}

// ParseExecConfigFromKubeconfig 从 kubeconfig 中解析 exec 配置
func ParseExecConfigFromKubeconfig(config *api.Config) (*ExecConfig, error) {
	for _, authInfo := range config.AuthInfos {
		if authInfo.Exec != nil {
			execConfig := &ExecConfig{
				Command: authInfo.Exec.Command,
				Args:    make([]string, len(authInfo.Exec.Args)),
				Env:     make(map[string]string),
			}
			
			// 复制参数
			copy(execConfig.Args, authInfo.Exec.Args)
			
			// 复制环境变量
			for _, env := range authInfo.Exec.Env {
				execConfig.Env[env.Name] = env.Value
			}
			
			return execConfig, nil
		}
	}
	
	return nil, NewEKSAuthError("NoExecConfig", "no exec configuration found in kubeconfig", nil)
}

// Constants for error types
const (
	ErrorTypeTokenExpired       = "TokenExpired"
	ErrorTypeAWSConfigMissing   = "AWSConfigMissing"
	ErrorTypeExecFailed         = "ExecFailed"
	ErrorTypeInvalidKubeconfig  = "InvalidKubeconfig"
	ErrorTypeNetworkError       = "NetworkError"
	ErrorTypePermissionDenied   = "PermissionDenied"
)