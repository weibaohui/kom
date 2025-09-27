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
	token := "<YOUR_TOKEN>"
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
