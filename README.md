
# Kom - Kubernetes Operations Manager

## 简介

`kom` 是一个用于 Kubernetes 操作的工具，提供了一系列功能来管理 Kubernetes 资源，包括创建、更新、删除和获取资源。这个项目支持多种 Kubernetes 资源类型的操作，并能够处理自定义资源定义（CRD）。
通过使用 `kom`，你可以轻松地进行资源的增删改查和日志获取以及操作POD内文件等动作。

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

### 1. 基本示例


### 2. YAML 应用与删除

此示例展示了如何应用和删除 YAML 配置。

```go
func YamlApplyDelete() {
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
	result := applier.Instance().WithContext(context.TODO()).Apply(yaml)
	for _, r := range result {
		fmt.Println(r)
	}
	result = applier.Instance().WithContext(context.TODO()).Delete(yaml)
	for _, r := range result {
		fmt.Println(r)
	}
}
```

### 3. 自定义资源定义（CRD）

演示如何管理自定义资源定义。

```go
func crdExample() {
	var crontab unstructured.Unstructured
	crontab = unstructured.Unstructured{
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
	// 删除CRD
	err := kom.Init().WithContext(context.TODO()).CRD("stable.example.com", "v1", "CronTab").Name(crontab.GetName()).Namespace(crontab.GetNamespace()).Delete().Error
	if err != nil {
		fmt.Printf("CronTab Delete error: %v\n", err)
	}
	// 创建CRD
	err = kom.Init().WithContext(context.TODO()).CRD("stable.example.com", "v1", "CronTab").Name(crontab.GetName()).Namespace(crontab.GetNamespace()).Create(&crontab).Error
	if err != nil {
		fmt.Printf("CRD Create error: %v\n", err)
	}
	// 获取CRD
	err = kom.Init().WithContext(context.TODO()).CRD("stable.example.com", "v1", "CronTab").Name(crontab.GetName()).Namespace(crontab.GetNamespace()).Get(&crontab).Error
	if err != nil {
		fmt.Printf("CRD Get error: %v\n", err)
	}
}
```

### 4. 资源的增删改查

以下示例展示了如何进行资源的增删改查操作。

#### 创建资源

```go
func CreateResource() {
	item := v1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-deploy",
			Namespace: "default",
		},
		Spec: v1.DeploymentSpec{
			Replicas: utils.Int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": "test"},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app": "test"},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{Name: "test", Image: "nginx:1.14.2"},
					},
				},
			},
		},
	}

	err := kom.Init().WithContext(context.TODO()).Resource(&item).Create(&item).Error
	if err != nil {
		fmt.Printf("Deployment Create error: %v\n", err)
	}
}
```

#### 获取资源

```go
func GetResource() {
	var item v1.Deployment
	err := kom.Init().WithContext(context.TODO()).Resource(&item).Namespace("default").Name("test-deploy").Get(&item).Error
	if err != nil {
		fmt.Printf("Deployment Get error: %v\n", err)
	}
	fmt.Printf("Get Item: %s\n", item.Spec.Template.Spec.Containers[0].Image)
}
```

#### 更新资源

```go
func UpdateResource() {
	var item v1.Deployment
	err := kom.Init().WithContext(context.TODO()).Resource(&item).Namespace("default").Name("test-deploy").Get(&item).Error
	if err == nil {
		item.Spec.Template.Annotations["kom.kubernetes.io/restartedAt"] = time.Now().Format(time.RFC3339)
		err = kom.Init().WithContext(context.TODO()).Resource(&item).Update(&item).Error
		if err != nil {
			fmt.Printf("Deployment Update error: %v\n", err)
		}
	}
}
```

#### 删除资源

```go
func DeleteResource() {
	var item v1.Deployment
	err := kom.Init().WithContext(context.TODO()).Resource(&item).Namespace("default").Name("test-deploy").Delete().Error
	if err != nil {
		fmt.Printf("Deployment Delete error: %v\n", err)
	}
}
```

### 5. 列表操作

以下是如何列出 Kubernetes 资源的示例。

```go
func ListResources() {
	var items []v1.Deployment
	err := kom.Init().WithContext(context.TODO()).Resource(&v1.Deployment{}).Namespace("default").List(&items).Error
	if err != nil {
		fmt.Printf("List Error: %v\n", err)
	}
	fmt.Printf("List Deployment count: %d\n", len(items))
	for _, d := range items {
		fmt.Printf("List Deployment Items: %s, %s\n", d.Namespace, d.Spec.Template.Spec.Containers[0].Image)
	}
}
```

#### 使用选项进行列表操作

你可以使用 `ListOptions` 来过滤列表结果。

```go
func ListWithOptions() {
	var items []v1.Deployment
	err := kom.Init().WithContext(context.TODO()).Resource(&v1.Deployment{}).Namespace("default").
		List(&items, metav1.ListOptions{LabelSelector: "app=nginx"}).Error
	if err != nil {
		fmt.Printf("List Error: %v\n", err)
	}
	fmt.Printf("List Deployment WithLabelSelector app=nginx count: %d\n", len(items))
	for _, d := range items {
		fmt.Printf("List Deployment WithLabelSelector Items: %s, %s\n", d.Namespace, d.Spec.Template.Spec.Containers[0].Image)
	}
}
```

### 6. Pod 日志获取

获取特定 Pod 的日志。

```go
func PodLogs() {
	yaml := `apiVersion: v1
kind: Pod
metadata:
  name: random-char-pod
  namespace: default
spec:
  containers:
  - name: container-b
    image: alpine
    command: ["/bin/sh", "-c", "while true; do echo $(cat /dev/urandom | tr -dc 'a-zA-Z0-9' | head -c 1); sleep 5; done"]
`
	_ = applier.Instance().WithContext(context.TODO()).Delete(yaml)
	_ = applier.Instance().WithContext(context.TODO()).Apply(yaml)

	time.Sleep(5 * time.Second)
	options := corev1.PodLogOptions{Container: "container-b"}
	podLogs := poder.Instance().WithContext(context.TODO()).Namespace("default").Name("random-char-pod").GetLogs("random-char-pod", &options)
	logStream, err := podLogs.Stream(context.TODO())
	if err != nil {
		fmt.Println("Error getting pod logs:", err)
		return
	}
	// 逐行读取日志
	reader := bufio.NewReader(logStream)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break


			}
			fmt.Println("Error reading log:", err)
			break
		}
		fmt.Print(line)
	}
}
```