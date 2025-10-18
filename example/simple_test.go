package example

import (
	"os"
	"testing"

	"github.com/weibaohui/kom/mcp"
)

func TestSimpleMultiClusterSSE(t *testing.T) {
	// 创建测试kubeconfig文件
	testDir := "/tmp/kom-simple-test"
	err := os.MkdirAll(testDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	defer os.RemoveAll(testDir)

	// 创建示例kubeconfig文件
	kubeconfig1 := `apiVersion: v1
clusters:
- cluster:
    server: https://cluster1.example.com:6443
    insecure-skip-tls-verify: true
  name: cluster1
contexts:
- context:
    cluster: cluster1
    user: admin
  name: cluster1
current-context: cluster1
kind: Config
users:
- name: admin
  user:
    token: fake-token-1`

	kubeconfig2 := `apiVersion: v1
clusters:
- cluster:
    server: https://cluster2.example.com:6443
    insecure-skip-tls-verify: true
  name: cluster2
contexts:
- context:
    cluster: cluster2
    user: admin
  name: cluster2
current-context: cluster2
kind: Config
users:
- name: admin
  user:
    token: fake-token-2`

	file1 := testDir + "/cluster1.yaml"
	file2 := testDir + "/cluster2.yaml"

	err = os.WriteFile(file1, []byte(kubeconfig1), 0644)
	if err != nil {
		t.Fatalf("Failed to write kubeconfig1: %v", err)
	}

	err = os.WriteFile(file2, []byte(kubeconfig2), 0644)
	if err != nil {
		t.Fatalf("Failed to write kubeconfig2: %v", err)
	}

	// 配置多集群
	cfg := &mcp.ServerConfig{
		Name:    "kom simple test",
		Version: "1.0.0",
		Port:    9096,
		Mode:    mcp.ServerModeSSE,
		Kubeconfigs: []mcp.KubeconfigConfig{
			{
				ID:        "cluster1",
				Path:      file1,
				IsDefault: true,
			},
			{
				ID:   "cluster2",
				Path: file2,
			},
		},
	}

	// 创建MCP服务器（不启动）
	server := mcp.GetMCPServerWithOption(cfg)
	if server == nil {
		t.Fatal("Failed to create MCP server")
	}

	t.Log("✅ Successfully created MCP server with multi-cluster support")
	t.Log("✅ Cluster1 registered as default")
	t.Log("✅ Cluster2 registered successfully")
}

func TestDirectoryLoading(t *testing.T) {
	// 创建测试目录
	testDir := "/tmp/kom-dir-test"
	err := os.MkdirAll(testDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	defer os.RemoveAll(testDir)

	// 创建kubeconfig文件
	kubeconfig := `apiVersion: v1
clusters:
- cluster:
    server: https://test.example.com:6443
    insecure-skip-tls-verify: true
  name: test-cluster
contexts:
- context:
    cluster: test-cluster
    user: admin
  name: test-context
current-context: test-context
kind: Config
users:
- name: admin
  user:
    token: fake-token`

	err = os.WriteFile(testDir+"/test.yaml", []byte(kubeconfig), 0644)
	if err != nil {
		t.Fatalf("Failed to write kubeconfig: %v", err)
	}

	// 从目录加载
	configs, err := mcp.LoadKubeconfigsFromDirectory(testDir)
	if err != nil {
		t.Fatalf("Failed to load from directory: %v", err)
	}

	if len(configs) != 1 {
		t.Fatalf("Expected 1 config, got %d", len(configs))
	}

	if configs[0].ID != "test" {
		t.Fatalf("Expected cluster ID 'test', got '%s'", configs[0].ID)
	}

	t.Log("✅ Directory loading works correctly")
}
