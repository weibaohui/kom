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
