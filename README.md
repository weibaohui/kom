
# Kom - Kubernetes Operations Manager

## 简介

`kom` 是一个用于 Kubernetes 操作的工具，提供了一系列功能来管理 Kubernetes 资源，包括创建、更新、删除和获取资源。这个项目支持多种 Kubernetes 资源类型的操作，并能够处理自定义资源定义（CRD）。
通过使用 `kom`，你可以轻松地进行资源的增删改查和日志获取以及操作POD内文件等动作。

## **特点**
1. 简单易用：kom 提供了丰富的功能，包括创建、更新、删除、获取、列表等。
2. 多集群支持：通过RegisterCluster，你可以轻松地管理多个 Kubernetes 集群。
3. 链式调用：kom 提供了链式调用，使得操作资源更加简单和直观。

## 示例程序
**k8m** 是一个轻量级的 Kubernetes 管理工具，它基于kom、amis实现，单文件，支持多平台架构。
1. **下载**：从 [https://github.com/weibaohui/k8m](https://github.com/weibaohui/k8m) 下载最新版本。
2. **运行**：使用 `./k8m` 命令启动,访问[http://127.0.0.1:3618](http://127.0.0.1:3618)。




## 安装

确保你的环境中已经安装 Go 语言和 Kubernetes CLI（kubectl）。

```bash
go get github.com/weibaohui/kom
```

## 使用示例

以下是一些基本的使用示例。

### 1. 内置资源对象的增删改查示例
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
// 查询 default 命名空间下 标签为 app=nginx 的 Deployment 列表
err := kom.DefaultCluster().Resource(&item).Namespace("default").List(&items, metav1.ListOptions{LabelSelector: "app=nginx"}).Error
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
### 2. YAML 创建、更新、删除
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

### 3. 自定义资源定义（CRD）
在没有CR定义的情况下，如何进行增删改查操作。操作方式同k8s内置资源。
只不过将对象定义为unstructured.Unstructured，并且需要指定Group、Version、Kind。
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
### 4. 多集群管理
#### 注册多集群
```go
// 注册InCluster集群，名称为InCluster
kom.Clusters().RegisterInCluster()
// 注册两个带名称的集群,分别名为orb和docker-desktop
kom.Clusters().RegisterByPathWithID("/Users/kom/.kube/orb", "orb")
kom.Clusters().RegisterByPathWithID("/Users/kom/.kube/config", "docker-desktop")
```
#### 显示已注册集群
```go
kom.Clusters().Show()
```
#### 多集群选择
```go
// 选择集群,查询集群内kube-system命名空间下的pod
var pods []corev1.Pod
err = kom.Cluster("orb").Resource(&corev1.Pod{}).Namespace("kube-system").List(&pods).Error
```


### 5. Pod 日志获取
```go
// 获取Pod日志
options := corev1.PodLogOptions{
    Container: "nginx",
}
podLogs := kom.DefaultCluster().Poder().Namespace("default").Name("nginx").GetLogs("nginx", &options)
logStream, err := podLogs.Stream(context.TODO())
reader := bufio.NewReader(logStream)
```
### 6. Pod 操作
#### 执行命令
```go
// 在Pod内执行ps -ef命令
stdout, stderr, err := kom.DefaultCluster().Poder().Namespace("default").Name("nginx").ExecuteCommand("ps", "-ef")
```
#### 文件列表
```go
// 获取Pod内/etc文件夹列表
kom.DefaultCluster().Poder().Namespace("default").Name("nginx").ContainerName("nginx").GetFileList("/etc")
```
#### 文件下载
```go
// 下载Pod内/etc/hosts文件
kom.DefaultCluster().Poder().Namespace("default").Name("nginx").ContainerName("nginx").DownloadFile("/etc/hosts")
```
#### 文件上传
```go
// 上传文件内容到Pod内/etc/demo.txt文件
kom.DefaultCluster().Poder().Namespace("default").Name("nginx").ContainerName("nginx").SaveFile("/etc/demo.txt", "txt-context")
```
### 7. 集群参数信息
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
