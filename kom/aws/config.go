package aws

import (
	"fmt"
	"strings"

	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/klog/v2"
)

// ConfigParser EKS 配置解析器
type ConfigParser struct{}

// NewConfigParser 创建新的配置解析器
func NewConfigParser() *ConfigParser {
	return &ConfigParser{}
}

// IsEKSConfig 检测 kubeconfig 是否为 EKS 配置
func (cp *ConfigParser) IsEKSConfig(kubeconfigContent []byte) bool {
	config, err := clientcmd.Load(kubeconfigContent)
	if err != nil {
		klog.V(4).Infof("Failed to load kubeconfig: %v", err)
		return false
	}

	return cp.hasEKSExecConfig(config)
}

// hasEKSExecConfig 检查是否包含 EKS exec 配置
func (cp *ConfigParser) hasEKSExecConfig(config *api.Config) bool {
	for _, authInfo := range config.AuthInfos {
		if authInfo.Exec != nil {
			// 检查是否为 AWS EKS 相关的 exec 命令
			if cp.isAWSEKSExecCommand(authInfo.Exec) {
				return true
			}
		}
	}
	return false
}

// isAWSEKSExecCommand 检查是否为 AWS EKS exec 命令
func (cp *ConfigParser) isAWSEKSExecCommand(exec *api.ExecConfig) bool {
	if exec == nil {
		return false
	}

	// 检查命令是否为 aws
	if exec.Command != "aws" {
		return false
	}

	// 检查参数中是否包含 eks get-token
	args := strings.Join(exec.Args, " ")
	return strings.Contains(args, "eks") && strings.Contains(args, "get-token")
}

// ParseEKSConfig 解析 EKS kubeconfig 配置
func (cp *ConfigParser) ParseEKSConfig(kubeconfigContent []byte) (*EKSAuthConfig, error) {
	config, err := clientcmd.Load(kubeconfigContent)
	if err != nil {
		return nil, NewEKSAuthError(ErrorTypeInvalidKubeconfig, "failed to load kubeconfig", err)
	}

	// 检查是否为 EKS 配置
	if !cp.hasEKSExecConfig(config) {
		return nil, NewEKSAuthError(ErrorTypeInvalidKubeconfig, "not an EKS kubeconfig", nil)
	}

	eksConfig := &EKSAuthConfig{
		TokenCache: &TokenCache{},
	}

	// 解析 exec 配置
	execConfig, err := ParseExecConfigFromKubeconfig(config)
	if err != nil {
		return nil, err
	}
	eksConfig.ExecConfig = execConfig

	// 从 exec 参数中提取集群名称和区域
	err = cp.extractClusterInfoFromArgs(execConfig.Args, eksConfig)
	if err != nil {
		return nil, err
	}

	// 从环境变量中提取 AWS Profile
	if profile, exists := execConfig.Env["AWS_PROFILE"]; exists {
		eksConfig.Profile = profile
	}

	// 从环境变量中提取 AWS 区域（如果命令行参数中没有）
	if eksConfig.Region == "" {
		if region, exists := execConfig.Env["AWS_REGION"]; exists {
			eksConfig.Region = region
		} else if region, exists := execConfig.Env["AWS_DEFAULT_REGION"]; exists {
			eksConfig.Region = region
		}
	}

	klog.V(2).Infof("Parsed EKS config: cluster=%s, region=%s, profile=%s", 
		eksConfig.ClusterName, eksConfig.Region, eksConfig.Profile)

	return eksConfig, nil
}

// extractClusterInfoFromArgs 从命令行参数中提取集群信息
func (cp *ConfigParser) extractClusterInfoFromArgs(args []string, eksConfig *EKSAuthConfig) error {
	for i, arg := range args {
		switch arg {
		case "--cluster-name":
			if i+1 < len(args) {
				eksConfig.ClusterName = args[i+1]
			}
		case "--region":
			if i+1 < len(args) {
				eksConfig.Region = args[i+1]
			}
		case "--role-arn":
			if i+1 < len(args) {
				eksConfig.RoleARN = args[i+1]
			}
		}
	}

	// 验证必需的参数
	if eksConfig.ClusterName == "" {
		return NewEKSAuthError(ErrorTypeInvalidKubeconfig, "cluster name not found in exec args", nil)
	}

	return nil
}

// GetClusterEndpoint 从 kubeconfig 中获取集群端点
func (cp *ConfigParser) GetClusterEndpoint(kubeconfigContent []byte) (string, error) {
	config, err := clientcmd.Load(kubeconfigContent)
	if err != nil {
		return "", NewEKSAuthError(ErrorTypeInvalidKubeconfig, "failed to load kubeconfig", err)
	}

	// 获取当前上下文
	currentContext := config.CurrentContext
	if currentContext == "" && len(config.Contexts) > 0 {
		// 如果没有当前上下文，使用第一个
		for name := range config.Contexts {
			currentContext = name
			break
		}
	}

	if currentContext == "" {
		return "", NewEKSAuthError(ErrorTypeInvalidKubeconfig, "no context found in kubeconfig", nil)
	}

	context, exists := config.Contexts[currentContext]
	if !exists {
		return "", NewEKSAuthError(ErrorTypeInvalidKubeconfig, fmt.Sprintf("context %s not found", currentContext), nil)
	}

	cluster, exists := config.Clusters[context.Cluster]
	if !exists {
		return "", NewEKSAuthError(ErrorTypeInvalidKubeconfig, fmt.Sprintf("cluster %s not found", context.Cluster), nil)
	}

	return cluster.Server, nil
}

// ValidateEKSConfig 验证 EKS 配置的完整性
func (cp *ConfigParser) ValidateEKSConfig(config *EKSAuthConfig) error {
	if config == nil {
		return NewEKSAuthError(ErrorTypeInvalidKubeconfig, "EKS config is nil", nil)
	}

	if config.ClusterName == "" {
		return NewEKSAuthError(ErrorTypeInvalidKubeconfig, "cluster name is required", nil)
	}

	if config.ExecConfig == nil {
		return NewEKSAuthError(ErrorTypeInvalidKubeconfig, "exec config is required", nil)
	}

	if config.ExecConfig.Command == "" {
		return NewEKSAuthError(ErrorTypeInvalidKubeconfig, "exec command is required", nil)
	}

	return nil
}

// ExtractRegionFromClusterEndpoint 从集群端点中提取区域信息
func (cp *ConfigParser) ExtractRegionFromClusterEndpoint(endpoint string) string {
	// EKS 集群端点格式: https://ABC123.gr7.us-east-1.eks.amazonaws.com
	parts := strings.Split(endpoint, ".")
	if len(parts) >= 4 {
		for i, part := range parts {
			if part == "eks" && i > 0 {
				return parts[i-1]
			}
		}
	}
	return ""
}