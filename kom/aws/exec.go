package aws

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"time"

	"k8s.io/klog/v2"
)

// ExecExecutor 命令执行器
type ExecExecutor struct{}

// NewExecExecutor 创建新的命令执行器
func NewExecExecutor() *ExecExecutor {
	return &ExecExecutor{}
}

// TokenResponse AWS CLI 返回的 token 响应结构
type TokenResponse struct {
	Kind       string    `json:"kind"`
	APIVersion string    `json:"apiVersion"`
	Spec       TokenSpec `json:"spec"`
	Status     TokenStatus `json:"status"`
}

// TokenSpec token 规格
type TokenSpec struct {
	Interactive bool `json:"interactive"`
}

// TokenStatus token 状态
type TokenStatus struct {
	ExpirationTimestamp time.Time `json:"expirationTimestamp"`
	Token               string    `json:"token"`
}

// ExecuteCommand 执行 AWS CLI 命令获取 token
func (ee *ExecExecutor) ExecuteCommand(ctx context.Context, execConfig *ExecConfig) (*TokenResponse, error) {
	if execConfig == nil {
		return nil, NewEKSAuthError(ErrorTypeExecFailed, "exec config is nil", nil)
	}

	klog.V(4).Infof("Executing AWS command: %s %v", execConfig.Command, execConfig.Args)

	// 创建命令
	cmd := exec.CommandContext(ctx, execConfig.Command, execConfig.Args...)

	// 设置环境变量
	if len(execConfig.Env) > 0 {
		env := make([]string, 0, len(execConfig.Env))
		for key, value := range execConfig.Env {
			env = append(env, fmt.Sprintf("%s=%s", key, value))
		}
		cmd.Env = append(cmd.Env, env...)
	}

	// 执行命令
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		klog.Errorf("Failed to execute AWS command: %v, stderr: %s", err, stderr.String())
		return nil, NewEKSAuthError(ErrorTypeExecFailed, 
			fmt.Sprintf("failed to execute AWS command: %s", stderr.String()), err)
	}

	// 解析输出
	var tokenResponse TokenResponse
	if err := json.Unmarshal(stdout.Bytes(), &tokenResponse); err != nil {
		klog.Errorf("Failed to parse AWS command output: %v, output: %s", err, stdout.String())
		return nil, NewEKSAuthError(ErrorTypeExecFailed, 
			fmt.Sprintf("failed to parse AWS command output: %v", err), err)
	}

	// 验证响应
	if tokenResponse.Status.Token == "" {
		return nil, NewEKSAuthError(ErrorTypeExecFailed, "empty token in AWS response", nil)
	}

	klog.V(4).Infof("Successfully obtained AWS token, expires at: %v", 
		tokenResponse.Status.ExpirationTimestamp)

	return &tokenResponse, nil
}

// GetTokenWithRetry 带重试机制的获取 token
func (ee *ExecExecutor) GetTokenWithRetry(ctx context.Context, execConfig *ExecConfig, maxRetries int) (*TokenResponse, error) {
	var lastErr error
	
	for i := 0; i <= maxRetries; i++ {
		if i > 0 {
			klog.V(2).Infof("Retrying AWS token fetch, attempt %d/%d", i+1, maxRetries+1)
			
			// 指数退避
			waitTime := time.Duration(i) * time.Second
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(waitTime):
			}
		}

		tokenResponse, err := ee.ExecuteCommand(ctx, execConfig)
		if err == nil {
			return tokenResponse, nil
		}
		
		lastErr = err
		klog.V(3).Infof("AWS token fetch attempt %d failed: %v", i+1, err)
	}

	return nil, fmt.Errorf("failed to get AWS token after %d retries: %w", maxRetries+1, lastErr)
}

// ValidateCommand 验证命令是否可执行
func (ee *ExecExecutor) ValidateCommand(execConfig *ExecConfig) error {
	if execConfig == nil {
		return NewEKSAuthError(ErrorTypeExecFailed, "exec config is nil", nil)
	}

	if execConfig.Command == "" {
		return NewEKSAuthError(ErrorTypeExecFailed, "command is empty", nil)
	}

	// 检查命令是否存在
	_, err := exec.LookPath(execConfig.Command)
	if err != nil {
		return NewEKSAuthError(ErrorTypeExecFailed, 
			fmt.Sprintf("command %s not found in PATH", execConfig.Command), err)
	}

	return nil
}

// BuildGetTokenCommand 构建获取 token 的命令配置
func (ee *ExecExecutor) BuildGetTokenCommand(clusterName, region, profile, roleARN string) *ExecConfig {
	args := []string{"eks", "get-token", "--cluster-name", clusterName}
	
	if region != "" {
		args = append(args, "--region", region)
	}
	
	if roleARN != "" {
		args = append(args, "--role-arn", roleARN)
	}

	execConfig := &ExecConfig{
		Command: "aws",
		Args:    args,
		Env:     make(map[string]string),
	}

	if profile != "" {
		execConfig.Env["AWS_PROFILE"] = profile
	}

	return execConfig
}

// IsCommandAvailable 检查 AWS CLI 是否可用
func (ee *ExecExecutor) IsCommandAvailable() bool {
	_, err := exec.LookPath("aws")
	return err == nil
}

// GetAWSCLIVersion 获取 AWS CLI 版本信息
func (ee *ExecExecutor) GetAWSCLIVersion(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "aws", "--version")
	
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	
	if err := cmd.Run(); err != nil {
		return "", err
	}
	
	return stdout.String(), nil
}