package poder

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"strings"

	"github.com/weibaohui/kom/kom"
	"github.com/weibaohui/kom/utils"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/klog/v2"
)

type Poder struct {
	Kom           *kom.Kom
	containerName string
}

func Instance() *Poder {
	return &Poder{
		Kom: kom.Init(),
	}
}
func (p *Poder) WithContext(ctx context.Context) *Poder {
	p.Kom = p.Kom.WithContext(ctx)
	return p
}
func (p *Poder) Namespace(ns string) *Poder {
	p.Kom = p.Kom.Namespace(ns)
	return p
}
func (p *Poder) Name(name string) *Poder {
	p.Kom = p.Kom.Name(name)
	return p
}
func (p *Poder) ContainerName(name string) *Poder {
	p.containerName = name
	return p
}
func (p *Poder) GetLogs(name string, opts *v1.PodLogOptions) *rest.Request {
	return p.Kom.Client.CoreV1().Pods(p.Kom.Statement.Namespace).GetLogs(name, opts)
}

// PodFileNode 文件节点结构
type PodFileNode struct {
	Name        string `json:"name"`
	Type        string `json:"type"` // file or directory
	Permissions string `json:"permissions"`
	Size        int64  `json:"size"`
	ModTime     string `json:"modTime"`
	Path        string `json:"path"`  // 存储路径
	IsDir       bool   `json:"isDir"` // 指示是否
}

// GetFileList  获取容器中指定路径的文件和目录列表
func (p *Poder) GetFileList(path string) ([]*PodFileNode, error) {
	cmd := []string{"ls", "-l", path}
	req := p.Kom.Client.CoreV1().RESTClient().
		Get().
		Namespace(p.Kom.Statement.Namespace).
		Resource("pods").
		Name(p.Kom.Statement.Name).
		SubResource("exec").
		Param("container", p.containerName).
		Param("command", cmd[0]).
		Param("command", cmd[1]).
		Param("command", cmd[2]).
		Param("tty", "false").
		Param("stdin", "false").
		Param("stdout", "true").
		Param("stderr", "true")

	executor, err := remotecommand.NewSPDYExecutor(p.Kom.RestConfig(), "POST", req.URL())
	if err != nil {
		return nil, fmt.Errorf("error creating executor: %v", err)
	}

	var stdout bytes.Buffer
	err = executor.Stream(remotecommand.StreamOptions{
		Stdout: &stdout,
		Stderr: os.Stderr,
	})

	if err != nil {
		return nil, fmt.Errorf("error executing command: %v", err)
	}

	return parseFileList(path, stdout.String()), nil
}

// DownloadFile 从指定容器下载文件
func (p *Poder) DownloadFile(filePath string) ([]byte, error) {
	cmd := []string{"cat", filePath}
	klog.V(8).Infof("DownloadFile %s", filePath)

	req := p.Kom.Client.CoreV1().RESTClient().
		Get().
		Namespace(p.Kom.Statement.Namespace).
		Resource("pods").
		Name(p.Kom.Statement.Name).
		SubResource("exec").
		Param("container", p.containerName).
		Param("command", cmd[0]).
		Param("command", cmd[1]).
		Param("tty", "false").
		Param("stdin", "false").
		Param("stdout", "true").
		Param("stderr", "true")

	executor, err := remotecommand.NewSPDYExecutor(p.Kom.RestConfig(), "POST", req.URL())
	if err != nil {
		return nil, fmt.Errorf("error creating executor: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	err = executor.Stream(remotecommand.StreamOptions{
		Stdout: &stdout,
		Stderr: &stderr,
	})

	if err != nil {
		s := stderr.String()
		if strings.Contains(s, "Invalid argument") {
			return nil, fmt.Errorf("系统参数错误 %v", s)
		}
		return nil, fmt.Errorf("error executing command: %v %v", err, s)
	}

	return stdout.Bytes(), nil
}

// UploadFile 将文件上传到指定容器
func (p *Poder) UploadFile(destPath string, file multipart.File) error {
	// 创建临时文件
	tempFile, err := os.CreateTemp("", "upload-*")
	if err != nil {
		return fmt.Errorf("error creating temp file: %v", err)
	}
	defer func(name string) {
		err := os.Remove(name)
		if err != nil {
			klog.V(6).Infof("remve %s error:%v", name, err)
		}
	}(tempFile.Name()) // 确保临时文件在函数结束时被删除

	// 将上传的文件内容写入临时文件
	_, err = io.Copy(tempFile, file)
	if err != nil {
		return fmt.Errorf("error writing to temp file: %v", err)
	}

	// 确保文件关闭
	if err := tempFile.Close(); err != nil {
		return fmt.Errorf("error closing temp file: %v", err)
	}

	cmd := []string{"sh", "-c", fmt.Sprintf("cat > %s", destPath)}

	req := p.Kom.Client.CoreV1().RESTClient().
		Get().
		Namespace(p.Kom.Statement.Namespace).
		Resource("pods").
		Name(p.Kom.Statement.Name).
		SubResource("exec").
		Param("container", p.containerName).
		Param("tty", "false").
		Param("command", cmd[0]).
		Param("command", cmd[1]).
		Param("command", cmd[2]).
		Param("stdin", "true").
		Param("stdout", "true").
		Param("stderr", "true")

	executor, err := remotecommand.NewSPDYExecutor(p.Kom.RestConfig(), "POST", req.URL())
	if err != nil {
		return fmt.Errorf("error creating executor: %v", err)
	}

	// 打开本地文件进行传输
	readFile, err := os.Open(tempFile.Name())
	if err != nil {
		return fmt.Errorf("error opening file: %v", err)
	}
	defer func(readFile *os.File) {
		err := readFile.Close()
		if err != nil {
			klog.V(6).Infof("readFile.Close() error:%v", err)
		}
	}(readFile)
	var stdout, stderr bytes.Buffer
	err = executor.Stream(remotecommand.StreamOptions{
		Stdin:  readFile,
		Stdout: &stdout,
		Stderr: &stderr,
	})

	if err != nil {
		return fmt.Errorf("error executing command: %v: %s", err, stderr.String())
	}

	return nil
}

func (p *Poder) SaveFile(path string, context string) error {

	// 创建临时文件
	tempFile, err := os.CreateTemp("", "upload-*")
	if err != nil {
		return fmt.Errorf("error creating temp file: %v", err)
	}
	defer func(name string) {
		err := os.Remove(name)
		if err != nil {
			klog.V(6).Infof("remve %s error:%v", name, err)
		}
	}(tempFile.Name()) // 确保临时文件在函数结束时被删除

	// 将上传的文件内容写入临时文件
	_, err = io.WriteString(tempFile, context)
	if err != nil {
		return fmt.Errorf("error writing to temp file: %v", err)
	}

	// 确保文件关闭
	if err := tempFile.Close(); err != nil {
		return fmt.Errorf("error closing temp file: %v", err)
	}

	cmd := []string{"sh", "-c", fmt.Sprintf("cat > %s", path)}

	req := p.Kom.Client.CoreV1().RESTClient().
		Get().
		Namespace(p.Kom.Statement.Namespace).
		Resource("pods").
		Name(p.Kom.Statement.Name).
		SubResource("exec").
		Param("container", p.containerName).
		Param("tty", "false").
		Param("command", cmd[0]).
		Param("command", cmd[1]).
		Param("command", cmd[2]).
		Param("stdin", "true").
		Param("stdout", "true").
		Param("stderr", "true")

	executor, err := remotecommand.NewSPDYExecutor(p.Kom.RestConfig(), "POST", req.URL())
	if err != nil {
		return fmt.Errorf("error creating executor: %v", err)
	}

	// 打开本地文件进行传输
	file, err := os.Open(tempFile.Name())
	if err != nil {
		return fmt.Errorf("error opening file: %v", err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			klog.V(6).Infof("file.Close() error:%v", err)
		}
	}(file)
	var stdout, stderr bytes.Buffer
	err = executor.Stream(remotecommand.StreamOptions{
		Stdin:  file,
		Stdout: &stdout,
		Stderr: &stderr,
	})

	if err != nil {
		return fmt.Errorf("error executing command: %v: %s", err, stderr.String())
	}

	return nil
}

// getFileType 根据文件权限获取文件类型
//
// l 代表符号链接（Symbolic link）
// - 代表普通文件（Regular file）
// d 代表目录（Directory）
// b 代表块设备（Block device）
// c 代表字符设备（Character device）
// p 代表命名管道（Named pipe）
// s 代表套接字（Socket）
func getFileType(permissions string) string {
	// 获取文件类型标志位
	p := permissions[0]
	var fileType string

	switch p {
	case 'd':
		fileType = "directory" // 目录
	case '-':
		fileType = "file" // 普通文件
	case 'l':
		fileType = "link" // 符号链接
	case 'b':
		fileType = "block" // 块设备
	case 'c':
		fileType = "character" // 字符设备
	case 'p':
		fileType = "pipe" // 命名管道
	case 's':
		fileType = "socket" // 套接字
	default:
		fileType = "unknown" // 未知类型
	}

	return fileType
}

// parseFileList 解析输出并生成 PodFileNode 列表
func parseFileList(path, output string) []*PodFileNode {
	var nodes []*PodFileNode
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) < 9 {
			continue // 不完整的行
		}

		permissions := parts[0]
		name := parts[8]
		size := parts[4]
		modTime := strings.Join(parts[5:8], " ")

		// 判断文件类型

		fileType := getFileType(permissions)

		// 封装成 PodFileNode
		node := PodFileNode{
			Path:        fmt.Sprintf("/%s", name),
			Name:        name,
			Type:        fileType,
			Permissions: permissions,
			Size:        utils.ToInt64(size),
			ModTime:     modTime,
			IsDir:       fileType == "directory",
		}
		if path != "/" {
			node.Path = fmt.Sprintf("%s/%s", path, name)
		}
		nodes = append(nodes, &node)
	}

	return nodes
}
