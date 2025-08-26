package example

import (
	"testing"

	"github.com/weibaohui/kom/kom"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd/api"
)

func TestRegisterByConfigWithExecAuth(t *testing.T) {
	t.Run("AWS EKS exec config", func(t *testing.T) {
		// 创建一个模拟的 AWS EKS exec 配置
		config := &rest.Config{
			Host: "https://test-cluster.eks.us-east-1.amazonaws.com",
			ExecProvider: &api.ExecConfig{
				Command: "aws",
				Args: []string{
					"eks", "get-token",
					"--cluster-name", "test-cluster",
					"--region", "us-east-1",
				},
				Env: []api.ExecEnvVar{
					{Name: "AWS_PROFILE", Value: "default"},
				},
			},
		}

		// 注意：这个测试可能会失败，因为需要真实的 AWS 凭证
		_, err := kom.Clusters().RegisterByConfigWithID(config, "test-exec-cluster")
		if err != nil {
			t.Logf("Expected error for exec authentication without valid AWS credentials: %v", err)
		} else {
			t.Log("Successfully registered cluster with exec authentication")
		}
	})

	t.Run("Non-AWS exec config", func(t *testing.T) {
		// 创建一个模拟的非 AWS exec 配置
		config := &rest.Config{
			Host: "https://test-cluster.example.com",
			ExecProvider: &api.ExecConfig{
				Command: "kubectl-auth-plugin",
				Args: []string{
					"--cluster", "test-cluster",
				},
			},
		}

		// 这应该跳过 AWS 特定的处理
		_, err := kom.Clusters().RegisterByConfigWithID(config, "test-non-aws-exec")
		if err != nil {
			t.Logf("Expected error for non-AWS exec command: %v", err)
		} else {
			t.Log("Successfully registered cluster with non-AWS exec authentication")
		}
	})

	t.Run("Standard config without exec", func(t *testing.T) {
		// 创建一个标准的配置（无 exec 提供者）
		config := &rest.Config{
			Host:        "https://test-cluster.example.com",
			BearerToken: "test-token",
		}

		kubectl, err := kom.Clusters().RegisterByConfigWithID(config, "test-standard-cluster")
		if err != nil {
			t.Errorf("Failed to register standard cluster: %v", err)
		} else {
			t.Log("Successfully registered standard cluster")
			if kubectl == nil {
				t.Error("kubectl should not be nil")
			}
		}
	})
}

func TestExecAuthIntegration(t *testing.T) {
	t.Run("Check cluster exists after registration", func(t *testing.T) {
		config := &rest.Config{
			Host:        "https://integration-test.example.com",
			BearerToken: "test-token",
		}

		id := "integration-test-cluster"
		kubectl1, err1 := kom.Clusters().RegisterByConfigWithID(config, id)
		if err1 != nil {
			t.Fatalf("Failed to register cluster: %v", err1)
		}

		// 再次注册相同的集群，应该返回已存在的实例
		kubectl2, err2 := kom.Clusters().RegisterByConfigWithID(config, id)
		if err2 != nil {
			t.Fatalf("Failed to get existing cluster: %v", err2)
		}

		if kubectl1 != kubectl2 {
			t.Error("Should return the same kubectl instance for the same cluster ID")
		}

		t.Log("Successfully verified cluster caching behavior")
	})
}