# Kom - Kubernetes Operations Manager

[English](README_en.md) | [中文](README.md)
[![kom](https://img.shields.io/badge/License-MIT-blue?style=flat-square)](https://github.com/weibaohui/kom/blob/master/LICENSE)


## 简介

`kom` 是一个用于 Kubernetes 操作的工具，相当于SDK级的kubectl、client-go的使用封装。
它提供了一系列功能来管理 Kubernetes 资源，包括创建、更新、删除和获取资源。这个项目支持多种 Kubernetes 资源类型的操作，并能够处理自定义资源定义（CRD）。
通过使用 `kom`，你可以轻松地进行资源的增删改查和日志获取以及操作POD内文件等动作，甚至可以使用SQL语句来查询、管理k8s资源。

## **特点**
1. 简单易用：kom 提供了丰富的功能，包括创建、更新、删除、获取、列表等，包括对内置资源以及CRD资源的操作。
2. 多集群支持：通过RegisterCluster，你可以轻松地管理多个 Kubernetes 集群。
3. MCP支持：支持多集群的MCP管理,同时支持stdio、sse两种模式，内置48种工具，支持SSE模式，支持私有化部署，多人共享。支持超过百种组合操作。
4. 支持跨命名空间：通过kom.Namespace("default","kube-system").List(&items) 跨命名空间查询资源。
5. 链式调用：kom 提供了链式调用，使得操作资源更加简单和直观。
6. 支持自定义资源定义（CRD）：kom 支持自定义资源定义（CRD），你可以轻松地定义和操作自定义资源。
7. 支持回调机制，轻松拓展业务逻辑，而不必跟k8s操作强耦合。
8. 支持POD内文件操作，轻松上传、下载、删除文件。
9. 支持高频操作封装，如deployment的restart重启、scale扩缩容、启停等20余项操作功能。
10. 支持SQL查询k8s资源。select * from pod where metadata.namespace='kube-system' or metadata.namespace='default' order by  metadata.creationTimestamp desc 
11. 支持查询缓存，在高频、批量查询场景下，可设置缓存过期时间，提升查询性能。列表过滤条件不受缓存影响。



## 示例程序
**k8m** 是一个轻量级的 Kubernetes 管理工具，它基于kom、amis实现，单文件，支持多平台架构。
1. **下载**：从 [https://github.com/weibaohui/k8m](https://github.com/weibaohui/k8m) 下载最新版本。
2. **运行**：使用 `./k8m` 命令启动,访问[http://127.0.0.1:3618](http://127.0.0.1:3618)。




## 安装

```bash
import (
    "github.com/weibaohui/kom"
    "github.com/weibaohui/kom/callbacks"
)
func main() {
    // 注册回调，务必先注册
    callbacks.RegisterInit()
    // 注册集群
	defaultKubeConfig := os.Getenv("KUBECONFIG")
	if defaultKubeConfig == "" {
		defaultKubeConfig = filepath.Join(homedir.HomeDir(), ".kube", "config")
	}
	_, _ = kom.Clusters().RegisterInCluster()
	_, _ = kom.Clusters().RegisterByPathWithID(defaultKubeConfig, "default")
	kom.Clusters().Show()
	// 其他逻辑
}
```

## 使用示例

### 0. 多集群 k8s MCP 支持
同时支持stdio、sse两种模式
支持多个tools 支持。包括对任意资源的查询列表删除描述操作，以及POD日志读取操作。
#### 1.集成到代码中
```go
// 一行代码启动MCP Server
mcp.RunMCPServer("kom mcp server", "0.0.1", 9096)



```
#### 2. 编译
```shell
# 源码启动
go build main.go 
//编译为kom
```
#### 3. 启动
启动后支持两种模式，一种为stdio，一种sse。
管理k8s默认使用KUBECONFIG env环境变量。
```shell
# 设置KUBECONFIG环境变量
export KUBECONFIG = /Users/xxx/.kube/config
```
```shell
# 运行
./kom 
# MCP Server 访问地址
http://IP:9096/sse
```
此时，编译得到的二进制文件，可当做stdio 模式使用。
http://IP:9096/sse 模式，可以当做sse 模式使用。


#### 4. 集成到MCP工具中
支持stdio\sse 两种方式集成。
适合MCP工具集成，如Cursor、Claude Desktop(仅支持stdio模式)、Windsurf等，此外也可以使用这些软件的UI操作界面进行添加。
```json
{
  "mcpServers": {
    "kom": {
      "type": "sse",
      "url": "http://IP:9096/sse"
    }
  }
}
```
```json
{
    "mcpServers": {
        "k8m": {
            "command": "path/to/kom",
            "args": []
        }
    }
}
```

####  MCP工具列表（49种）

| 类别                 | 方法                             | 描述                                      |
|--------------------|--------------------------------|-----------------------------------------|
| **集群管理（1）**        | `list_clusters`                | 列出所有已注册的Kubernetes集群                    |
| **部署管理（12）**       | `scale_deployment`             | 扩缩容Deployment                           |
|                    | `restart_deployment`           | 重启Deployment                            |
|                    | `stop_deployment`              | 停止Deployment                            |
|                    | `restore_deployment`           | 恢复Deployment                            |
|                    | `update_tag_deployment`        | 更新Deployment镜像标签                        |
|                    | `rollout_history_deployment`   | 查询Deployment升级历史                        |
|                    | `rollout_undo_deployment`      | 回滚Deployment                            |
|                    | `rollout_pause_deployment`     | 暂停Deployment升级                          |
|                    | `rollout_resume_deployment`    | 恢复Deployment升级                          |
|                    | `rollout_status_deployment`    | 查询Deployment升级状态                        |
|                    | `hpa_list_deployment`          | 查询Deployment的HPA列表                      |
|                    | `list_deployment_pods`         | 获取Deployment管理的Pod列表                    |
| **动态资源管理(含CRD，8)** | `get_k8s_resource`             | 获取k8s资源                                 |
|                    | `describe_k8s_resource`        | 描述k8s资源                                 |
|                    | `delete_k8s_resource`          | 删除k8s资源                                 |
|                    | `list_k8s_resource`            | 列表形式获取k8s资源                             |
|                    | `list_k8s_event`               | 列表形式获取k8s事件                             |
|                    | `patch_k8s_resource`           | 更新k8s资源，以JSON Patch方式更新                 |                               |
|                    | `label_k8s_resource`           | 为k8s资源添加或删除标签                           |
|                    | `annotate_k8s_resource`        | 为k8s资源添加或删除注解                           |
| **节点管理（8）**        | `taint_node`                   | 为节点添加污点                                 |
|                    | `untaint_node`                 | 为节点移除污点                                 |
|                    | `cordon_node`                  | 为节点设置Cordon                             |
|                    | `uncordon_node`                | 为节点取消Cordon                             |
|                    | `drain_node`                   | 为节点执行Drain                              |
|                    | `get_node_resource_usage`      | 查询节点的资源使用情况                             |
|                    | `get_node_ip_usage`            | 查询节点上Pod IP资源使用情况                       |
|                    | `get_node_pod_count`           | 查询节点上的Pod数量                             |
| **Pod 管理（14）**     | `list_pod_files`               | 列出Pod文件                                 |
|                    | `list_all_pod_files`           | 列出Pod所有文件                               |
|                    | `delete_pod_file`              | 删除Pod文件                                 |
|                    | `upload_file_to_pod`           | 上传文件到Pod内，支持传递文本内容，存储为Pod内文件            |
|                    | `get_pod_logs`                 | 获取Pod日志                                 |
|                    | `run_command_in_pod`           | 在Pod中执行命令                               |
|                    | `get_pod_linked_service`       | 获取Pod关联的Service                         |
|                    | `get_pod_linked_ingress`       | 获取Pod关联的Ingress                         |
|                    | `get_pod_linked_endpoints`     | 获取Pod关联的Endpoints                       |
|                    | `get_pod_linked_pvc`           | 获取Pod关联的PVC                             |
|                    | `get_pod_linked_pv`            | 获取Pod关联的PV                              |
|                    | `get_pod_linked_env`           | 通过在pod内运行env命令获取Pod运行时环境变量              |
|                    | `get_pod_linked_env_from_yaml` | 通过Pod yaml定义获取Pod运行时环境变量                |
|                    | `get_pod_resource_usage`       | 获取Pod的资源使用情况，包括CPU和内存的请求值、限制值、可分配值和使用比例 |
| **YAML管理（2）**      | `apply_yaml`                   | 应用YAML资源                                |
|                    | `delete_yaml`                  | 删除YAML资源                                |
| **存储管理（3）**        | `set_default_storageclass`     | 设置默认StorageClass                        |
|                    | `get_storageclass_pvc_count`   | 获取StorageClass下的PVC数量                   |
|                    | `get_storageclass_pv_count`    | 获取StorageClass下的PV数量                    |
| **Ingress管理（1）**   | `set_default_ingressclass`     | 设置默认IngressClass                        |


#### 启动命令
```go
mcp.RunMCPServer("kom mcp server", "0.0.1", 3619)
```
 
#### AI工具集成

##### Claude Desktop
1. 打开Claude Desktop设置面板
2. 在API配置区域添加MCP Server地址
3. 启用SSE事件监听功能
4. 验证连接状态
```json
{
  "mcpServers": {
    "k8m": {
      "command": "path/to/kom",
      "args": []
    }
  }
}
```

##### Cursor
1. 进入Cursor设置界面
2. 找到扩展服务配置选项
3. 支持sse、stdio两种方式。sse 方式填写http://localhost:9096/sse,stdio方式填写kom的文件位置。

##### Windsurf
1. 访问配置中心
2. 设置API服务器地址
3. 支持sse、stdio两种方式。sse 方式填写http://localhost:9096/sse,stdio方式填写kom的文件位置。

#### cherry studio
1. 点击左下角设置
2. 点击MCP 服务器
3. 点击添加服务器
4. 支持sse、stdio两种方式。sse 方式填写http://localhost:9096/sse,stdio方式填写kom的文件位置。


### 1. 多集群管理
#### 注册多集群
```go
// 注册InCluster集群，名称为InCluster
kom.Clusters().RegisterInCluster()
// 注册两个带名称的集群,分别名为orb和docker-desktop
kom.Clusters().RegisterByPathWithID("/Users/kom/.kube/orb", "orb")
kom.Clusters().RegisterByPathWithID("/Users/kom/.kube/config", "docker-desktop")
// 注册一个名为default的集群，那么kom.DefaultCluster()则会返回该集群。
kom.Clusters().RegisterByPathWithID("/Users/kom/.kube/config", "default")
```
#### 显示已注册集群
```go
kom.Clusters().Show()
```
#### 选择默认集群
```go
// 使用默认集群,查询集群内kube-system命名空间下的pod
// 首先尝试返回 ID 为 "InCluster" 的实例，如果不存在，
// 则尝试返回 ID 为 "default" 的实例。
// 如果上述两个名称的实例都不存在，则返回 clusters 列表中的任意一个实例。
var pods []corev1.Pod
err = kom.DefaultCluster().Resource(&corev1.Pod{}).Namespace("kube-system").List(&pods).Error
```
#### 选择指定集群
```go
// 选择orb集群,查询集群内kube-system命名空间下的pod
var pods []corev1.Pod
err = kom.Cluster("orb").Resource(&corev1.Pod{}).Namespace("kube-system").List(&pods).Error
```

### 2. 内置资源对象的增删改查以及Watch示例
定义一个 Deployment 对象，并通过 kom 进行资源操作。
```go
var item v1.Deployment
var items []v1.Deployment
```
#### 创建某个资源
```go
item = v1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "nginx",
			Namespace: "default",
		},
		Spec: v1.DeploymentSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{Name: "test", Image: "nginx:1.14.2"},
					},
				},
			},
		},
	}
err := kom.DefaultCluster().Resource(&item).Create(&item).Error
```
#### Get查询某个资源
```go
// 查询 default 命名空间下名为 nginx 的 Deployment
err := kom.DefaultCluster().Resource(&item).Namespace("default").Name("nginx").Get(&item).Error
// 查询 default 命名空间下名为 nginx 的 Deployment，并使用缓存 5 秒
// 5秒内，不会再次查询，批量操作、高频操作下，建议启用缓存
err := kom.DefaultCluster().Resource(&item).Namespace("default").Name("nginx").WithCache(5 * time.Second).Get(&item).Error
```
#### List查询资源列表
```go
// 查询 default 命名空间下的 Deployment 列表
err := kom.DefaultCluster().Resource(&item).Namespace("default").List(&items).Error
// 查询 default、kube-system 命名空间下的 Deployment 列表
err := kom.DefaultCluster().Resource(&item).Namespace("default","kube-system").List(&items).Error
// 查询 所有 命名空间下的 Deployment 列表
err := kom.DefaultCluster().Resource(&item).Namespace("*").List(&items).Error
err := kom.DefaultCluster().Resource(&item).AllNamespace().List(&items).Error
// 设置5秒缓存，对列表生效
err := kom.DefaultCluster().Resource(&item).WithCache(5 * time.Second).List(&nodeList).Error
```
#### 通过Label查询资源列表
```go
// 查询 default 命名空间下 标签为 app:nginx 的 Deployment 列表
err := kom.DefaultCluster().Resource(&item).Namespace("default").WithLabelSelector("app=nginx").List(&items).Error
```
#### 通过多个Label查询资源列表
```go
// 查询 default 命名空间下 标签为 app:nginx m:n 的 Deployment 列表
err := kom.DefaultCluster().Resource(&item).Namespace("default").WithLabelSelector("app=nginx").WithLabelSelector("m=n").List(&items).Error
```
#### 通过Field查询资源列表
```go
// 查询 default 命名空间下 标签为 metadata.name=test-deploy 的 Deployment 列表
// filedSelector 一般支持原生的字段定义。如metadata.name,metadata.namespace,metadata.labels,metadata.annotations,metadata.creationTimestamp,spec.nodeName,spec.serviceAccountName,spec.schedulerName,status.phase,status.hostIP,status.podIP,status.qosClass,spec.containers.name等字段
err := kom.DefaultCluster().Resource(&item).Namespace("default").WithFieldSelector("metadata.name=test-deploy").List(&items).Error
```
#### 分页查询资源
```go
var list []corev1.Pod
var total int64
sql := "select * from pod where metadata.namespace=? or metadata.namespace=?     order by  metadata.creationTimestamp desc "
err := kom.DefaultCluster().Sql(sql, "kube-system", "default").
		FillTotalCount(&total).
		Limit(5).
		Offset(10).
		List(&list).Error
fmt.Printf("total %d\n", total)  //返回总数 480
fmt.Printf("Count %d\n", len(list)) //返回条目数=limit=5
```
#### 更新资源内容
```go
// 更新名为nginx 的 Deployment，增加一个注解
err := kom.DefaultCluster().Resource(&item).Namespace("default").Name("nginx").Get(&item).Error
if item.Spec.Template.Annotations == nil {
	item.Spec.Template.Annotations = map[string]string{}
}
item.Spec.Template.Annotations["kom.kubernetes.io/restartedAt"] = time.Now().Format(time.RFC3339)
err = kom.DefaultCluster().Resource(&item).Update(&item).Error
```
#### PATCH 更新资源
```go
// 使用 Patch 更新资源,为名为 nginx 的 Deployment 增加一个标签，并设置副本数为5
patchData := `{
    "spec": {
        "replicas": 5
    },
    "metadata": {
        "labels": {
            "new-label": "new-value"
        }
    }
}`
err := kom.DefaultCluster().Resource(&item).Patch(&item, types.StrategicMergePatchType, patchData).Error
```
#### 删除资源
```go
// 删除名为 nginx 的 Deployment
err := kom.DefaultCluster().Resource(&item).Namespace("default").Name("nginx").Delete().Error
```
#### 强制删除资源
```go
// 删除名为 nginx 的 Deployment
err := kom.DefaultCluster().Resource(&item).Namespace("default").Name("nginx").ForceDelete().Error
```
#### 通用类型资源的获取（适用于k8s内置类型以及CRD）
```go
// 指定GVK获取资源
var list []corev1.Event
err := kom.DefaultCluster().GVK("events.k8s.io", "v1", "Event").Namespace("default").List(&list).Error
```
#### Watch资源变更
```go
// watch default 命名空间下 Pod资源 的变更
var watcher watch.Interface
var pod corev1.Pod
err := kom.DefaultCluster().Resource(&pod).Namespace("default").Watch(&watcher).Error
if err != nil {
	fmt.Printf("Create Watcher Error %v", err)
	return err
}
go func() {
	defer watcher.Stop()

	for event := range watcher.ResultChan() {
		err := kom.DefaultCluster().Tools().ConvertRuntimeObjectToTypedObject(event.Object, &pod)
		if err != nil {
			fmt.Printf("无法将对象转换为 *v1.Pod 类型: %v", err)
			return
		}
		// 处理事件
		switch event.Type {
		case watch.Added:
			fmt.Printf("Added Pod [ %s/%s ]\n", pod.Namespace, pod.Name)
		case watch.Modified:
			fmt.Printf("Modified Pod [ %s/%s ]\n", pod.Namespace, pod.Name)
		case watch.Deleted:
			fmt.Printf("Deleted Pod [ %s/%s ]\n", pod.Namespace, pod.Name)
		}
	}
}()
```
#### Describe查询某个资源
```go
// Describe default 命名空间下名为 nginx 的 Deployment
var describeResult []byte
err := kom.DefaultCluster().Resource(&item).Namespace("default").Name("nginx").Describe(&item).Error
fmt.Printf("describeResult: %s", describeResult)
```

### 3. YAML 创建、更新、删除
```go
yaml := `apiVersion: v1
kind: ConfigMap
metadata:
  name: example-config
  namespace: default
data:
  key: value
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: example-deployment
  namespace: default
spec:
  replicas: 1
  selector:
    matchLabels:
      app: example
  template:
    metadata:
      labels:
        app: example
    spec:
      containers:
        - name: example-container
          image: nginx
`
// 第一次执行Apply为创建，返回每一条资源的执行结果 
results := kom.DefaultCluster().Applier().Apply(yaml)
// 第二次执行Apply为更新，返回每一条资源的执行结果
results = kom.DefaultCluster().Applier().Apply(yaml)
// 删除，返回每一条资源的执行结果
results = kom.DefaultCluster().Applier().Delete(yaml)
```

### 4. Pod 操作
#### 获取日志
```go
// 获取Pod日志
var stream io.ReadCloser
err := kom.DefaultCluster().Namespace("default").Name("random-char-pod").Ctl().Pod().ContainerName("container").GetLogs(&stream, &corev1.PodLogOptions{}).Error
reader := bufio.NewReader(stream)
line, _ := reader.ReadString('\n')
fmt.Println(line)
```
#### 执行命令
在Pod内执行命令，需要指定容器名称，并且会触发Exec()类型的callbacks。
```go
// 在Pod内执行ps -ef命令
var execResult string
err := kom.DefaultCluster().Namespace("default").Name("random-char-pod").Ctl().Pod().ContainerName("container").Command("ps", "-ef").ExecuteCommand(&execResult).Error
fmt.Printf("execResult: %s", execResult)
```
#### 端口转发
```go
err := kom.DefaultCluster().Resource(&v1.Pod{}).
		Namespace("default").
		Name("nginx-deployment-f576985cc-7czqr").
    Ctl().Pod().
		ContainerName("nginx").
		PortForward("20088", "80", stopCh).Error
// 监听0.0.0.0上的20088端口，转发到Pod的80端口
```
#### 流式执行命令
在Pod内执行命令，并且会触发StreamExec()类型的callbacks。适合执行ping 等命令
```go
cb := func(data []byte) error {
		fmt.Printf("Data %s\n", string(data))
		return nil
	}
err := kom.DefaultCluster().Namespace("kube-system").Name("traefik-d7c9c5778-p9nf4").Ctl().Pod().ContainerName("traefik").Command("ping", "127.0.0.1").StreamExecute(cb, cb).Error
//输出：
//Data PING 127.0.0.1 (127.0.0.1): 56 data bytes
//Data 64 bytes from 127.0.0.1: seq=0 ttl=42 time=0.023 ms
//Data 64 bytes from 127.0.0.1: seq=1 ttl=42 time=0.011 ms
//Data 64 bytes from 127.0.0.1: seq=2 ttl=42 time=0.012 ms
//Data 64 bytes from 127.0.0.1: seq=3 ttl=42 time=0.016 ms
```

#### 文件列表
```go
// 获取Pod内/etc文件夹列表
kom.DefaultCluster().Namespace("default").Name("nginx").Ctl().Pod().ContainerName("nginx").ListFiles("/etc")
```
#### 所有文件列表，包括隐藏文件
```go
// 获取Pod内/etc文件夹列表
kom.DefaultCluster().Namespace("default").Name("nginx").Ctl().Pod().ContainerName("nginx").ListAllFiles("/etc")
```
#### 文件下载
```go
// 下载Pod内/etc/hosts文件
kom.DefaultCluster().Namespace("default").Name("nginx").Ctl().Pod().ContainerName("nginx").DownloadFile("/etc/hosts")
```
#### 文件下载(Tar压缩)
```go
// 下载Pod内/etc/hosts文件，以tar方式进行打包后，获取，下载
kom.DefaultCluster().Namespace("default").Name("nginx").Ctl().Pod().ContainerName("nginx").DownloadTarFile("/etc/hosts")
```
#### 文件上传
```go
// 上传文件内容到Pod内/etc/demo.txt文件
kom.DefaultCluster().Namespace("default").Name("nginx").Ctl().Pod().ContainerName("nginx").SaveFile("/etc/demo.txt", "txt-context")
// os.File 类型文件直接上传到Pod内/etc/目录下
file, _ := os.Open(tempFilePath)
kom.DefaultCluster().Namespace("default").Name("nginx").Ctl().Pod().ContainerName("nginx").UploadFile("/etc/", file)
```
#### 文件删除
```go
// 删除Pod内/etc/xyz文件
kom.DefaultCluster().Namespace("default").Name("nginx").Ctl().Pod().ContainerName("nginx").DeleteFile("/etc/xyz")
```
#### 获取关联资源-Service
```go
// 获取Pod关联的Service
svcs, err := kom.DefaultCluster().Namespace("default").Name("nginx").Ctl().Pod().LinkedService()
for _, svc := range svcs {
	fmt.Printf("service name %v\n", svc.Name)
}
```
#### 获取关联资源-Ingress
```go
// 获取Pod关联的Ingress
ingresses, err := kom.DefaultCluster().Namespace("default").Name("nginx").Ctl().Pod().LinkedIngress()
for _, ingress := range ingresses {
	fmt.Printf("ingress name %v\n", ingress.Name)
}
```
#### 获取关联资源-PVC
```go
// 获取Pod关联的PVC
pvcs, err := kom.DefaultCluster().Namespace("default").Name("nginx").Ctl().Pod().LinkedPVC()
for _, pvc := range pvcs {
	fmt.Printf("pvc name %v\n", pvc.Name)
}
``` 
#### 获取关联资源-PV
```go
// 获取Pod关联的PVC
pvs, err := kom.DefaultCluster().Namespace("default").Name("nginx").Ctl().Pod().LinkedPV()
for _, pv := range pvs {
	fmt.Printf("pv name %v\n", pv.Name)
}
``` 
#### 获取关联资源-Endpoints
```go
// 获取Pod关联的Endpoints
endpoints, err := kom.DefaultCluster().Namespace("default").Name("nginx").Ctl().Pod().LinkedEndpoints()
for _, endpoint := range endpoints {
	fmt.Printf("endpoint name %v\n", endpoint.Name)
}
```
#### 获取关联资源-运行时Env
从Pod内执行env命令获得ENV配置信息
```go
envs, err := kom.DefaultCluster().Namespace("default").Name("nginx").Ctl().Pod().LinkedEnv()
for _, env := range envs {
		fmt.Printf("env %s %s=%s\n", env.ContainerName, env.EnvName, env.EnvValue)
	}
```
#### 获取关联资源-定义Env
从pod定义上提取ENV配置信息
```go
envs, err := kom.DefaultCluster().Namespace("default").Name("nginx").Ctl().Pod().LinkedEnvFromPod()
for _, env := range envs {
		fmt.Printf("env %s %s=%s\n", env.ContainerName, env.EnvName, env.EnvValue)
	}
```
#### 获取关联资源-节点
根据Pod 定义中声明的NodeSelector、NodeAffinity、污点容忍度、NodeName等配置信息，返回可用节点列表。暂未考虑Pod亲和性、CPU内存等运行时调度因素。
```go
nodes, err := kom.DefaultCluster().Namespace("default").Name("nginx").Ctl().Pod().LinkedNode()
for _, node := range nodes {
    fmt.Printf("reason:%s\t node name %s\n", node.Reason, node.Name)
}
```

### 5. 自定义资源定义（CRD）增删改查及Watch操作
在没有CR定义的情况下，如何进行增删改查操作。操作方式同k8s内置资源。
将对象定义为unstructured.Unstructured，并且需要指定Group、Version、Kind。
因此可以通过kom.DefaultCluster().GVK(group, version, kind)来替代kom.DefaultCluster().Resource(interface{})
为方便记忆及使用，kom提供了kom.DefaultCluster().CRD(group, version, kind)来简化操作。
下面给出操作CRD的示例：
首先定义一个通用的处理对象，用来接收CRD的返回结果。
```go
var item unstructured.Unstructured
```
#### 创建CRD
```go
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
    shortNames:
    - ct`
result := kom.DefaultCluster().Applier().Apply(yaml)
```
#### 创建CRD的CR对象
```go
item = unstructured.Unstructured{
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
err := kom.DefaultCluster().CRD("stable.example.com", "v1", "CronTab").Namespace(item.GetNamespace()).Name(item.GetName()).Create(&item).Error
```
#### Get获取单个CR对象
```go
err := kom.DefaultCluster().CRD("stable.example.com", "v1", "CronTab").Name(item.GetName()).Namespace(item.GetNamespace()).Get(&item).Error
```
#### List获取CR对象的列表
```go
var crontabList []unstructured.Unstructured
// 查询default命名空间下的CronTab
err := kom.DefaultCluster().CRD("stable.example.com", "v1", "CronTab").Namespace(crontab.GetNamespace()).List(&crontabList).Error
// 查询所有命名空间下的CronTab
err := kom.DefaultCluster().CRD("stable.example.com", "v1", "CronTab").AllNamespace().List(&crontabList).Error
err := kom.DefaultCluster().CRD("stable.example.com", "v1", "CronTab").Namespace("*").List(&crontabList).Error
```
#### 更新CR对象
```go
patchData := `{
    "spec": {
        "image": "patch-image"
    },
    "metadata": {
        "labels": {
            "new-label": "new-value"
        }
    }
}`
err := kom.DefaultCluster().CRD("stable.example.com", "v1", "CronTab").Name(crontab.GetName()).Namespace(crontab.GetNamespace()).Patch(&crontab, types.StrategicMergePatchType, patchData).Error
```
#### 删除CR对象
```go
err := kom.DefaultCluster().CRD("stable.example.com", "v1", "CronTab").Name(crontab.GetName()).Namespace(crontab.GetNamespace()).Delete().Error
```
#### 强制删除CR对象
```go
err := kom.DefaultCluster().CRD("stable.example.com", "v1", "CronTab").Name(crontab.GetName()).Namespace(crontab.GetNamespace()).ForceDelete().Error
```
#### Watch CR对象
```go
var watcher watch.Interface

err := kom.DefaultCluster().CRD("stable.example.com", "v1", "CronTab").Namespace("default").Watch(&watcher).Error
if err != nil {
    fmt.Printf("Create Watcher Error %v", err)
}
go func() {
    defer watcher.Stop()
    
    for event := range watcher.ResultChan() {
    var item *unstructured.Unstructured
    
    item, err := kom.DefaultCluster().Tools().ConvertRuntimeObjectToUnstructuredObject(event.Object)
    if err != nil {
        fmt.Printf("无法将对象转换为 Unstructured 类型: %v", err)
        return
    }
    // 处理事件
    switch event.Type {
        case watch.Added:
            fmt.Printf("Added Unstructured [ %s/%s ]\n", item.GetNamespace(), item.GetName())
        case watch.Modified:
            fmt.Printf("Modified Unstructured [ %s/%s ]\n", item.GetNamespace(), item.GetName())
        case watch.Deleted:
            fmt.Printf("Deleted Unstructured [ %s/%s ]\n", item.GetNamespace(), item.GetName())
        }
    }
}()
```
#### Describe查询某个CRD资源
```go
// Describe default 命名空间下名为 nginx 的 Deployment
var describeResult []byte
err := kom.DefaultCluster().CRD("stable.example.com", "v1", "CronTab").Namespace("default").Name(item.GetName()).Describe(&item).Error
fmt.Printf("describeResult: %s", describeResult)
```
#### 获取CRD下的Pod资源
```go
pods, err := kom.DefaultCluster().CRD("apps.kruise.io", "v1beta1", "StatefulSet").
Namespace("default").Name("sample").Ctl().CRD().ManagedPods()
	for _, pod := range pods {
		fmt.Printf("Get pods: %v", pod.GetName())
	}
```

### 6. 集群参数信息
```go
// 集群文档
kom.DefaultCluster().Status().Docs()
// 集群资源信息
kom.DefaultCluster().Status().APIResources()
// 集群已注册CRD列表
kom.DefaultCluster().Status().CRDList()
// 集群版本信息
kom.DefaultCluster().Status().ServerVersion()
```

### 7. callback机制
* 内置了callback机制，可以自定义回调函数，当执行完某项操作后，会调用对应的回调函数。
* 如果回调函数返回true，则继续执行后续操作，否则终止后续操作。
* 当前支持的callback有：get,list,create,update,patch,delete,exec,stream-exec,logs,watch,doc.
* 内置的callback名称有："kom:get","kom:list","kom:create","kom:update","kom:patch","kom:watch","kom:delete","kom:pod:exec","kom:pod:stream:exec","kom:pod:logs","kom:pod:port:forward","kom:doc"
* 支持回调函数排序，默认按注册顺序执行，可以通过kom.DefaultCluster().Callback().After("kom:get")或者.Before("kom:get")设置顺序。
* 支持删除回调函数，通过kom.DefaultCluster().Callback().Delete("kom:get")
* 支持替换回调函数，通过kom.DefaultCluster().Callback().Replace("kom:get",cb)
```go
// 为Get获取资源注册回调函数
kom.DefaultCluster().Callback().Get().Register("get", cb)
// 为List获取资源注册回调函数
kom.DefaultCluster().Callback().List().Register("list", cb)
// 为Create创建资源注册回调函数
kom.DefaultCluster().Callback().Create().Register("create", cb)
// 为Update更新资源注册回调函数
kom.DefaultCluster().Callback().Update().Register("update", cb)
// 为Patch更新资源注册回调函数
kom.DefaultCluster().Callback().Patch().Register("patch", cb)
// 为Delete删除资源注册回调函数
kom.DefaultCluster().Callback().Delete().Register("delete", cb)
// 为Watch资源注册回调函数
kom.DefaultCluster().Callback().Watch().Register("watch",cb)
// 为Exec Pod内执行命令注册回调函数
kom.DefaultCluster().Callback().Exec().Register("exec", cb)
// 为Logs获取日志注册回调函数
kom.DefaultCluster().Callback().Logs().Register("logs", cb)
// 删除回调函数
kom.DefaultCluster().Callback().Get().Delete("get")
// 替换回调函数
kom.DefaultCluster().Callback().Get().Replace("get", cb)
// 指定回调函数执行顺序，在内置的回调函数执行完之后再执行
kom.DefaultCluster().Callback().After("kom:get").Register("get", cb)
// 指定回调函数执行顺序，在内置的回调函数执行之前先执行
// 案例1.在Create创建资源前，进行权限检查，没有权限则返回error，后续创建动作将不再执行
// 案例2.在List获取资源列表后，进行特定的资源筛选，从列表(Statement.Dest)中删除不符合要求的资源，然后返回给用户
kom.DefaultCluster().Callback().Before("kom:create").Register("create", cb)

// 自定义回调函数
func cb(k *kom.Kubectl) error {
    stmt := k.Statement
    gvr := stmt.GVR
    ns := stmt.Namespace
    name := stmt.Name
    // 打印信息
    fmt.Printf("Get %s/%s(%s)\n", ns, name, gvr)
    fmt.Printf("Command %s/%s(%s %s)\n", ns, name, stmt.Command, stmt.Args)
    return nil
	// return fmt.Errorf("error") 返回error将阻止后续cb的执行
}
```

### 8. SQL查询k8s资源
* 通过SQL()方法查询k8s资源，简单高效。
* Table 名称支持集群内注册的所有资源的全称及简写，包括CRD资源。只要是注册到集群上了，就可以查。
* 典型的Table 名称有：pod,deployment,service,ingress,pvc,pv,node,namespace,secret,configmap,serviceaccount,role,rolebinding,clusterrole,clusterrolebinding,crd,cr,hpa,daemonset,statefulset,job,cronjob,limitrange,horizontalpodautoscaler,poddisruptionbudget,networkpolicy,endpoints,ingressclass,mutatingwebhookconfiguration,validatingwebhookconfiguration,customresourcedefinition,storageclass,persistentvolumeclaim,persistentvolume,horizontalpodautoscaler,podsecurity。统统都可以查。
* 查询字段目前仅支持*。也就是select *
* 查询条件目前支持 =，!=,>=,<=,<>,like,in,not in,and,or,between
* 排序字段目前支持对单一字段进行排序。默认按创建时间倒序排列
* 
#### 查询k8s内置资源
```go
    sql := "select * from deploy where metadata.namespace='kube-system' or metadata.namespace='default' order by  metadata.creationTimestamp asc   "

	var list []v1.Deployment
	err := kom.DefaultCluster().Sql(sql).List(&list).Error
	for _, d := range list {
		fmt.Printf("List Items foreach %s,%s at %s \n", d.GetNamespace(), d.GetName(), d.GetCreationTimestamp())
	}
```
#### 查询CRD资源
```go
    // vm 为kubevirt 的CRD
    sql := "select * from vm where (metadata.namespace='kube-system' or metadata.namespace='default' )  "
	var list []unstructured.Unstructured
	err := kom.DefaultCluster().Sql(sql).List(&list).Error
	for _, d := range list {
		fmt.Printf("List Items foreach %s,%s\n", d.GetNamespace(), d.GetName())
	}
```
#### 链式调研查询SQL
```go
// 查询pod 列表
err := kom.DefaultCluster().From("pod").
		Where("metadata.namespace = ?  or metadata.namespace= ? ", "kube-system", "default").
		Order("metadata.creationTimestamp desc").
		List(&list).Error
```
#### k8s资源嵌套列表属性支持
```go
// spec.containers为列表，其下的ports也为列表，我们查询ports的name
sql := "select * from pod where spec.containers.ports.name like '%k8m%'  "
var list []v1.Pod
err := kom.DefaultCluster().Sql(sql).List(&list).Error
for _, d := range list {
	t.Logf("List Items foreach %s,%s\n", d.GetNamespace(), d.GetName())
}
```

### 9. 其他操作
#### Deployment重启
```go
err = kom.DefaultCluster().Resource(&Deployment{}).Namespace("default").Name("nginx").Ctl().Rollout().Restart()
```
#### Deployment扩缩容
```go
// 将名称为nginx的deployment的副本数设置为3
err = kom.DefaultCluster().Resource(&Deployment{}).Namespace("default").Name("nginx").Ctl().Scaler().Scale(3)
```
#### Deployment 停止
```go
// 将名称为nginx的deployment的副本数设置为0
// 当前运行副本数量记录到注解中
err = kom.DefaultCluster().Resource(&Deployment{}).Namespace("default").Name("nginx").Ctl().Scaler().Stop()
```
#### Deployment 恢复
```go
// 将名称为nginx的deployment的副本数从注解中恢复，如果没有注解，默认恢复到1
err = kom.DefaultCluster().Resource(&Deployment{}).Namespace("default").Name("nginx").Ctl().Scaler().Restore()
```
#### Deployment更新Tag
```go
// 将名称为nginx的deployment的中的容器镜像tag升级为alpine
err = kom.DefaultCluster().Resource(&Deployment{}).Namespace("default").Name("nginx").Ctl().Deployment().ReplaceImageTag("main","20241124")
```
#### Deployment Rollout History
```go
// 查询名称为nginx的deployment的升级历史
result, err := kom.DefaultCluster().Resource(&Deployment{}).Namespace("default").Name("nginx").Ctl().Rollout().History()
```
#### Deployment Rollout Undo
```go
// 将名称为nginx的deployment进行回滚
result, err := kom.DefaultCluster().Resource(&Deployment{}).Namespace("default").Name("nginx").Ctl().Rollout().Undo()
// 将名称为nginx的deployment进行回滚到指定版本(history 查询)
result, err := kom.DefaultCluster().Resource(&Deployment{}).Namespace("default").Name("nginx").Ctl().Rollout().Undo("6")
```
#### Deployment Rollout Pause
```go
// 暂停升级过程
err := kom.DefaultCluster().Resource(&Deployment{}).Namespace("default").Name("nginx").Ctl().Rollout().Pause()
```
#### Deployment Rollout Resume 
```go
// 恢复升级过程
err := kom.DefaultCluster().Resource(&Deployment{}).Namespace("default").Name("nginx").Ctl().Rollout().Resume()
```
#### Deployment Rollout Status 
```go
// 将名称为nginx的deployment的中的容器镜像tag升级为alpine
result, err := kom.DefaultCluster().Resource(&Deployment{}).Namespace("default").Name("nginx").Ctl().Rollout().Status()
```
#### Deployment HPA
```go
// 显示deployment的hpa 
list, err := kom.DefaultCluster().Resource(&v1.Deployment{}).Namespace("default").Name("nginx-web").Ctl().Deployment().HPAList()
for _, item := range list {
    t.Logf("HPA %s\n", item.Name)
}
```
#### 节点打污点
```go
err = kom.DefaultCluster().Resource(&Node{}).Name("kind-control-plane").Ctl().Node().Taint("dedicated=special-user:NoSchedule")
```
#### 节点去除污点
```go
err = kom.DefaultCluster().Resource(&Node{}).Name("kind-control-plane").Ctl().Node().UnTaint("dedicated=special-user:NoSchedule")
```
#### 节点Cordon
```go
err = kom.DefaultCluster().Resource(&Node{}).Name("kind-control-plane").Ctl().Node().Cordon()
```
#### 节点UnCordon
```go
err = kom.DefaultCluster().Resource(&Node{}).Name("kind-control-plane").Ctl().Node().UnCordon()
```
#### 节点Drain
```go
err = kom.DefaultCluster().Resource(&Node{}).Name("kind-control-plane").Ctl().Node().Drain()
```
#### 查询节点IP资源情况
支持设置缓存时间，避免频繁查询k8s API
```go
nodeName := "lima-rancher-desktop"
total, used, available := kom.DefaultCluster().Resource(&corev1.Node{}).WithCache(5 * time.Second).Name(nodeName).Ctl().Node().IPUsage()
fmt.Printf("Total %d, Used %d, Available %d\n", total, used, available)
//Total 256, Used 6, Available 250
```
#### 节点IP资源使用情况统计
支持设置缓存时间，避免频繁查询k8s API
```go
nodeName := "lima-rancher-desktop"
total, used, available := kom.DefaultCluster().Resource(&corev1.Node{}).WithCache(5 * time.Second).Name(nodeName).Ctl().Node().PodCount()
fmt.Printf("Total %d, Used %d, Available %d\n", total, used, available)
//Total 110, Used 9, Available 101
```
#### 节点资源用量情况统计
支持设置缓存时间，避免频繁查询k8s API
```go
nodeName := "lima-rancher-desktop"
usage := kom.DefaultCluster().Resource(&corev1.Node{}).WithCache(5 * time.Second).Name(nodeName).Ctl().Node().ResourceUsage()
fmt.Printf("Node Usage %s\n", utils.ToJSON(usage))
```
包括当前的请求值、限制值、可分配值、使用比例
```json
{
  "requests": {
    "cpu": "200m",
    "memory": "140Mi"
  },
  "limits": {
    "memory": "170Mi"
  },
  "allocatable": {
    "cpu": "4",
    "ephemeral-storage": "99833802265",
    "hugepages-1Gi": "0",
    "hugepages-2Mi": "0",
    "hugepages-32Mi": "0",
    "hugepages-64Ki": "0",
    "memory": "8127096Ki",
    "pods": "110"
  },
  "usageFractions": {
    "cpu": {
      "requestFraction": 5,
      "limitFraction": 0
    },
    "ephemeral-storage": {
      "requestFraction": 0,
      "limitFraction": 0
    },
    "memory": {
      "requestFraction": 1.76397571777176,
      "limitFraction": 2.1419705144371375
    }
  }
}
```
#### 给资源增加标签
```go
err = kom.DefaultCluster().Resource(&Node{}).Name("kind-control-plane").Ctl().Label("name=zhangsan")
```
#### 给资源删除标签
```go
err = kom.DefaultCluster().Resource(&Node{}).Name("kind-control-plane").Ctl().Label("name-")
```
#### 给资源增加注解
```go
err = kom.DefaultCluster().Resource(&Node{}).Name("kind-control-plane").Ctl().Annotate("name=zhangsan")
```
#### 给资源删除注解
```go
err = kom.DefaultCluster().Resource(&Node{}).Name("kind-control-plane").Ctl().Annotate("name-")
```
#### 创建NodeSell
```go
ns, pod, container, err  := kom.DefaultCluster().Resource(&v1.Node{}).Name("kind-control-plane").Ctl().Node().CreateNodeShell()
fmt.Printf("Node Shell ns=%s podName=%s containerName=%s", ns, pod, container)
```
#### 创建kubectl Shell
```go
ns, pod, container, err := kom.DefaultCluster().Resource(&v1.Node{}).Name(name).Ctl().Node().CreateKubectlShell(kubeconfig)
fmt.Printf("Kubectl Shell ns=%s podName=%s containerName=%s", ns, pod, container)

```
#### 统计StorageClass下的PVC数量
```go
count, err := kom.DefaultCluster().Resource(&v1.StorageClass{}).Name("hostpath").Ctl().StorageClass().PVCCount()
fmt.Printf("pvc count %d\n", count)
```
#### 统计StorageClass下的PV数量
```go
count, err := kom.DefaultCluster().Resource(&v1.StorageClass{}).Name("hostpath").Ctl().StorageClass().PVCount()
fmt.Printf("pv count %d\n", count)
```
#### 设置StorageClass为默认
```go
err := kom.DefaultCluster().Resource(&v1.StorageClass{}).Name("hostpath").Ctl().StorageClass().SetDefault()
```
#### 设置IngressClass为默认
```go
err := kom.DefaultCluster().Resource(&v1.IngressClass{}).Name("nginx").Ctl().IngressClass().SetDefault()
```
#### 统计Deployment/StatefulSet/DaemonSet下的Pod列表
```go
list, err := kom.DefaultCluster().Namespace("default").Name("managed-pods").Ctl().Deployment().ManagedPods()
for _, pod := range list {
	fmt.Printf("ManagedPod: %v", pod.Name)
}
```
#### 获取所有节点的标签集合
```go
// labels 类型为map[string]string
labels, err := kom.DefaultCluster().Resource(&v1.Node{}).Ctl().Node().AllNodeLabels()
fmt.Printf("%s", utils.ToJSON(labels))
```
```json
{
          "beta.kubernetes.io/arch": "arm64",
          "beta.kubernetes.io/os": "linux",
          "kubernetes.io/arch": "arm64",
          "kubernetes.io/hostname": "kind-control-plane",
          "kubernetes.io/os": "linux",
          "kubernetes.io/role": "agent",
          "node-role.kubernetes.io/agent": "",
          "node-role.kubernetes.io/control-plane": "",
          "type": "kwok",
          "uat": "test",
          "x": "x"
}
```
#### 查看Pod资源占用率
```go
podName := "coredns-ccb96694c-jprpf"
ns := "kube-system"
usage := kom.DefaultCluster().Resource(&corev1.Pod{}).Name(podName).Namespace(ns).Ctl().Pod().ResourceUsage()
fmt.Printf("Pod Usage %s\n", utils.ToJSON(usage))
```
包括当前的请求值、限制值、可分配值、使用比例
```json
{
  "requests": {
    "cpu": "100m",
    "memory": "70Mi"
  },
  "limits": {
    "memory": "170Mi"
  },
  "allocatable": {
    "cpu": "4",
    "ephemeral-storage": "99833802265",
    "hugepages-1Gi": "0",
    "hugepages-2Mi": "0",
    "hugepages-32Mi": "0",
    "hugepages-64Ki": "0",
    "memory": "8127096Ki",
    "pods": "110"
  },
  "usageFractions": {
    "cpu": {
      "requestFraction": 2.5,
      "limitFraction": 0
    },
    "memory": {
      "requestFraction": 0.88198785888588,
      "limitFraction": 2.1419705144371375
    }
  }
}
```
#### 获取字段文档解释
```go
var docResult []byte
	item := v1.Deployment{}
	field := "spec.replicas"
	field = "spec.template.spec.containers.name"
	field = "spec.template.spec.containers.imagePullPolicy"
	field = "spec.template.spec.containers.livenessProbe.successThreshold"
	err := kom.DefaultCluster().
		Resource(&item).DocField(field).Doc(&docResult).Error
	fmt.Printf("Get Deployment Doc [%s] :%s", field, string(docResult))
```
## 联系我

微信（大罗马的太阳） 搜索ID：daluomadetaiyang,备注kom。
<br><img width="214" alt="Image" src="https://github.com/user-attachments/assets/166db141-42c5-42c4-9964-8e25cf12d04c" />

## 微信群


![输入图片说明](https://foruda.gitee.com/images/1751469149301951831/3f8cf2ad_77493.png "屏幕截图")
