package aws

import (
	"fmt"
	"strings"

	"k8s.io/client-go/tools/clientcmd"
)

// ConfigParser EKS 配置解析器
type ConfigParser struct{}

// NewConfigParser 创建新的配置解析器
func NewConfigParser() *ConfigParser {
	return &ConfigParser{}
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
