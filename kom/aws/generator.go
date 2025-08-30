package aws

import (
	"crypto/rand"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"k8s.io/klog/v2"
)

// KubeconfigGenerator Kubeconfig生成器
type KubeconfigGenerator struct {
	maxRetryAttempts int
	retryInterval    time.Duration
	awsCLITimeout    time.Duration
}

// NewKubeconfigGenerator 创建Kubeconfig生成器
func NewKubeconfigGenerator() *KubeconfigGenerator {
	return &KubeconfigGenerator{
		maxRetryAttempts: 3,
		retryInterval:    2 * time.Second,
		awsCLITimeout:    30 * time.Second,
	}
}

// GenerateFromAWS 通过AWS CLI生成kubeconfig
func (kg *KubeconfigGenerator) GenerateFromAWS(config *EKSAuthConfig) (string, error) {
	if err := kg.validateConfig(config); err != nil {
		return "", err
	}

	// 生成临时kubeconfig文件路径
	kubeconfigPath, err := kg.generateTempKubeconfigPath(config.ClusterName)
	if err != nil {
		return "", NewEKSAuthError(ErrorTypeFileSystemError, "Failed to generate temp kubeconfig path", err)
	}

	// 确保在函数退出时清理临时文件
	defer func() {
		if err := os.Remove(kubeconfigPath); err != nil && !os.IsNotExist(err) {
			klog.V(3).Infof("Failed to cleanup temp kubeconfig file %s: %v", kubeconfigPath, err)
		}
	}()

	// 构建环境变量
	envVars := kg.buildEnvVariables(config, kubeconfigPath)
	// 执行AWS CLI命令
	if err := kg.executeAWSCommandWithEnv(config, envVars); err != nil {
		return "", err
	}

	// 验证生成的kubeconfig文件
	if err := kg.validateKubeconfig(kubeconfigPath); err != nil {
		return "", err
	}

	// 读取kubeconfig内容
	kubeconfigContent, err := ioutil.ReadFile(kubeconfigPath)
	if err != nil {
		return "", NewEKSAuthError(ErrorTypeFileSystemError, "Failed to read generated kubeconfig", err)
	}

	return string(kubeconfigContent), nil
}

// validateConfig 验证配置
func (kg *KubeconfigGenerator) validateConfig(config *EKSAuthConfig) error {
	if config.AccessKey == "" {
		return NewEKSAuthError(ErrorTypeInvalidCredentials, "AccessKey is required", nil)
	}
	if config.SecretAccessKey == "" {
		return NewEKSAuthError(ErrorTypeInvalidCredentials, "SecretAccessKey is required", nil)
	}
	if config.Region == "" {
		return NewEKSAuthError(ErrorTypeInvalidCredentials, "Region is required", nil)
	}
	if config.ClusterName == "" {
		return NewEKSAuthError(ErrorTypeInvalidCredentials, "ClusterName is required", nil)
	}
	return nil
}

// generateTempKubeconfigPath 生成临时kubeconfig文件路径
func (kg *KubeconfigGenerator) generateTempKubeconfigPath(clusterName string) (string, error) {
	// 生成随机后缀以避免冲突
	randomSuffix, err := generateRandomString(8)
	if err != nil {
		return "", err
	}

	fileName := fmt.Sprintf("%s-kubeconfig-%s.yaml", clusterName, randomSuffix)
	return filepath.Join("/tmp", fileName), nil
}

// generateRandomString 生成安全的随机字符串
func generateRandomString(length int) (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	for i := range b {
		b[i] = charset[b[i]%byte(len(charset))]
	}
	return string(b), nil
}

// buildEnvVariables 构建环境变量
func (kg *KubeconfigGenerator) buildEnvVariables(config *EKSAuthConfig, kubeconfigPath string) []string {
	envVars := []string{
		fmt.Sprintf("AWS_ACCESS_KEY_ID=%s", config.AccessKey),
		fmt.Sprintf("AWS_SECRET_ACCESS_KEY=%s", config.SecretAccessKey),
		fmt.Sprintf("KUBECONFIG=%s", kubeconfigPath),
	}

	// 添加可选的角色ARN
	if config.RoleARN != "" {
		envVars = append(envVars, fmt.Sprintf("AWS_ROLE_ARN=%s", config.RoleARN))
		if config.SessionName != "" {
			envVars = append(envVars, fmt.Sprintf("AWS_ROLE_SESSION_NAME=%s", config.SessionName))
		}
	}

	return envVars
}

// executeAWSCommandWithEnv 执行AWS CLI命令
func (kg *KubeconfigGenerator) executeAWSCommandWithEnv(config *EKSAuthConfig, envVars []string) error {
	// 构建AWS CLI命令
	args := []string{
		"eks",
		"update-kubeconfig",
		"--region", config.Region,
		"--name", config.ClusterName,
	}

	var lastErr error
	for attempt := 0; attempt < kg.maxRetryAttempts; attempt++ {
		if attempt > 0 {
			klog.V(3).Infof("Retrying AWS CLI execution (attempt %d/%d) after %v", attempt+1, kg.maxRetryAttempts, kg.retryInterval)
			time.Sleep(kg.retryInterval)
		}
		cmd := exec.Command("aws", args...)
		cmd.Env = envVars

		klog.V(8).Infof("Executing AWS CLI command: aws %s", strings.Join(args, " "))

		output, err := cmd.CombinedOutput()
		if err == nil {
			klog.V(3).Infof("AWS CLI command executed successfully")
			return nil
		}

		lastErr = NewEKSAuthError(ErrorTypeExecFailed,
			fmt.Sprintf("AWS CLI execution failed (attempt %d/%d): %s", attempt+1, kg.maxRetryAttempts, string(output)), err)

		klog.V(2).Infof("AWS CLI execution failed (attempt %d/%d): %v", attempt+1, kg.maxRetryAttempts, lastErr)
	}

	return lastErr
}

// validateKubeconfig 验证生成的kubeconfig文件
func (kg *KubeconfigGenerator) validateKubeconfig(kubeconfigPath string) error {
	// 检查文件是否存在
	if _, err := os.Stat(kubeconfigPath); os.IsNotExist(err) {
		return NewEKSAuthError(ErrorTypeKubeconfigInvalid, "Generated kubeconfig file does not exist", err)
	}

	// 检查文件是否可读
	content, err := ioutil.ReadFile(kubeconfigPath)
	if err != nil {
		return NewEKSAuthError(ErrorTypeFileSystemError, "Cannot read generated kubeconfig file", err)
	}

	// 简单验证文件内容
	contentStr := string(content)
	if !strings.Contains(contentStr, "apiVersion: v1") ||
		!strings.Contains(contentStr, "kind: Config") ||
		!strings.Contains(contentStr, "clusters:") {
		return NewEKSAuthError(ErrorTypeKubeconfigInvalid, "Generated kubeconfig file format is invalid", nil)
	}

	klog.V(8).Infof("Kubeconfig validation successful for file: %s", kubeconfigPath)
	return nil
}

// SetMaxRetryAttempts 设置最大重试次数
func (kg *KubeconfigGenerator) SetMaxRetryAttempts(attempts int) {
	kg.maxRetryAttempts = attempts
}

// SetRetryInterval 设置重试间隔
func (kg *KubeconfigGenerator) SetRetryInterval(interval time.Duration) {
	kg.retryInterval = interval
}

// SetAWSCLITimeout 设置AWS CLI超时时间
func (kg *KubeconfigGenerator) SetAWSCLITimeout(timeout time.Duration) {
	kg.awsCLITimeout = timeout
}
