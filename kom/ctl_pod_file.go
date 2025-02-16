package kom

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/weibaohui/kom/utils"
	"k8s.io/klog/v2"
)

// FileInfo 文件节点结构
type FileInfo struct {
	Name        string `json:"name"`
	Type        string `json:"type"` // file or directory
	Permissions string `json:"permissions"`
	Owner       string `json:"owner"`
	Group       string `json:"group"`
	Size        int64  `json:"size"`
	ModTime     string `json:"modTime"`
	Path        string `json:"path"`  // 存储路径
	IsDir       bool   `json:"isDir"` // 指示是否
}

// ListFiles  获取容器中指定路径的文件和目录列表
func (p *pod) ListFiles(path string) ([]*FileInfo, error) {
	klog.V(6).Infof("ListFiles %s from [%s/%s:%s]\n", path, p.kubectl.Statement.Namespace, p.kubectl.Statement.Name, p.kubectl.Statement.ContainerName)

	var result []byte
	err := p.Command("ls", "-l", path).Execute(&result).Error
	if err != nil {
		return nil, fmt.Errorf("error executing ListFiles: %v", err)
	}

	return parseFileList(path, string(result)), nil
}
func (p *pod) DownloadFile(filePath string) ([]byte, error) {
	klog.V(6).Infof("DownloadFile %s from [%s/%s:%s]\n", filePath, p.kubectl.Statement.Namespace, p.kubectl.Statement.Name, p.kubectl.Statement.ContainerName)

	result, err := p.DownloadTarFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("error executing DownloadTarFile: %v", err)
	}

	tr := tar.NewReader(bytes.NewReader(result))
	var fileContent []byte
	found := false

	// 遍历 tar 中的每个文件
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error reading tar header: %v", err)
		}

		if header.Name == strings.TrimPrefix(filePath, "/") {
			found = true
			// 使用带有大小限制的读取方式
			if header.Size > 500*1024*1024 { // 500MB 限制
				return nil, fmt.Errorf("file size %d exceeds maximum allowed size", header.Size)
			}

			buf := bytes.NewBuffer(make([]byte, 0, header.Size))
			if _, err := io.Copy(buf, tr); err != nil {
				return nil, fmt.Errorf("error reading file content: %v", err)
			}
			fileContent = buf.Bytes()
			break
		}
	}

	if !found {
		return nil, fmt.Errorf("file %s not found in container", filePath)
	}

	return fileContent, nil
}
func (p *pod) DownloadTarFile(filePath string) ([]byte, error) {
	klog.V(6).Infof("DownloadTarFile %s from [%s/%s:%s]\n", filePath, p.kubectl.Statement.Namespace, p.kubectl.Statement.Name, p.kubectl.Statement.ContainerName)

	var result []byte
	err := p.Command("tar", "cf", "-", filePath).Execute(&result).Error
	if err != nil {
		return nil, fmt.Errorf("error executing DownloadTarFile: %v", err)
	}

	return result, nil
}
func (p *pod) DeleteFile(filePath string) ([]byte, error) {
	klog.V(6).Infof("DeleteFile %s from [%s/%s:%s]\n", filePath, p.kubectl.Statement.Namespace, p.kubectl.Statement.Name, p.kubectl.Statement.ContainerName)
	var result []byte
	err := p.Command("rm", "-rf", filePath).Execute(&result).Error
	if err != nil {
		return nil, fmt.Errorf("error executing DeleteFile : %v", err)
	}

	return result, nil
}

// UploadFile 将文件上传到指定容器
func (p *pod) UploadFile(destPath string, file *os.File) error {
	klog.V(6).Infof("UploadFile %s to [%s/%s:%s] \n", destPath, p.kubectl.Statement.Namespace, p.kubectl.Statement.Name, p.kubectl.Statement.ContainerName)

	// 读取并压缩文件内容
	var buf bytes.Buffer
	if err := createTar(file, &buf); err != nil {
		panic(err.Error())
	}
	var result []byte
	err := p.
		Stdin(&buf).
		Command("tar", "-xmf", "-", "-C", destPath).
		Execute(&result).Error
	if err != nil {
		return fmt.Errorf("error executing UploadFile: %v", err)
	}
	return nil
}

// createTar 创建一个 tar 格式的压缩文件
func createTar(file *os.File, buf *bytes.Buffer) error {
	// 创建 tar writer
	tw := tar.NewWriter(buf)
	defer tw.Close()

	// 获取文件信息
	stat, err := file.Stat()
	if err != nil {
		return err
	}

	// 添加文件头信息
	hdr := &tar.Header{
		Name: stat.Name(),
		Mode: int64(stat.Mode()),
		Size: stat.Size(),
	}
	if err := tw.WriteHeader(hdr); err != nil {
		return err
	}

	// 将文件内容写入到 tar
	_, err = io.Copy(tw, file)
	return err
}

// SaveFile
// TODO 要写入文件的字节数据
//
//	data := []byte("这是一些字节数据。")
//
//	// 创建或打开文件
//	file, err := os.Create("output.txt")
//	if err != nil {
//	    fmt.Println("无法创建文件:", err)
//	    return
//	}
//	defer file.Close() // 确保在函数结束时关闭文件
//
//	// 将 []byte 写入文件
//	_, err = file.Write(data)
//	if err != nil {
//	    fmt.Println("写入文件失败:", err)
//	    return
//	}
//
//	fmt.Println("字节数据已成功写入文件.")
func (p *pod) SaveFile(destPath string, context string) error {
	klog.V(6).Infof("SaveFile %s to [%s/%s:%s]\n", destPath, p.kubectl.Statement.Namespace, p.kubectl.Statement.Name, p.kubectl.Statement.ContainerName)
	klog.V(8).Infof("SaveFile %s \n", context)

	var result []byte
	err := p.
		Stdin(strings.NewReader(context)).
		Command("sh", "-c", fmt.Sprintf("cat > %s", destPath)).
		Execute(&result).Error
	if err != nil {
		return fmt.Errorf("error executing command: %v", err)
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

// parseFileList 解析输出并生成 FileInfo 列表
func parseFileList(path, output string) []*FileInfo {
	var nodes []*FileInfo
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		klog.V(6).Infof("parseFileList path %s %s\n", path, line)
		if len(parts) < 9 {
			continue // 不完整的行
		}

		permissions := parts[0]
		name := parts[8]
		size := parts[4]
		owner := parts[2]
		group := parts[3]
		modTime := strings.Join(parts[5:8], " ")

		// 判断文件类型

		fileType := getFileType(permissions)

		// 封装成 FileInfo
		node := FileInfo{
			Path:        fmt.Sprintf("/%s", name),
			Name:        name,
			Type:        fileType,
			Owner:       owner,
			Group:       group,
			Permissions: permissions,
			Size:        utils.ToInt64(size),
			ModTime:     modTime,
			IsDir:       fileType == "directory",
		}
		if strings.HasPrefix(name, "/") {
			node.Path = name
		} else if path != "/" && path != name {
			node.Path = fmt.Sprintf("%s/%s", path, name)
		} else {
			node.Path = fmt.Sprintf("/%s", name)
		}

		nodes = append(nodes, &node)
	}

	return nodes
}
