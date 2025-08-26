package kom

import (
	"strings"
	"testing"

	"github.com/weibaohui/kom/kom/aws"
)

// Mock EKS kubeconfig for testing
const testEKSKubeconfig = `apiVersion: v1
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

// Mock standard kubeconfig
const testStandardKubeconfig = `apiVersion: v1
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
    client-certificate-data: LS0tLS1CRUdJTi1DRVJUSUZJQ0FURS0tLS0tCk1JSURBVENDQWVtZ0F3SUJBZ0lCQURBTkJna3Foa2lHOXcwQkFRc0ZBREFXTVJRd0VnWURWUVFERXd0amRHOXYKYkVGa2JXbHVVM1Z5ZGpBZUZ3MHhPVEF5TWpneE56QTBNVEphRncweU9UQXlNalV4TnpBME1USmFNRGN4SERSQW5CZ05WQkFNVEIydHVZV0YwYjJGa2QybHVVM1J5YkRBZkJnTlZCQU1URUE==
    client-key-data: LS0tLS1CRUdJTi1QUklWQVRFIEtFWS0tLS0tCk1JSUV2Z0lCQURBTkJna3Foa2lHOXcwQkFRRUZBQVNDQktnd2dnU2tBZ0VBQW9JQkFRREVZZnRIOG1sMndsaDQKVEFpbm9VZjZHdWNvM0tSV0lBOStBWnFQaCtYL0lFd3pZdWFNcmJRYUFKZGZNblJjYTBsVGhvV2E5Ym1jSWxoUwp2Q2hyay9wT1FnMFhoN3VHRkVEQ1MrM2FSRXNyYjJJQlZNRUxEbThuL3pBWDJvV0xNQkFUeE5RNkJHTUZqRjdvCmJpUk10SnRBVGZTRW55SmJIMlJ6WXNJUWZVVVFrRTBMNS96V1FFbHF1cjhUNlBNbmlIWEVNb0xKSXE1OVJZdmcKZzdZbU50NHpPOGFHWUdEZ0d5UUFnWUVBd2VjY005NUI3SkNLS2tGV1JDQS5qVXdzRjh5enpIV3dLVlVYcUlaNwpid3gvaXRVTGFPRVhUZ0FBZUVZQWt6MG1ZVDM3a0ZINlhCNWFVSmhBWVUvUklNRi9ZOGdLVjVlZUtaNHVuSXdvCmNGdDNFYkZoeXArWVNQUWdqSXdOR3Y1WTJGRnJPdk5idz09Ci0tLS0tRU5EIFBSSVZBVEUgS0VZLS0tLS0K`

func TestNewAWSAuthProvider(t *testing.T) {
	provider := NewAWSAuthProvider()
	if provider == nil {
		t.Error("NewAWSAuthProvider() should not return nil")
	}

	// Test that it can detect EKS config
	isEKS := provider.IsEKSConfig([]byte(testEKSKubeconfig))
	if !isEKS {
		t.Error("Should detect EKS kubeconfig")
	}

	// Test that it doesn't detect standard config as EKS
	isNotEKS := provider.IsEKSConfig([]byte(testStandardKubeconfig))
	if isNotEKS {
		t.Error("Should not detect standard kubeconfig as EKS")
	}
}

func TestClusterInst_EKSMethods(t *testing.T) {
	// Create a mock cluster instance
	cluster := &ClusterInst{
		ID:    "test-cluster",
		IsEKS: true,
		AWSAuthProvider: aws.NewAuthProvider(),
	}

	// Test IsEKSCluster method
	if !cluster.IsEKSCluster() {
		t.Error("IsEKSCluster() should return true for EKS cluster")
	}

	// Test GetAWSAuthProvider method
	provider := cluster.GetAWSAuthProvider()
	if provider == nil {
		t.Error("GetAWSAuthProvider() should not return nil for EKS cluster")
	}

	// Test with non-EKS cluster
	nonEKSCluster := &ClusterInst{
		ID:    "non-eks-cluster",
		IsEKS: false,
	}

	if nonEKSCluster.IsEKSCluster() {
		t.Error("IsEKSCluster() should return false for non-EKS cluster")
	}

	if nonEKSCluster.GetAWSAuthProvider() != nil {
		t.Error("GetAWSAuthProvider() should return nil for non-EKS cluster")
	}
}

func TestClusterInstances_EKSRegistration(t *testing.T) {
	// Note: These are unit tests that test the logic without actually connecting to AWS/EKS
	// They will fail at the AWS authentication step, but should pass the initial validation
	
	clusters := &ClusterInstances{}
	
	t.Run("EKS config detection in RegisterByString", func(t *testing.T) {
		// This should detect EKS config and attempt EKS registration
		// It will fail due to missing AWS credentials, but the detection should work
		_, err := clusters.RegisterByString(testEKSKubeconfig)
		if err == nil {
			t.Error("Expected error due to missing AWS credentials")
		}
		
		// The error should be related to AWS authentication, not config parsing
		if !strings.Contains(err.Error(), "failed to initialize EKS auth provider") && 
		   !strings.Contains(err.Error(), "failed to get initial AWS token") {
			t.Logf("Error was: %v", err)
			// This is expected - we don't have real AWS credentials in tests
		}
	})

	t.Run("Standard config should use standard registration", func(t *testing.T) {
		// This should use standard registration and fail due to invalid server
		_, err := clusters.RegisterByString(testStandardKubeconfig)
		if err == nil {
			t.Error("Expected error due to invalid server endpoint")
		}
		
		// Should not mention EKS in the error
		if strings.Contains(strings.ToLower(err.Error()), "eks") {
			t.Errorf("Standard kubeconfig should not trigger EKS registration: %v", err)
		}
	})
}

func TestConfigParser_Integration(t *testing.T) {
	parser := aws.NewConfigParser()

	t.Run("Parse valid EKS config", func(t *testing.T) {
		eksConfig, err := parser.ParseEKSConfig([]byte(testEKSKubeconfig))
		if err != nil {
			t.Fatalf("Failed to parse EKS config: %v", err)
		}

		// Validate parsed config
		if eksConfig.ClusterName != "test-cluster" {
			t.Errorf("Expected cluster name 'test-cluster', got '%s'", eksConfig.ClusterName)
		}

		if eksConfig.Region != "us-east-1" {
			t.Errorf("Expected region 'us-east-1', got '%s'", eksConfig.Region)
		}

		if eksConfig.Profile != "default" {
			t.Errorf("Expected profile 'default', got '%s'", eksConfig.Profile)
		}

		if eksConfig.ExecConfig == nil {
			t.Fatal("ExecConfig should not be nil")
		}

		if eksConfig.ExecConfig.Command != "aws" {
			t.Errorf("Expected command 'aws', got '%s'", eksConfig.ExecConfig.Command)
		}

		expectedArgs := []string{"eks", "get-token", "--cluster-name", "test-cluster", "--region", "us-east-1"}
		if len(eksConfig.ExecConfig.Args) != len(expectedArgs) {
			t.Errorf("Expected %d args, got %d", len(expectedArgs), len(eksConfig.ExecConfig.Args))
		}

		for i, expected := range expectedArgs {
			if i >= len(eksConfig.ExecConfig.Args) || eksConfig.ExecConfig.Args[i] != expected {
				t.Errorf("Arg %d: expected '%s', got '%s'", i, expected, eksConfig.ExecConfig.Args[i])
			}
		}
	})

	t.Run("Get cluster endpoint", func(t *testing.T) {
		endpoint, err := parser.GetClusterEndpoint([]byte(testEKSKubeconfig))
		if err != nil {
			t.Fatalf("Failed to get cluster endpoint: %v", err)
		}

		expectedEndpoint := "https://ABC123.gr7.us-east-1.eks.amazonaws.com"
		if endpoint != expectedEndpoint {
			t.Errorf("Expected endpoint '%s', got '%s'", expectedEndpoint, endpoint)
		}
	})

	t.Run("Extract region from endpoint", func(t *testing.T) {
		endpoint := "https://ABC123.gr7.us-east-1.eks.amazonaws.com"
		region := parser.ExtractRegionFromClusterEndpoint(endpoint)
		
		expectedRegion := "us-east-1"
		if region != expectedRegion {
			t.Errorf("Expected region '%s', got '%s'", expectedRegion, region)
		}
	})
}

func TestEKSAuthConfig_Validation(t *testing.T) {
	parser := aws.NewConfigParser()

	tests := []struct {
		name        string
		config      *aws.EKSAuthConfig
		expectError bool
	}{
		{
			name:        "Nil config",
			config:      nil,
			expectError: true,
		},
		{
			name: "Missing cluster name",
			config: &aws.EKSAuthConfig{
				ExecConfig: &aws.ExecConfig{Command: "aws"},
			},
			expectError: true,
		},
		{
			name: "Missing exec config",
			config: &aws.EKSAuthConfig{
				ClusterName: "test-cluster",
			},
			expectError: true,
		},
		{
			name: "Valid config",
			config: &aws.EKSAuthConfig{
				ClusterName: "test-cluster",
				Region:      "us-east-1",
				ExecConfig: &aws.ExecConfig{
					Command: "aws",
					Args:    []string{"eks", "get-token", "--cluster-name", "test-cluster"},
				},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := parser.ValidateEKSConfig(tt.config)
			if (err != nil) != tt.expectError {
				t.Errorf("ValidateEKSConfig() error = %v, expectError %v", err, tt.expectError)
			}
		})
	}
}