package example

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/weibaohui/kom/kom"
	"github.com/weibaohui/kom/utils"
	"k8s.io/klog/v2"
)

// TestMain 是测试的入口函数

func TestUploadFile(t *testing.T) {
	yaml := `apiVersion: v1
kind: Pod
metadata:
  name: random-char-pod
  namespace: default
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
    name: container
`
	result := kom.DefaultCluster().Applier().Apply(yaml)
	for _, str := range result {
		fmt.Println(str)
	}
	time.Sleep(time.Second * 1)

	context := utils.RandNLengthString(20)
	fmt.Printf("将%s写入/etc/xyz\n", context)
	err := kom.DefaultCluster().Namespace("default").
		Name("random-char-pod").
		ContainerName("container").Poder().
		SaveFile("/etc/xyz", context)
	if err != nil {
		klog.Errorf("Error executing command: %v", err)
	}

	file, err := kom.DefaultCluster().Namespace("default").
		Name("random-char-pod").
		ContainerName("container").Poder().DownloadFile("/etc/xyz")
	if err != nil {
		klog.Errorf("Error executing command: %v", err)
	}
	fmt.Printf("从/etc/xyz读取到%s\n", string(file))

	// 模拟上传一个100M的临时文件
	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "upload-*")
	if err != nil {
		fmt.Printf("error creating temp directory: %v", err)
	}

	// 使用原始文件名生成临时文件路径
	tempFilePath := filepath.Join(tempDir, "test")

	err = dd(tempFilePath, 100*1024*1024)
	if err != nil {
		fmt.Printf("error creating temp test file %s: %v", tempFilePath, err)
	}
	op, err := os.Open(tempFilePath)
	if err != nil {
		log.Fatalf("Failed to open file: %v", err)
	}
	defer op.Close()
	err = kom.DefaultCluster().Namespace("default").
		Name("random-char-pod").
		ContainerName("container").Poder().
		UploadFile("/etc/", op)

	// 执行ls -lh /etc/ | grep test 查看文件是否已上传
	var execResult []byte
	err = kom.DefaultCluster().Namespace("default").
		Name("random-char-pod").
		ContainerName("container").
		Command("ls", "-lh", "/etc/", "|", "grep", "test").
		Execute(&execResult).Error
	if err != nil {
		klog.Errorf("Error executing command: %v", err)
	}
	fmt.Printf("ls test file info :\n%s", execResult)

}

// 创建一个指定大小的文件。
// 10 * 1024 * 1024 =10MB
func dd(filePath string, size int64) error {
	// 指定文件路径和大小（以字节为单位）
	fileSize := size // 10MB

	// 创建文件
	file, err := os.Create(filePath)
	if err != nil {
		fmt.Println("Error creating file:", err)
		return err
	}
	defer file.Close()

	// 定义写入的数据块，通常是一个固定的字节，重复写入直到达到指定大小
	buffer := make([]byte, 1024) // 1KB的缓冲区，可以选择其他大小

	// 写入数据直到文件达到指定大小
	var written int64
	for written < fileSize {
		n, err := file.Write(buffer)
		if err != nil {
			fmt.Println("Error writing to file:", err)
			return err
		}
		written += int64(n)
	}

	fmt.Printf("Successfully created a file of size %d bytes: %s\n", fileSize, filePath)
	return nil
}
