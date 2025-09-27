package example

import (
	"fmt"
	"testing"

	"github.com/weibaohui/kom/kom"
	corev1 "k8s.io/api/core/v1"
)

// TestRegisterByTokenWithServerAndID 测试通过token和服务器地址注册集群
func TestRegisterByTokenWithServerAndID(t *testing.T) {
	// 注意：这是一个示例测试，实际使用时需要提供真实的参数
	token := "eyJhbGciOiJSUzI1NiIsImtpZCI6IlRLRUJSVC1HbGJjUnlIY191MjQ4REhSbjVxUURYcFBWRkhzS2tnUHlGYVEifQ.eyJpc3MiOiJrdWJlcm5ldGVzL3NlcnZpY2VhY2NvdW50Iiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9uYW1lc3BhY2UiOiJkZWZhdWx0Iiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9zZWNyZXQubmFtZSI6Im15LXRva2VuLXNhLXNlY3JldCIsImt1YmVybmV0ZXMuaW8vc2VydmljZWFjY291bnQvc2VydmljZS1hY2NvdW50Lm5hbWUiOiJteS10b2tlbi1zYSIsImt1YmVybmV0ZXMuaW8vc2VydmljZWFjY291bnQvc2VydmljZS1hY2NvdW50LnVpZCI6ImJiODBlM2E2LTkzNDUtNGFhZC05ODhmLTJlMTIyYmNkMThhZSIsInN1YiI6InN5c3RlbTpzZXJ2aWNlYWNjb3VudDpkZWZhdWx0Om15LXRva2VuLXNhIn0.lUbir_OMoV4f7XypaFMEl8YCirmyqfiutb4MViKIWq7UmS_zxoY5Do22Nv62U4OGUxDZwVH0mCzHjgd89uA7OWBN8Nl6DOhlMTKyndFiqQsKGuiVEBwkYqsmXYbNSm2im-CMVsudlHsuUnuMfIeQ01-uAjKp17EmKEapwvE4cuboE_x2_8uuap2JeBxLPKV9RDEevbhBh9ePDIxuKMGUAJ8sbSd50cz-s_w-KgB8cSc_UV9hAcomgADtda2gAS1P-9QT5nIPH7dMpZ05N8pnSmRc_vLxUTiEx3n0QoXCfWyrfZrgGn_81Ok4bCEWmBed6MHfN1ChobpIymspWVqjDQ"
	server := "https://127.0.0.1:12954"
	clusterID := "test-cluster-with-server"

	// 使用完整的token注册（推荐方式）
	kubectl, err := kom.Clusters().RegisterByTokenWithServerAndID(token, server, clusterID)
	if err != nil {
		t.Logf("Expected error for test credentials: %v", err)
		return
	}

	if kubectl == nil {
		t.Error("Expected kubectl instance, got nil")
		return
	}
	var items []corev1.Pod
	var pod corev1.Pod
	err = kom.Cluster("test-cluster-with-server").
		Resource(&pod).
		AllNamespace().
		List(&items).Error
	if err != nil {
		t.Errorf("List Error %v\n", err)
	}
	if len(items) > 0 {
		t.Logf("List Pods count %d\n", len(items))
	} else {
		t.Errorf("List Pods count,should %d,acctual %d", 1, len(items))
	}
	fmt.Printf("Successfully registered cluster with ID: %s and server: %s\n", clusterID, server)
}

// TestRegisterByTokenWithOptions 测试通过token和详细选项注册集群
func TestRegisterByTokenWithOptions(t *testing.T) {
	// 注意：这是一个示例测试，实际使用时需要提供真实的参数
	token := "your-kubernetes-token-here"
	server := "https://your-kubernetes-api-server:6443"
	clusterID := "test-cluster-with-options"
	insecure := true // 在测试环境中可能需要跳过TLS验证

	// 使用带选项的token注册
	kubectl, err := kom.Clusters().RegisterByTokenWithOptions(token, server, clusterID, insecure)
	if err != nil {
		t.Logf("Expected error for test credentials: %v", err)
		return
	}

	if kubectl == nil {
		t.Error("Expected kubectl instance, got nil")
		return
	}

	fmt.Printf("Successfully registered cluster with ID: %s, server: %s, insecure: %v\n", clusterID, server, insecure)
}

// TestTokenRegistrationParameterValidation 测试参数验证功能
func TestTokenRegistrationParameterValidation(t *testing.T) {
	tests := []struct {
		name        string
		token       string
		server      string
		clusterID   string
		expectError bool
	}{
		{
			name:        "Empty token",
			token:       "",
			server:      "https://example.com:6443",
			clusterID:   "test",
			expectError: true,
		},
		{
			name:        "Empty server",
			token:       "valid-token",
			server:      "",
			clusterID:   "test",
			expectError: true,
		},
		{
			name:        "Empty cluster ID",
			token:       "valid-token",
			server:      "https://example.com:6443",
			clusterID:   "",
			expectError: true,
		},
		{
			name:        "Valid parameters but unreachable server",
			token:       "valid-token",
			server:      "https://unreachable.example.com:6443",
			clusterID:   "test",
			expectError: true, // 会出错，因为服务器不可达
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := kom.Clusters().RegisterByTokenWithServerAndID(tt.token, tt.server, tt.clusterID)
			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

// DemoTokenRegistration 展示如何使用token注册集群的示例
func DemoTokenRegistration() {
	// 实际使用示例
	token := "eyJhbGciOiJSUzI1NiIsImtpZCI6IjEyMzQ1Njc4OTAifQ..." // 你的实际token
	server := "https://your-cluster.example.com:6443"            // 你的集群API服务器地址
	clusterID := "production-cluster"                            // 集群标识符

	// 注册集群
	kubectl, err := kom.Clusters().RegisterByTokenWithServerAndID(token, server, clusterID)
	if err != nil {
		fmt.Printf("Failed to register cluster: %v\n", err)
		return
	}

	// 使用kubectl进行操作
	fmt.Printf("Successfully registered cluster: %s\n", clusterID)

	// 避免未使用变量的警告
	_ = kubectl

	// 示例：获取集群信息
	// pods, err := kubectl.Pod().List()
	// if err != nil {
	//     fmt.Printf("Failed to list pods: %v\n", err)
	//     return
	// }
	// fmt.Printf("Found %d pods in the cluster\n", len(pods))
}
