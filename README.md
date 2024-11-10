
# Kom - Kubernetes Operations Manager

## 简介

`kom` 是一个用于 Kubernetes 操作的工具，提供了一系列功能来管理 Kubernetes 资源，包括创建、更新、删除和获取资源。这个项目支持多种 Kubernetes 资源类型的操作，并能够处理自定义资源定义（CRD）。
通过使用 `kom`，你可以轻松地进行资源的增删改查和日志获取以及操作POD内文件等动作。

## **特点**
1. 简单易用：kom 提供了丰富的功能，包括创建、更新、删除、获取、列表等，包括对内置资源以及CRD资源的操作。
2. 多集群支持：通过RegisterCluster，你可以轻松地管理多个 Kubernetes 集群。
3. 链式调用：kom 提供了链式调用，使得操作资源更加简单和直观。
4. 支持自定义资源定义（CRD）：kom 支持自定义资源定义（CRD），你可以轻松地定义和操作自定义资源。
5. 支持回调机制，轻松拓展业务逻辑，而不必跟k8s操作强耦合。
6. 支持POD内文件操作，轻松上传、下载、删除文件。

## 示例程序
**k8m** 是一个轻量级的 Kubernetes 管理工具，它基于kom、amis实现，单文件，支持多平台架构。
1. **下载**：从 [https://github.com/weibaohui/k8m](https://github.com/weibaohui/k8m) 下载最新版本。
2. **运行**：使用 `./k8m` 命令启动,访问[http://127.0.0.1:3618](http://127.0.0.1:3618)。




## 安装

```bash
import github.com/weibaohui/kom

func main() {
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
```
#### List查询资源列表
```go
// 查询 default 命名空间下的 Deployment 列表
err := kom.DefaultCluster().Resource(&item).Namespace("default").List(&items).Error
```
#### 通过Label查询资源列表
```go
// 查询 default 命名空间下 标签为 app:nginx 的 Deployment 列表
err := kom.DefaultCluster().Resource(&item).Namespace("default").List(&items, metav1.ListOptions{LabelSelector: "app=nginx"}).Error
```
#### 通过多个Label查询资源列表
```go
// 查询 default 命名空间下 标签为 app:nginx m:n 的 Deployment 列表
err := kom.DefaultCluster().Resource(&item).Namespace("default").List(&items, metav1.ListOptions{LabelSelector: "app=nginx,m=n"}).Error
```
#### 通过Field查询资源列表
```go
// 查询 default 命名空间下 标签为 metadata.name=test-deploy 的 Deployment 列表
// filedSelector 一般支持原生的字段定义。如metadata.name,metadata.namespace,metadata.labels,metadata.annotations,metadata.creationTimestamp,spec.nodeName,spec.serviceAccountName,spec.schedulerName,status.phase,status.hostIP,status.podIP,status.qosClass,spec.containers.name等字段
err := kom.DefaultCluster().Resource(&item).Namespace("default").List(&items, metav1.ListOptions{FieldSelector: "metadata.name=test-deploy"}).Error
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
err := kom.DefaultCluster().Resource(&item).Patch(&item, types.MergePatchType, patchData).Error
```
#### 删除资源
```go
// 删除名为 nginx 的 Deployment
err := kom.DefaultCluster().Resource(&item).Namespace("default").Name("nginx").Delete().Error
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
		err := kom.Tools().ConvertRuntimeObjectToTypedObject(event.Object, &pod)
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
err := kom.DefaultCluster().Namespace("default").Name("random-char-pod").ContainerName("container").GetLogs(&stream, &corev1.PodLogOptions{}).Error
reader := bufio.NewReader(stream)
line, _ := reader.ReadString('\n')
fmt.Println(line)
```
#### 执行命令
在Pod内执行命令，需要指定容器名称，并且会触发Exec()类型的callbacks。
```go
// 在Pod内执行ps -ef命令
var execResult string
err := kom.DefaultCluster().Namespace("default").Name("random-char-pod").ContainerName("container").Command("ps", "-ef").ExecuteCommand(&execResult).Error
fmt.Printf("execResult: %s", execResult)
```
#### 文件列表
```go
// 获取Pod内/etc文件夹列表
kom.DefaultCluster().Namespace("default").Name("nginx").ContainerName("nginx").Poder().ListFiles("/etc")
```
#### 文件下载
```go
// 下载Pod内/etc/hosts文件
kom.DefaultCluster().Namespace("default").Name("nginx").ContainerName("nginx").Poder().DownloadFile("/etc/hosts")
```
#### 文件上传
```go
// 上传文件内容到Pod内/etc/demo.txt文件
kom.DefaultCluster().Namespace("default").Name("nginx").ContainerName("nginx").Poder().SaveFile("/etc/demo.txt", "txt-context")
// multipart.File 类型文件直接上传到Pod内
kom.DefaultCluster().Namespace("default").Name("nginx").ContainerName("nginx").Poder().UploadFile("/etc/demo.txt", multipart.File)
```
#### 文件删除
```go
// 删除Pod内/etc/xyz文件
kom.DefaultCluster().Namespace("default").Name("nginx").ContainerName("nginx").Poder().DeleteFile("/etc/xyz")
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
err := kom.DefaultCluster().CRD("stable.example.com", "v1", "CronTab").Namespace(crontab.GetNamespace()).List(&crontabList).Error
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
err := kom.DefaultCluster().CRD("stable.example.com", "v1", "CronTab").Name(crontab.GetName()).Namespace(crontab.GetNamespace()).Patch(&crontab, types.MergePatchType, patchData).Error
```
#### 删除CR对象
```go
err := kom.DefaultCluster().CRD("stable.example.com", "v1", "CronTab").Name(crontab.GetName()).Namespace(crontab.GetNamespace()).Delete().Error
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
    
    item, err := kom.Tools().ConvertRuntimeObjectToUnstructuredObject(event.Object)
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

### 7.callback机制
* 内置了callback机制，可以自定义回调函数，当执行完某项操作后，会调用对应的回调函数。
* 如果回调函数返回true，则继续执行后续操作，否则终止后续操作。
* 当前支持的callback有：get,list,create,update,patch,delete,exec,logs,watch.
* 内置的callback名称有："kom:get","kom:list","kom:create","kom:update","kom:patch","kom:watch","kom:delete","kom:pod:exec","kom:pod:logs"
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