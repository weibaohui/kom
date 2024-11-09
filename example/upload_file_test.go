package example

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/weibaohui/kom/kom"
	"github.com/weibaohui/kom/utils"
)

func TestUploadFile(t *testing.T) {

	// 模拟上传一个100M的临时文件
	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "upload-*")
	if err != nil {
		t.Fatalf("error creating temp directory: %v", err)
	}

	// 使用原始文件名生成临时文件路径
	tempFilePath := filepath.Join(tempDir, "test")

	err = dd(tempFilePath, 100*1024*1024)
	if err != nil {
		t.Fatalf("error creating temp test file %s: %v", tempFilePath, err)
	}
	op, err := os.Open(tempFilePath)
	if err != nil {
		t.Fatalf("Failed to open file: %v", err)
	}
	defer op.Close()
	err = kom.DefaultCluster().Namespace("default").
		Name("random").
		ContainerName("random").Poder().
		UploadFile("/etc/", op)

	// 执行ls -lh /etc/ | grep test 查看文件是否已上传
	var execResult []byte
	err = kom.DefaultCluster().Namespace("default").
		Name("random").
		ContainerName("random").
		Command("sh", "-c", "ls -lh /etc/ | grep test").
		Execute(&execResult).Error
	if err != nil {
		t.Fatalf("Error executing command: %v", err)
	}
	t.Logf("ls test file info :\n%s", execResult)
	if !strings.Contains(string(execResult), "100.0M") {
		t.Fatalf("未找到上传文件，测试失败")
	}
}
func TestSaveFile(t *testing.T) {

	context := utils.RandNLengthString(20)
	t.Logf("将%s写入/etc/xyz\n", context)
	err := kom.DefaultCluster().Namespace("default").
		Name("random").
		ContainerName("random").Poder().
		SaveFile("/etc/xyz", context)
	if err != nil {
		t.Errorf("Error executing command: %v", err)
	}

	result, err := kom.DefaultCluster().Namespace("default").
		Name("random").
		ContainerName("random").Poder().DownloadFile("/etc/xyz")
	if err != nil {
		t.Errorf("Error executing command: %v", err)
	}
	t.Logf("从/etc/xyz读取到%s\n", string(result))

	if !strings.Contains(string(result), context) {
		t.Fatalf("读取文件失败，应为%s,实际%s", context, string(result))
	}
}
func TestListFile(t *testing.T) {

	result, err := kom.DefaultCluster().Namespace("default").
		Name("random").
		ContainerName("random").Poder().ListFiles("/etc")
	if err != nil {
		t.Errorf("Error executing command: %v", err)
	}
	t.Logf("读取文件数量%d", len(result))
	if len(result) == 0 {
		t.Fatalf("读取文件失败，不应为空,实际%d", len(result))
	}
}

func TestDeleteFile(t *testing.T) {

	// 先创建一个文件，读取验证存在，然后删除，然后再读取，验证不存在

	// 创建文件
	context := utils.RandNLengthString(20)
	t.Logf("将%s写入/etc/xyz\n", context)
	err := kom.DefaultCluster().Namespace("default").
		Name("random").
		ContainerName("random").Poder().
		SaveFile("/etc/xyz", context)
	if err != nil {
		t.Errorf("Error executing SaveFile command: %v", err)
	}

	// 读取
	result, err := kom.DefaultCluster().Namespace("default").
		Name("random").
		ContainerName("random").Poder().DownloadFile("/etc/xyz")
	if err != nil {
		t.Errorf("Error executing DownloadFile command: %v", err)
	}
	t.Logf("从/etc/xyz读取到%s\n", string(result))

	// 验证存在
	if !strings.Contains(string(result), context) {
		t.Fatalf("读取文件失败，应为%s,实际%s", context, string(result))
	}

	// 删除文件
	_, err = kom.DefaultCluster().Namespace("default").
		Name("random").
		ContainerName("random").Poder().DeleteFile("/etc/xyz")
	if err != nil {
		t.Errorf("Error executing DeleteFile command: %v", err)
	}

	// 尝试该读取文件
	result, err = kom.DefaultCluster().Namespace("default").
		Name("random").
		ContainerName("random").Poder().DownloadFile("/etc/xyz")
	if err != nil {
		// 文件已经不存在，看看报错中是否包含文件不存在，包含成功
		if strings.Contains(err.Error(), "No such file or directory") {
			t.Logf("Error executing DownloadFile command: %v", err)
			t.Logf("文件不存在，已成功删除")
		} else {
			t.Fatalf("删除文件失败，%v", err)
		}
	}

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
