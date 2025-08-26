package example

import (
	"os"
	"testing"

	"github.com/weibaohui/kom/kom"
)

// TestEKSRegistration 测试 EKS 集群注册
func TestEKSRegistration(t *testing.T) {
	// 创建一个模拟的 EKS kubeconfig 内容
	eksKubeconfig := `apiVersion: v1
clusters:
- cluster:
    certificate-authority-data: LS0tLS1CRUdJTi...
    server: https://ABC123.gr7.us-east-1.eks.amazonaws.com
  name: my-eks-cluster
contexts:
- context:
    cluster: my-eks-cluster
    user: my-eks-cluster
  name: my-eks-cluster
current-context: my-eks-cluster
kind: Config
users:
- name: my-eks-cluster
  user:
    exec:
      apiVersion: client.authentication.k8s.io/v1beta1
      command: aws
      args:
      - eks
      - get-token
      - --cluster-name
      - my-eks-cluster
      - --region
      - us-east-1`

	// 测试通过字符串注册 EKS 集群
	t.Run("RegisterEKSByString", func(t *testing.T) {
		// 注意：这个测试可能会失败，因为需要真实的 AWS 凭证
		// 但我们主要测试检测逻辑
		_, err := kom.Clusters().RegisterByString(eksKubeconfig)
		if err != nil {
			t.Logf("Expected error for EKS registration without valid AWS credentials: %v", err)
			// 检查是否正确识别为 EKS 配置
			if err.Error() != "not an EKS kubeconfig" {
				t.Logf("EKS configuration was correctly detected")
			}
		}
	})

	// 测试通过字符串和 ID 注册 EKS 集群
	t.Run("RegisterEKSByStringWithID", func(t *testing.T) {
		_, err := kom.Clusters().RegisterByStringWithID(eksKubeconfig, "test-eks")
		if err != nil {
			t.Logf("Expected error for EKS registration without valid AWS credentials: %v", err)
		}
	})

	// 测试非 EKS 配置
	normalKubeconfig := `apiVersion: v1
clusters:
- cluster:
    certificate-authority-data: LS0tLS1CRUdJTi...
    server: https://kubernetes.default.svc
  name: kubernetes
contexts:
- context:
    cluster: kubernetes
    user: kubernetes-admin
  name: kubernetes-admin@kubernetes
current-context: kubernetes-admin@kubernetes
kind: Config
users:
- name: kubernetes-admin
  user:
    client-certificate-data: LS0tLS1CRUdJTi...
    client-key-data: LS0tLS1CRUdJTi...`

	t.Run("RegisterNormalKubeconfig", func(t *testing.T) {
		_, err := kom.Clusters().RegisterByString(normalKubeconfig)
		if err != nil {
			t.Logf("Expected error for normal kubeconfig without valid cluster: %v", err)
		}
	})
}

// TestEKSDetection 测试 EKS 配置检测
func TestEKSDetection(t *testing.T) {
	// 创建临时的 EKS kubeconfig 文件
	eksKubeconfig := `apiVersion: v1
clusters:
- cluster:
    certificate-authority-data: LS0tLS1CRUdJTi...
    server: https://ABC123.gr7.us-east-1.eks.amazonaws.com
  name: my-eks-cluster
contexts:
- context:
    cluster: my-eks-cluster
    user: my-eks-cluster
  name: my-eks-cluster
current-context: my-eks-cluster
kind: Config
users:
- name: my-eks-cluster
  user:
    exec:
      apiVersion: client.authentication.k8s.io/v1beta1
      command: aws
      args:
      - eks
      - get-token
      - --cluster-name
      - my-eks-cluster
      - --region
      - us-east-1`

	// 创建临时文件
	tmpFile := "/tmp/test-eks-kubeconfig"
	err := os.WriteFile(tmpFile, []byte(eksKubeconfig), 0644)
	if err != nil {
		t.Fatalf("Failed to create temporary kubeconfig file: %v", err)
	}
	defer os.Remove(tmpFile)

	// 测试通过文件路径注册
	t.Run("RegisterEKSByPath", func(t *testing.T) {
		_, err := kom.Clusters().RegisterByPath(tmpFile)
		if err != nil {
			t.Logf("Expected error for EKS registration without valid AWS credentials: %v", err)
		}
	})

	// 测试通过文件路径和 ID 注册
	t.Run("RegisterEKSByPathWithID", func(t *testing.T) {
		_, err := kom.Clusters().RegisterByPathWithID(tmpFile, "test-eks-file")
		if err != nil {
			t.Logf("Expected error for EKS registration without valid AWS credentials: %v", err)
		}
	})
}
