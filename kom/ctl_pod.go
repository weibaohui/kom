package kom

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/weibaohui/kom/utils"
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/klog/v2"
	resourcehelper "k8s.io/kubectl/pkg/util/resource"
)

type pod struct {
	kubectl *Kubectl
	Error   error
}

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

func (p *pod) ContainerName(c string) *pod {
	tx := p.kubectl.getInstance()
	tx.Statement.ContainerName = c
	p.kubectl = tx
	return p
}

func (p *pod) Command(command string, args ...string) *pod {
	tx := p.kubectl.getInstance()
	tx.Statement.Command = command
	tx.Statement.Args = args
	p.kubectl = tx
	return p
}
func (p *pod) Execute(dest interface{}) *pod {
	tx := p.kubectl.getInstance()
	tx.Statement.Dest = dest
	tx.Error = tx.Callback().Exec().Execute(tx)
	p.Error = tx.Error
	return p
}

func (p *pod) StreamExecute(stdout, stderr func(data []byte) error) *Kubectl {
	tx := p.kubectl.getInstance()
	tx.Statement.StdoutCallback = stdout
	tx.Statement.StderrCallback = stderr
	tx.Error = tx.Callback().StreamExec().Execute(tx)
	p.Error = tx.Error
	return tx
}
func (p *pod) Stdin(reader io.Reader) *pod {
	tx := p.kubectl.getInstance()
	tx.Statement.Stdin = reader
	return p
}
func (p *pod) GetLogs(requestPtr interface{}, opt *v1.PodLogOptions) *pod {
	tx := p.kubectl.getInstance()
	if tx.Statement.ContainerName == "" {
		p.Error = fmt.Errorf("请先设置ContainerName")
		return p
	}
	tx.Statement.PodLogOptions = opt
	tx.Statement.PodLogOptions.Container = tx.Statement.ContainerName
	tx.Statement.Dest = requestPtr
	tx.Error = tx.Callback().Logs().Execute(tx)
	p.Error = tx.Error
	return p
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

// DownloadFile 从指定容器下载文件
func (p *pod) DownloadFile(filePath string) ([]byte, error) {
	klog.V(6).Infof("DownloadFile %s from [%s/%s:%s]\n", filePath, p.kubectl.Statement.Namespace, p.kubectl.Statement.Name, p.kubectl.Statement.ContainerName)
	var result []byte
	err := p.Command("cat", filePath).Execute(&result).Error
	if err != nil {
		return nil, fmt.Errorf("error executing DownloadFile: %v", err)
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

// ResourceUsage 获取节点的资源使用情况，包括资源的请求和限制，还有当前使用占比
func (p *pod) ResourceUsage() *ResourceUsageResult {

	var inst *v1.Pod
	cacheTime := p.kubectl.Statement.CacheTTL
	if cacheTime == 0 {
		cacheTime = 5 * time.Second
	}
	klog.V(6).Infof("Pod ResourceUsage cacheTime %v\n", cacheTime)
	err := p.kubectl.newInstance().Resource(&v1.Pod{}).
		Namespace(p.kubectl.Statement.Namespace).
		Name(p.kubectl.Statement.Name).
		WithCache(cacheTime).
		Get(&inst).Error
	if err != nil {
		klog.V(6).Infof("Get ResourceUsage in pod/%s  error %v\n", p.kubectl.Statement.Name, err.Error())
		return nil
	}

	nodeName := inst.Spec.NodeName
	if nodeName == "" {
		klog.V(6).Infof("Get Pod ResourceUsage in pod/%s  error %v\n", p.kubectl.Statement.Name, "nodeName is empty")
		return nil
	}
	var n *v1.Node
	err = p.kubectl.newInstance().Resource(&v1.Node{}).
		WithCache(cacheTime).
		Name(nodeName).Get(&n).Error
	if err != nil {
		klog.V(6).Infof("Get Pod ResourceUsage in node/%s  error %v\n", nodeName, err.Error())
		return nil
	}

	req, limit := resourcehelper.PodRequestsAndLimits(inst)
	if req == nil || limit == nil {
		return nil
	}
	allocatable := n.Status.Capacity
	if len(n.Status.Allocatable) > 0 {
		allocatable = n.Status.Allocatable
	}

	klog.V(8).Infof("allocatable=:\n%s", utils.ToJSON(allocatable))
	cpuReq, cpuLimit, memoryReq, memoryLimit := req[v1.ResourceCPU], limit[v1.ResourceCPU], req[v1.ResourceMemory], limit[v1.ResourceMemory]
	fractionCpuReq := float64(cpuReq.MilliValue()) / float64(allocatable.Cpu().MilliValue()) * 100
	fractionCpuLimit := float64(cpuLimit.MilliValue()) / float64(allocatable.Cpu().MilliValue()) * 100
	fractionMemoryReq := float64(memoryReq.Value()) / float64(allocatable.Memory().Value()) * 100
	fractionMemoryLimit := float64(memoryLimit.Value()) / float64(allocatable.Memory().Value()) * 100

	usageFractions := map[v1.ResourceName]ResourceUsageFraction{
		v1.ResourceCPU: {
			RequestFraction: fractionCpuReq,
			LimitFraction:   fractionCpuLimit,
		},
		v1.ResourceMemory: {
			RequestFraction: fractionMemoryReq,
			LimitFraction:   fractionMemoryLimit,
		},
	}

	klog.V(6).Infof("%s\t%s\t\t%s (%d%%)\t%s (%d%%)\t%s (%d%%)\t%s (%d%%)\n", inst.Namespace, inst.Name,
		cpuReq.String(), int64(fractionCpuReq), cpuLimit.String(), int64(fractionCpuLimit),
		memoryReq.String(), int64(fractionMemoryReq), memoryLimit.String(), int64(fractionMemoryLimit))
	return &ResourceUsageResult{
		Requests:       req,
		Limits:         limit,
		Allocatable:    allocatable,
		UsageFractions: usageFractions,
	}
}
func (p *pod) ResourceUsageTable() []*ResourceUsageRow {
	usage := p.ResourceUsage()
	data, err := convertToTableData(usage)
	if err != nil {
		klog.V(6).Infof("convertToTableData error %v\n", err.Error())
		return make([]*ResourceUsageRow, 0)
	}
	return data
}

// 获取pod相关的Service
func (p *pod) LinkedService() ([]*v1.Service, error) {
	// 	查询流程
	// 获取目标 Pod 的详细信息：

	// 使用 Pod 的 API 获取其 metadata.labels。
	// 确定 Pod 所在的 Namespace。
	// 获取 Namespace 内的所有 Services：

	// 使用 kubectl get services -n {namespace} 或调用 API /api/v1/namespaces/{namespace}/services。
	// 逐个匹配 Service 的 selector：

	// 对每个 Service：
	// 提取其 spec.selector。
	// 遍历 selector 的所有键值对，检查 Pod 是否包含这些标签且值相等。
	// 如果所有标签条件都满足，将此 Service 记录为与该 Pod 关联。
	// 返回结果：

	// 将所有匹配的 Service 名称及相关信息返回。
	var pod *v1.Pod
	err := p.kubectl.Get(&pod).Error
	if err != nil {
		return nil, fmt.Errorf("get pod %s/%s error %v", p.kubectl.Statement.Namespace, p.kubectl.Statement.Name, err.Error())
	}
	if pod == nil {
		return nil, fmt.Errorf("get pod %s/%s error %v", p.kubectl.Statement.Namespace, p.kubectl.Statement.Name, "pod is nil")
	}
	podLabels := pod.GetLabels()

	var services []*v1.Service
	err = p.kubectl.newInstance().WithContext(p.kubectl.Statement.Context).
		Resource(&v1.Service{}).
		Namespace(p.kubectl.Statement.Namespace).
		List(&services).Error

	if err != nil {
		return nil, fmt.Errorf("get service error %v", err.Error())
	}

	var result []*v1.Service
	for _, svc := range services {
		serviceLabels := svc.Spec.Selector
		// 遍历selector
		// serviceLabels中所有的kv,都必须在podLabels中存在,且值相等
		// 如果有一个不满足,则跳过
		for k, v := range serviceLabels {
			if podLabels[k] != v {
				continue
			}
			result = append(result, svc)
		}

	}
	return result, nil
}

func (p *pod) LinkedEndpoints() ([]*v1.Endpoints, error) {

	services, err := p.LinkedService()
	if err != nil {
		return nil, err
	}
	if len(services) == 0 {
		return nil, nil
	}
	// endpoints 与 svc 同名
	// 1.获取service 名称
	// 2.获取endpoints
	// 3.返回endpoints

	var names []string
	for _, svc := range services {
		names = append(names, svc.Name)
	}

	var endpoints []*v1.Endpoints

	err = p.kubectl.newInstance().
		WithContext(p.kubectl.Statement.Context).
		Resource(&v1.Endpoints{}).
		Namespace(p.kubectl.Statement.Namespace).
		Where("metadata.name in " + utils.StringListToSQLIn(names)).
		List(&endpoints).Error
	if err != nil {
		return nil, err
	}
	return endpoints, nil
}

func (p *pod) LinkedPVC() ([]*v1.PersistentVolumeClaim, error) {

	var pod v1.Pod
	err := p.kubectl.Get(&pod).Error
	if err != nil {
		return nil, err
	}
	// 找打pvc 名称列表
	var pvcNames []string
	for _, volume := range pod.Spec.Volumes {
		if volume.PersistentVolumeClaim != nil {
			pvcNames = append(pvcNames, volume.PersistentVolumeClaim.ClaimName)
		}
	}

	// 找出同ns下pvc的列表，过滤pvcNames
	var pvcList []*v1.PersistentVolumeClaim
	err = p.kubectl.newInstance().WithContext(p.kubectl.Statement.Context).
		Resource(&v1.PersistentVolumeClaim{}).
		Namespace(p.kubectl.Statement.Namespace).
		List(&pvcList).Error
	if err != nil {
		return nil, err
	}

	// 过滤pvcList，只保留pvcNames
	var result []*v1.PersistentVolumeClaim
	for _, pvc := range pvcList {
		if slices.Contains(pvcNames, pvc.Name) {
			result = append(result, pvc)
		}
	}
	return result, nil
}

func (p *pod) LinkedIngress() ([]*networkingv1.Ingress, error) {

	var pod v1.Pod
	err := p.kubectl.Get(&pod).Error
	if err != nil {
		return nil, err
	}
	services, err := p.LinkedService()
	if err != nil {
		return nil, err
	}

	var servicesName []string
	for _, svc := range services {
		servicesName = append(servicesName, svc.Name)
	}

	// 获取ingress
	// Ingress 通过 spec.rules 或 spec.defaultBackend 中的 service.name 指定关联的 Service。
	// 遍历services，获取ingress
	var ingressList []networkingv1.Ingress
	err = p.kubectl.newInstance().WithContext(p.kubectl.Statement.Context).
		Resource(&networkingv1.Ingress{}).
		Namespace(p.kubectl.Statement.Namespace).
		WithCache(p.kubectl.Statement.CacheTTL).
		List(&ingressList).Error
	if err != nil {
		return nil, err
	}

	// 过滤ingressList，只保留与services关联的ingress
	var result []*networkingv1.Ingress
	for _, ingress := range ingressList {
		if slices.Contains(servicesName, ingress.Spec.Rules[0].Host) {
			result = append(result, &ingress)
		}
	}
	// 遍历 Ingress 检查关联
	for _, ingress := range ingressList {
		if ingress.Spec.DefaultBackend != nil {
			if ingress.Spec.DefaultBackend.Service != nil && ingress.Spec.DefaultBackend.Service.Name != "" {
				if slices.Contains(servicesName, ingress.Spec.DefaultBackend.Service.Name) {
					result = append(result, &ingress)
				}
			}
		}

		for _, rule := range ingress.Spec.Rules {
			if rule.HTTP != nil {
				for _, path := range rule.HTTP.Paths {
					if path.Backend.Service != nil && path.Backend.Service.Name != "" {

						backName := path.Backend.Service.Name
						if slices.Contains(servicesName, backName) {
							result = append(result, &ingress)
						}
					}
				}

			}

		}
	}

	return result, nil
}
