package example

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/weibaohui/kom/kom"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// TestMain 是测试的入口函数
func TestMain(m *testing.M) {
	// 初始化操作
	fmt.Println("Initializing test environment...")
	// 在这里可以设置数据库连接、启动服务、创建临时文件等

	Connect()

	// 创建测试必须得Pod，后面不会再创建了
	InitTestDeploy()
	// 调用 m.Run() 运行所有测试
	exitCode := m.Run()

	// 清理操作
	fmt.Println("Cleaning up test environment...")
	// 在这里可以关闭数据库连接、删除临时文件等
	CleanTestDeploy()
	// 退出程序
	os.Exit(exitCode)
}

// 每 2 秒检查一次，超时设定为 60 秒
var interval = 2 * time.Second
var timeout = 60 * time.Second

func InitTestDeploy() {
	yaml := `apiVersion: v1
kind: Pod
metadata:
  name: random
  namespace: default
  labels:
    "x": "y"
    "app": "random"
spec:
  containers:
  - args:
    - |
      mkdir -p /var/log;
      while true; do
        random_char="A$(cat /dev/urandom | tr -dc 'a-zA-Z0-9' | head -c 1)";
        echo $random_char | tee -a /var/log/random_a.log;
        sleep 5;
      done
    command:
    - /bin/sh
    - -c
    image: alpine
    name: random
`
	result := kom.DefaultCluster().Applier().Apply(yaml)
	for _, s := range result {
		fmt.Printf("%s\n", s)
	}

	// if utils.WaitUntil(checkCondition, interval, timeout) {
	// 	fmt.Println("Check succeeded, main process exiting.")
	// } else {
	// 	fmt.Println("Check failed due to timeout.")
	// }

	// fmt.Println("Stopped checking condition.")

}
func CleanTestDeploy() {
	kom.DefaultCluster().
		Resource(&corev1.Pod{}).
		Name("random").
		Namespace("default").Delete()
}

// 定义一个函数，用于检查条件
func checkCondition() bool {
	fmt.Println("Checking condition at", time.Now())

	var pod corev1.Pod
	err := kom.DefaultCluster().Resource(&pod).Namespace("default").
		Name("random").Get(&pod).Error
	if err != nil {
		return false
	}
	if pod.Status.Phase == "Running" {
		fmt.Println("pod is running at", time.Now())
		return true
	}
	return false
}

func TestCrdExample(t *testing.T) {
	yaml := `apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: crontabs.stable.example.com
spec:
  group: stable.example.com
  versions:
    - name: v1
      served: true
      storage: true
      schema:
        openAPIV3Schema:
          type: object
          properties:
            spec:
              type: object
              properties:
                cronSpec:
                  type: string
                image:
                  type: string
                replicas:
                  type: integer
  scope: Namespaced
  names:
    plural: crontabs
    singular: crontab
    kind: CronTab
`

	t.Run("Apply CRD", func(t *testing.T) {
		result := kom.DefaultCluster().Applier().Apply(yaml)
		for _, str := range result {
			fmt.Println(str)
		}
	})

	t.Run("Create CR", func(t *testing.T) {
		time.Sleep(10 * time.Second)
		var crontab = unstructured.Unstructured{
			Object: map[string]interface{}{
				"apiVersion": "stable.example.com/v1",
				"kind":       "CronTab",
				"metadata": map[string]interface{}{
					"name":      "test-crontab",
					"namespace": "default",
				},
				"spec": map[string]interface{}{
					"cronSpec": "* * * * */8",
					"image":    "test-crontab-image",
				},
			},
		}

		err := kom.DefaultCluster().
			WithContext(context.TODO()).
			CRD("stable.example.com", "v1", "CronTab").
			Name(crontab.GetName()).
			Namespace(crontab.GetNamespace()).
			Create(&crontab).Error
		if err != nil {
			t.Logf("CRD Create error: %v\n", err)
		}
	})

	t.Run("Get CR", func(t *testing.T) {
		var crontab unstructured.Unstructured
		err := kom.DefaultCluster().
			WithContext(context.TODO()).
			CRD("stable.example.com", "v1", "CronTab").
			Name("test-crontab").
			Namespace("default").
			Get(&crontab).Error
		if err != nil {
			t.Logf("CRD Get error: %v\n", err)
		}
	})

	t.Run("List CR", func(t *testing.T) {
		var crontabList []unstructured.Unstructured
		err := kom.DefaultCluster().
			WithContext(context.TODO()).
			CRD("stable.example.com", "v1", "CronTab").
			Namespace("default").
			List(&crontabList).Error
		if err != nil {
			t.Logf("CRD List error: %v\n", err)
		}
		t.Logf("CRD List count %d\n", len(crontabList))
	})

	t.Run("Delete CR", func(t *testing.T) {
		err := kom.DefaultCluster().
			WithContext(context.TODO()).
			CRD("stable.example.com", "v1", "CronTab").
			Name("test-crontab").
			Namespace("default").
			Delete().Error
		if err != nil {
			t.Logf("CRD Delete error: %v\n", err)
		}
	})
}
