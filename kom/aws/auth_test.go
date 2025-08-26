package aws

import (
	"context"
	"strings"
	"testing"
	"time"
)

// Mock EKS kubeconfig for testing
const mockEKSKubeconfig = `apiVersion: v1
clusters:
- cluster:
    certificate-authority-data: LS0tLS1CRUdJTi1DRVJUSUZJQ0FURS0tLS0tCk1JSURCVENDQWUyZ0F3SUJBZ0lJZHUyY3pZdFJaVUV3RFFZSktvWklodmNOQVFFTEJRQXdGVEVUTUJFR0ExVUUKQXhNS2EzVmlaWEp1WlhSbGN6QWVGdzB4T1RBeU1qZ3hOekE0TVRSYUZ3MHlPVEF5TWpVeE56QTRNVFJhTUJVeApFekFSQmdOVkJBTVRDbXQxWW1WeWJtVjBaWE13Z1Z3d0RRWUpLb1pJaHZjTkFRRUJCUUFEZ0VzQU1JSUJDZ0tCCmdFa0Foa2Q5NGV1UVFZTTBhL09kUEcwWEsxcjlaMVVRbGNERWtVWmZMMnlJdU5BMEgrY2dJbkcrQTlKdUg3dGwKb0lzV2hRWWVkN0xXalV0aGJaRHRQSXBFUVBONUtGOWtHWFJ1a2pDTGlCMm0zNHNkbjlRK1Y1M3htTGNONDhRbQpGcXptaGdmR3ZJUXFiVE5WYW1FdjBra3pxMUxQV3B1MkJtVUJORVh3SUJBd0RRWUpLb1pJaHZjTkFRRUxCUUFECmdZRUF2cG1vUWNRTGdwVG04Z01UQll0SWVtbFBYQlAySVhuK09SdHJkQ0dUK042czZpRDJzN1RQdDVqUDBhYVUKZVorU0ZUVzFWcFVnNmZXSWZPa1Y5VVlQUUdxckZXRDROUGdPUmFWVHJrRmtrVHZSaFA0K2xRcTA4OENVUGpxawppRDBJSFEzaU0rNzh4dkttOGE3U3N4Nm1ZaW1pcU4xY1lWRHd5TjJDbk1VPQotLS0tLUVORCBDRVJUSUZJQ0FURS0tLS0tCg==
    server: https://ABC123.gr7.us-east-1.eks.amazonaws.com
  name: test-eks-cluster
contexts:
- context:
    cluster: test-eks-cluster
    user: test-eks-user
  name: test-eks-context
current-context: test-eks-context
users:
- name: test-eks-user
  user:
    exec:
      apiVersion: client.authentication.k8s.io/v1beta1
      command: aws
      args:
      - eks
      - get-token
      - --cluster-name
      - test-cluster
      - --region
      - us-east-1
      env:
      - name: AWS_PROFILE
        value: default`

// Mock standard kubeconfig for testing
const mockStandardKubeconfig = `apiVersion: v1
clusters:
- cluster:
    certificate-authority-data: LS0tLS1CRUdJTi1DRVJUSUZJQ0FURS0tLS0tCk1JSURCVENDQWUyZ0F3SUJBZ0lJZHUyY3pZdFJaVUV3RFFZSktvWklodmNOQVFFTEJRQXdGVEVUTUJFR0ExVUUKQXhNS2EzVmlaWEp1WlhSbGN6QWVGdzB4T1RBeU1qZ3hOekE0TVRSYUZ3MHlPVEF5TWpVeE56QTRNVFJhTUJVeApFekFSQmdOVkJBTVRDbXQxWW1WeWJtVjBaWE13Z1Z3d0RRWUpLb1pJaHZjTkFRRUJCUUFEZ0VzQU1JSUJDZ0tCCmdFa0Foa2Q5NGV1UVFZTTBhL09kUEcwWEsxcjlaMVVRbGNERWtVWmZMMnlJdU5BMEgrY2dJbkcrQTlKdUg3dGwKb0lzV2hRWWVkN0xXalV0aGJaRHRQSXBFUVBONUtGOWtHWFJ1a2pDTGlCMm0zNHNkbjlRK1Y1M3htTGNONDhRbQpGcXptaGdmR3ZJUXFiVE5WYW1FdjBra3pxMUxQV3B1MkJtVUJORVh3SUJBd0RRWUpLb1pJaHZjTkFRRUxCUUFECmdZRUF2cG1vUWNRTGdwVG04Z01UQll0SWVtbFBYQlAySVhuK09SdHJkQ0dUK042czZpRDJzN1RQdDVqUDBhYVUKZVorU0ZUVzFWcFVnNmZXSWZPa1Y5VVlQUUdxckZXRDROUGdPUmFWVHJrRmtrVHZSaFA0K2xRcTA4OENVUGpxawppRDBJSFEzaU0rNzh4dkttOGE3U3N4Nm1ZaW1pcU4xY1lWRHd5TjJDbk1VPQotLS0tLUVORCBDRVJUSUZJQ0FURS0tLS0tCg==
    server: https://kubernetes.example.com:6443
  name: standard-cluster
contexts:
- context:
    cluster: standard-cluster
    user: standard-user
  name: standard-context
current-context: standard-context
users:
- name: standard-user
  user:
    client-certificate-data: LS0tLS1CRUdJTi1DRVJUSUZJQ0FURS0tLS0tCk1JSURBVENDQWVtZ0F3SUJBZ0lCQURBTkJna3Foa2lHOXcwQkFRc0ZBREFXTVJRd0VnWURWUVFERXd0amRHOXYKYkVGa2JXbHVVMlZ5ZGpBZUZ3MHhPVEF5TWpneE56QTBNVEphRncweU9UQXlNalV4TnpBME1USmFNRGN4SURSQW5CZ05WQkFNVEIydHVZV0YwYjJGa2QybHVVM1J5YkRBZkJnTlZCQU1URUE==
    client-key-data: LS0tLS1CRUdJTi1QUklWQVRFIEtFWS0tLS0tCk1JSUV2Z0lCQURBTkJna3Foa2lHOXcwQkFRRUZBQVNDQktnd2dnU2tBZ0VBQW9JQkFRREVZZnRIOG1sMndsaDQKVEFpbm9VZjZHdWNvM0tSV0lBOStBWnFQaCtYL0lFd3pZdWFNcmJRYUFKZGZNblJjYTBsVGhvV2E5Ym1jSWxoUwp2Q2hyay9wT1FnMFhoN3VHRkVEQ1MrM2FSRXNyYjJJQlZNRUxEbThuL3pBWDJvV0xNQkFUeE5RNkJHTUZqRjdvCmJpUk10SnRBVGZTRW55SmJIMlJ6WXNJUWZVVVFrRTBMNS96V1FFbHF1cjhUNlBNbmlIWEVNb0xKSXE1OVJZdmcKZzdZbU50NHpPOGFHWUdEZ0d5UUFnWUVBd2VjY005NUI3SkNLS2tGV1JDQS5qVXdzRjh5enpIV3dLVlVYcUlaNwpid3gvaXRVTGFPRVhUZ0FBZUVZQWt6MG1ZVDM3a0ZINlhCNWFVSmhBWVUvUklNRi9ZOGdLVjVlZUtaNHVuSXdvCmNGdDNFYkZoeXArWVNQUWdqSXdOR3Y1WTJGRnJPdk5idz09Ci0tLS0tRU5EIFBSSVZBVEUgS0VZLS0tLS0K`

func TestConfigParser_IsEKSConfig(t *testing.T) {
	parser := NewConfigParser()

	tests := []struct {
		name       string
		kubeconfig string
		expected   bool
	}{
		{
			name:       "Valid EKS config",
			kubeconfig: mockEKSKubeconfig,
			expected:   true,
		},
		{
			name:       "Standard kubeconfig",
			kubeconfig: mockStandardKubeconfig,
			expected:   false,
		},
		{
			name:       "Invalid kubeconfig",
			kubeconfig: "invalid yaml",
			expected:   false,
		},
		{
			name:       "Empty kubeconfig",
			kubeconfig: "",
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.IsEKSConfig([]byte(tt.kubeconfig))
			if result != tt.expected {
				t.Errorf("IsEKSConfig() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestConfigParser_ParseEKSConfig(t *testing.T) {
	parser := NewConfigParser()

	t.Run("Valid EKS config", func(t *testing.T) {
		config, err := parser.ParseEKSConfig([]byte(mockEKSKubeconfig))
		if err != nil {
			t.Fatalf("ParseEKSConfig() error = %v", err)
		}

		if config.ClusterName != "test-cluster" {
			t.Errorf("ClusterName = %v, expected %v", config.ClusterName, "test-cluster")
		}

		if config.Region != "us-east-1" {
			t.Errorf("Region = %v, expected %v", config.Region, "us-east-1")
		}

		if config.Profile != "default" {
			t.Errorf("Profile = %v, expected %v", config.Profile, "default")
		}

		if config.ExecConfig == nil {
			t.Error("ExecConfig should not be nil")
		}

		if config.ExecConfig.Command != "aws" {
			t.Errorf("ExecConfig.Command = %v, expected %v", config.ExecConfig.Command, "aws")
		}
	})

	t.Run("Invalid kubeconfig", func(t *testing.T) {
		_, err := parser.ParseEKSConfig([]byte("invalid yaml"))
		if err == nil {
			t.Error("ParseEKSConfig() should return error for invalid yaml")
		}
	})

	t.Run("Standard kubeconfig", func(t *testing.T) {
		_, err := parser.ParseEKSConfig([]byte(mockStandardKubeconfig))
		if err == nil {
			t.Error("ParseEKSConfig() should return error for non-EKS config")
		}
	})
}

func TestTokenCache(t *testing.T) {
	cache := &TokenCache{}

	t.Run("Initial state", func(t *testing.T) {
		if cache.IsValid() {
			t.Error("TokenCache should not be valid initially")
		}

		token, expiry := cache.GetToken()
		if token != "" {
			t.Error("Initial token should be empty")
		}
		if !expiry.IsZero() {
			t.Error("Initial expiry should be zero time")
		}
	})

	t.Run("Set and get token", func(t *testing.T) {
		testToken := "test-token"
		testExpiry := time.Now().Add(time.Hour)

		cache.SetToken(testToken, testExpiry)

		if !cache.IsValid() {
			t.Error("TokenCache should be valid after setting token")
		}

		token, expiry := cache.GetToken()
		if token != testToken {
			t.Errorf("GetToken() token = %v, expected %v", token, testToken)
		}
		if !expiry.Equal(testExpiry) {
			t.Errorf("GetToken() expiry = %v, expected %v", expiry, testExpiry)
		}
	})

	t.Run("Expired token", func(t *testing.T) {
		pastTime := time.Now().Add(-time.Hour)
		cache.SetToken("expired-token", pastTime)

		if cache.IsValid() {
			t.Error("TokenCache should not be valid for expired token")
		}
	})

	t.Run("Clear token", func(t *testing.T) {
		cache.SetToken("test-token", time.Now().Add(time.Hour))
		cache.ClearToken()

		if cache.IsValid() {
			t.Error("TokenCache should not be valid after clearing")
		}

		token, expiry := cache.GetToken()
		if token != "" {
			t.Error("Token should be empty after clearing")
		}
		if !expiry.IsZero() {
			t.Error("Expiry should be zero time after clearing")
		}
	})
}

func TestExecExecutor_ValidateCommand(t *testing.T) {
	executor := NewExecExecutor()

	tests := []struct {
		name        string
		execConfig  *ExecConfig
		expectError bool
	}{
		{
			name:        "Nil config",
			execConfig:  nil,
			expectError: true,
		},
		{
			name: "Empty command",
			execConfig: &ExecConfig{
				Command: "",
				Args:    []string{},
			},
			expectError: true,
		},
		{
			name: "Valid command",
			execConfig: &ExecConfig{
				Command: "echo",
				Args:    []string{"test"},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := executor.ValidateCommand(tt.execConfig)
			if (err != nil) != tt.expectError {
				t.Errorf("ValidateCommand() error = %v, expectError %v", err, tt.expectError)
			}
		})
	}
}

func TestExecExecutor_BuildGetTokenCommand(t *testing.T) {
	executor := NewExecExecutor()

	tests := []struct {
		name        string
		clusterName string
		region      string
		profile     string
		roleARN     string
	}{
		{
			name:        "Basic command",
			clusterName: "test-cluster",
			region:      "us-east-1",
			profile:     "",
			roleARN:     "",
		},
		{
			name:        "With profile",
			clusterName: "test-cluster",
			region:      "us-east-1",
			profile:     "dev",
			roleARN:     "",
		},
		{
			name:        "With role ARN",
			clusterName: "test-cluster",
			region:      "us-east-1",
			profile:     "",
			roleARN:     "arn:aws:iam::123456789012:role/EKSRole",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			execConfig := executor.BuildGetTokenCommand(tt.clusterName, tt.region, tt.profile, tt.roleARN)

			if execConfig.Command != "aws" {
				t.Errorf("Command = %v, expected %v", execConfig.Command, "aws")
			}

			argsStr := strings.Join(execConfig.Args, " ")
			if !strings.Contains(argsStr, "eks get-token") {
				t.Error("Args should contain 'eks get-token'")
			}

			if !strings.Contains(argsStr, "--cluster-name "+tt.clusterName) {
				t.Errorf("Args should contain cluster name %v", tt.clusterName)
			}

			if tt.region != "" && !strings.Contains(argsStr, "--region "+tt.region) {
				t.Errorf("Args should contain region %v", tt.region)
			}

			if tt.roleARN != "" && !strings.Contains(argsStr, "--role-arn "+tt.roleARN) {
				t.Errorf("Args should contain role ARN %v", tt.roleARN)
			}

			if tt.profile != "" {
				if execConfig.Env["AWS_PROFILE"] != tt.profile {
					t.Errorf("AWS_PROFILE = %v, expected %v", execConfig.Env["AWS_PROFILE"], tt.profile)
				}
			}
		})
	}
}

func TestAuthProvider_IsEKSConfig(t *testing.T) {
	provider := NewAuthProvider()

	tests := []struct {
		name       string
		kubeconfig string
		expected   bool
	}{
		{
			name:       "Valid EKS config",
			kubeconfig: mockEKSKubeconfig,
			expected:   true,
		},
		{
			name:       "Standard kubeconfig",
			kubeconfig: mockStandardKubeconfig,
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := provider.IsEKSConfig([]byte(tt.kubeconfig))
			if result != tt.expected {
				t.Errorf("IsEKSConfig() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestEKSAuthError(t *testing.T) {
	t.Run("Basic error", func(t *testing.T) {
		err := NewEKSAuthError("TestError", "test message", nil)
		if err.Type != "TestError" {
			t.Errorf("Type = %v, expected %v", err.Type, "TestError")
		}
		if err.Message != "test message" {
			t.Errorf("Message = %v, expected %v", err.Message, "test message")
		}
		if err.Error() != "test message" {
			t.Errorf("Error() = %v, expected %v", err.Error(), "test message")
		}
	})

	t.Run("Error with cause", func(t *testing.T) {
		cause := context.DeadlineExceeded
		err := NewEKSAuthError("TestError", "test message", cause)
		expectedMsg := "test message: " + cause.Error()
		if err.Error() != expectedMsg {
			t.Errorf("Error() = %v, expected %v", err.Error(), expectedMsg)
		}
	})
}

func TestConfigParser_ExtractRegionFromClusterEndpoint(t *testing.T) {
	parser := NewConfigParser()

	tests := []struct {
		name     string
		endpoint string
		expected string
	}{
		{
			name:     "Valid EKS endpoint",
			endpoint: "https://ABC123.gr7.us-east-1.eks.amazonaws.com",
			expected: "us-east-1",
		},
		{
			name:     "Different region",
			endpoint: "https://DEF456.gr7.eu-west-1.eks.amazonaws.com",
			expected: "eu-west-1",
		},
		{
			name:     "Invalid endpoint",
			endpoint: "https://invalid.endpoint.com",
			expected: "",
		},
		{
			name:     "Empty endpoint",
			endpoint: "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.ExtractRegionFromClusterEndpoint(tt.endpoint)
			if result != tt.expected {
				t.Errorf("ExtractRegionFromClusterEndpoint() = %v, expected %v", result, tt.expected)
			}
		})
	}
}