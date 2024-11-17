package example

import (
	"bufio"
	"fmt"
	"io"
	"time"

	"github.com/weibaohui/kom/kom"
	"github.com/weibaohui/kom/utils"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/klog/v2"
)

func Example() {
	callbacks()
	_ = InitPodWatcher()
	builtInExample()
	// crdExample()
	// yamlApplyDelete()
	// multiCluster()
	// newEventList()
	// coreEventList()
	// doc()
	// fetchDoc1()
	// fetchDoc2()
	// podCommand()
	// podFileCommand()
	// podLogs()

}
func callbacks() {
	_ = kom.DefaultCluster().Callback().Get().Register("get", GetCB)
	_ = kom.DefaultCluster().Callback().Exec().Register("get", GetCB)
}
func GetCB(k *kom.Kubectl) error {

	stmt := k.Statement
	gvr := stmt.GVR
	ns := stmt.Namespace
	name := stmt.Name

	// 打印信息
	fmt.Printf("Callback:Get %s/%s(%s)\n", ns, name, gvr)
	fmt.Printf("Callback:Command %s/%s(%s %s)\n", ns, name, stmt.Command, stmt.Args)
	return nil
}

// 初始化CRD的Watch
func initCRDWatcher() error {
	var watcher watch.Interface

	err := kom.DefaultCluster().
		CRD("stable.example.com", "v1", "CronTab").
		Namespace("default").Watch(&watcher).Error
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
	return err
}
func InitPodWatcher() error {

	var watcher watch.Interface

	var pod corev1.Pod
	err := kom.DefaultCluster().Resource(&pod).
		Namespace("default").Watch(&watcher).Error
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

	return nil
}
func yamlApplyDelete() {
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
	results := kom.DefaultCluster().Applier().Apply(yaml)
	for _, r := range results {
		fmt.Println(r)
	}
	results = kom.DefaultCluster().Applier().Apply(yaml)
	for _, r := range results {
		fmt.Println(r)
	}
	results = kom.DefaultCluster().Applier().Delete(yaml)
	for _, r := range results {
		fmt.Println(r)
	}
}
func crdExample() {
	err := initCRDWatcher()

	yaml := `apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  # 名字必需与下面的 spec 字段匹配，并且格式为 '<名称的复数形式>.<组名>'
  name: crontabs.stable.example.com
spec:
  # 组名称，用于 REST API: /apis/<组>/<版本>
  group: stable.example.com
  # 列举此 CustomResourceDefinition 所支持的版本
  versions:
    - name: v1
      # 每个版本都可以通过 served 标志来独立启用或禁止
      served: true
      # 其中一个且只有一个版本必需被标记为存储版本
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
  # 可以是 Namespaced 或 clusterInst
  scope: Namespaced
  names:
    # 名称的复数形式，用于 URL：/apis/<组>/<版本>/<名称的复数形式>
    plural: crontabs
    # 名称的单数形式，作为命令行使用时和显示时的别名
    singular: crontab
    # kind 通常是单数形式的驼峰命名（CamelCased）形式。你的资源清单会使用这一形式。
    kind: CronTab
    # shortNames 允许你在命令行使用较短的字符串来匹配资源
    shortNames:
    - ct`
	result := kom.DefaultCluster().Applier().Apply(yaml)
	for _, str := range result {
		fmt.Println(str)
	}

	var crontab *unstructured.Unstructured
	crontab = &unstructured.Unstructured{
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

	err = kom.DefaultCluster().
		CRD("stable.example.com", "v1", "CronTab").
		Name(crontab.GetName()).
		Namespace(crontab.GetNamespace()).
		Create(&crontab).Error
	if err != nil {
		fmt.Printf("CRD Get %v\n", err)
	}
	err = kom.DefaultCluster().
		CRD("stable.example.com", "v1", "CronTab").
		Name(crontab.GetName()).
		Namespace(crontab.GetNamespace()).
		Get(&crontab).Error
	if err != nil {
		fmt.Printf("CRD Get %v\n", err)
	}

	var crontabList []unstructured.Unstructured
	err = kom.DefaultCluster().
		CRD("stable.example.com", "v1", "CronTab").
		Namespace(crontab.GetNamespace()).
		List(&crontabList).Error
	fmt.Printf("CRD List  count %d\n", len(crontabList))
	for _, d := range crontabList {
		fmt.Printf("CRD  List Items foreach %s\n", d.GetName())
	}

	// 定义 Patch 内容
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
	err = kom.DefaultCluster().
		CRD("stable.example.com", "v1", "CronTab").
		Name(crontab.GetName()).
		Namespace(crontab.GetNamespace()).
		Get(&crontab).Error
	if err != nil {
		klog.Errorf("CronTab Get(&item) error :%v", err)
	}
	err = kom.DefaultCluster().
		CRD("stable.example.com", "v1", "CronTab").
		Name(crontab.GetName()).
		Namespace(crontab.GetNamespace()).
		Patch(&crontab, types.MergePatchType, patchData).Error

	if err != nil {
		klog.Errorf("CronTab Patch(&item) error :%v", err)
	}
	// 删除CRD
	err = kom.DefaultCluster().
		CRD("stable.example.com", "v1", "CronTab").
		Name(crontab.GetName()).
		Namespace(crontab.GetNamespace()).
		Delete().Error
	if err != nil {
		klog.Errorf("CronTab Delete(&item) error :%v", err)
	}
}

func builtInExample() {
	yaml := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx
  labels:
    app: nginx
spec:
  replicas: 2
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - name: nginx
        image: nginx:latest
        ports:
        - containerPort: 80
---
apiVersion: v1
kind: Service
metadata:
  name: nginx-service
spec:
  type: ClusterIP
  ports:
  - port: 80
    targetPort: 80
  selector:
    app: nginx
---
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: nginx-ingress
spec:
  rules:
  - host: example.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: nginx-service
            port:
              number: 80
`
	result := kom.DefaultCluster().Applier().Apply(yaml)
	for _, str := range result {
		fmt.Println(str)
	}
	time.Sleep(10 * time.Second)
	var item v1.Deployment
	err := kom.DefaultCluster().
		Resource(&item).
		Namespace("default").
		Name("nginx").
		Get(&item).Error
	if err != nil {
		klog.Errorf("Deployment Get(&item) error :%v", err)
		return
	}
	fmt.Printf("Get Item %s\n", item.Spec.Template.Spec.Containers[0].Image)

	result = kom.DefaultCluster().Applier().Delete(yaml)
	for _, str := range result {
		fmt.Println(str)
	}
	// Create test-deploy
	createItem := v1.Deployment{

		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-deploy",
			Namespace: "default",
			Labels: map[string]string{
				"app": "nginx",
				"m":   "n",
			},
		},
		Spec: v1.DeploymentSpec{
			Replicas: utils.Int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "nginx",
					"m":   "n",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "nginx",
						"m":   "n",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "test",
							Image: "nginx:1.14.2",
						},
					},
				},
			},
		},
	}
	// 创建 test-deploy
	err = kom.DefaultCluster().
		Resource(&createItem).
		Create(&createItem).Error
	if err != nil {
		klog.Errorf("Deployment Create(&item) error :%v", err)
	}

	// 获取 test-deploy
	err = kom.DefaultCluster().
		Resource(&createItem).
		Namespace(createItem.Namespace).
		Name(createItem.Name).
		Get(&createItem).Error
	if err != nil {
		klog.Errorf("Deployment Get(&item) error :%v", err)
	}

	// 更新 test-deploy
	if createItem.Spec.Template.Annotations == nil {
		createItem.Spec.Template.Annotations = map[string]string{}
	}
	createItem.Spec.Template.Annotations["kom.kubernetes.io/restartedAt"] = time.Now().Format(time.RFC3339)
	err = kom.DefaultCluster().
		Resource(&createItem).
		Update(&createItem).Error
	if err != nil {
		klog.Errorf("Deployment Update(&item) error :%v", err)
	}
	// 定义 Patch 内容
	patchData := `{
    "spec": {
        "replicas": 5
    },
    "metadata": {
        "labels": {
            "new-label": "new-value",
            "x": "y"
        }
    }
}`
	// Patch test-deploy
	err = kom.DefaultCluster().
		Resource(&createItem).
		Namespace(createItem.Namespace).
		Name(createItem.Name).
		Get(&createItem).Error
	err = kom.DefaultCluster().
		Resource(&createItem).
		Patch(&createItem, types.MergePatchType, patchData).Error
	if err != nil {
		klog.Errorf("Deployment Patch(&item) error :%v", err)
	}

	// List Deploy
	var items []v1.Deployment
	err = kom.DefaultCluster().
		Resource(&item).
		Namespace("default").
		List(&items).Error
	if err != nil {
		fmt.Printf("List Error %v\n", err)
	}
	fmt.Printf("List Deployment count %d\n", len(items))
	for _, d := range items {
		fmt.Printf("List Deployment  Items foreach %s,%s\n", d.Namespace, d.Spec.Template.Spec.Containers[0].Image)
	}

	// 通过 LabelSelector 获取
	err = kom.DefaultCluster().
		Resource(&item).
		Namespace("default").
		WithLabelSelector("app=nginx").
		List(&items).Error
	if err != nil {
		fmt.Printf("List Error %v\n", err)
	}
	fmt.Printf("List Deployment WithLabelSelector app=nginx count =%v \n", len(items))
	for _, d := range items {
		fmt.Printf("List Deployment WithLabelSelector Items foreach %s,%s\n", d.Namespace, d.Spec.Template.Spec.Containers[0].Image)
	}
	// 通过 LabelSelector 获取
	err = kom.DefaultCluster().
		Resource(&item).
		Namespace("default").
		WithLabelSelector("app=nginx").
		WithLabelSelector("m=n").
		List(&items).Error
	if err != nil {
		fmt.Printf("List Error %v\n", err)
	}
	fmt.Printf("List Deployment WithLabelSelector app=nginx,m=n count =%v \n", len(items))
	for _, d := range items {
		fmt.Printf("List Deployment WithLabelSelector app=nginx,m=n Items foreach %s,%s\n", d.Namespace, d.Spec.Template.Spec.Containers[0].Image)
	}

	// 通过 FieldSelector 获取
	err = kom.DefaultCluster().
		Resource(&item).
		Namespace("default").
		WithFieldSelector("metadata.name=test-deploy").
		List(&items).Error
	if err != nil {
		fmt.Printf("List Error %v\n", err)
	}
	fmt.Printf("List Deployment WithFieldSelector metadata.name=test-deploy count =%v \n", len(items))
	for _, d := range items {
		fmt.Printf("List Deployment WithFieldSelector Items foreach %s,%s\n", d.Namespace, d.Spec.Template.Spec.Containers[0].Image)
	}

	// 删除 test-deploy
	err = kom.DefaultCluster().
		Resource(&createItem).
		Namespace(createItem.Namespace).
		Name(createItem.Name).
		Delete().Error
}

func podLogs() {
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
	time.Sleep(time.Second * 5)
	var stream io.ReadCloser
	err := kom.DefaultCluster().
		Namespace("default").
		Name("random-char-pod").
		ContainerName("container").
		GetLogs(&stream, &corev1.PodLogOptions{}).Error
	if err != nil {
		fmt.Printf("Error getting pod logs:%v\n", err)
	}
	// 逐行读取日志并发送到 Channel
	reader := bufio.NewReader(stream)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			// 处理读取错误，向客户端发送错误消息
			fmt.Printf("Error reading stream: %v\n", err)
			break
		}
		fmt.Println(line)
	}
	result = kom.DefaultCluster().Applier().Delete(yaml)
	for _, str := range result {
		fmt.Println(str)
	}

}
func podCommand() {
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
	time.Sleep(time.Second * 10)

	var execResult []byte
	err := kom.DefaultCluster().Namespace("default").
		Name("random-char-pod").
		ContainerName("container").
		Command("ps", "-ef").
		Execute(&execResult).Error
	if err != nil {
		klog.Errorf("Error executing command: %v", err)
	}
	fmt.Printf("Standard Output:\n%s", execResult)

	result = kom.DefaultCluster().Applier().Delete(yaml)
	for _, str := range result {
		fmt.Println(str)
	}

}
func podFileCommand() {
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

	list, err := kom.DefaultCluster().Namespace("default").
		Name("random-char-pod").
		ContainerName("container").Poder().ListFiles("/etc")
	if err != nil {
		klog.Errorf("Error executing command: %v", err)
	}
	for _, tree := range list {
		fmt.Println(utils.ToJSON(tree))
	}

	file, err := kom.DefaultCluster().Namespace("default").
		Name("random-char-pod").
		ContainerName("container").Poder().DownloadFile("/etc/hosts")
	if err != nil {
		klog.Errorf("Error executing command: %v", err)
	}
	fmt.Println(string(file))

}

func multiCluster() {
	_, err := kom.Clusters().RegisterByPathWithID("/Users/weibh/.kube/orb", "orb")
	if err != nil {
		fmt.Println(err)
		return
	}
	_, err = kom.Clusters().RegisterByPathWithID("/Users/weibh/.kube/docker", "docker")
	if err != nil {
		fmt.Println(err)
		return
	}
	kom.Clusters().Show()
	var pods []corev1.Pod
	err = kom.Cluster("orb").Resource(&corev1.Pod{}).Namespace("kube-system").List(&pods).Error
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("orb pods count=%v\n", len(pods))
	err = kom.Cluster("docker").Resource(&corev1.Pod{}).Namespace("kube-system").List(&pods).Error
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("docker pods count=%v\n", len(pods))
}

func doc() {
	docs := kom.DefaultCluster().Status().Docs()
	docs.ListNames()
}

func newEventList() {
	var list []corev1.Event
	err := kom.DefaultCluster().GVK("events.k8s.io", "v1", "Event").Namespace("default").List(&list).Error
	if err != nil {
		fmt.Printf("events.k8s.io list err %v\n", err)
	}
	if len(list) > 0 {
		json := utils.ToJSON(list[0])
		fmt.Printf("events.k8s.io item json \n %s \n", json)
	} else {
		fmt.Printf("events.k8s.io list count %v\n", len(list))
	}
}
func coreEventList() {
	var list []corev1.Event
	err := kom.DefaultCluster().GVK("", "v1", "Event").Namespace("default").List(&list).Error
	if err != nil {
		fmt.Printf("core events list err %v\n", err)
	}
	if len(list) > 0 {
		json := utils.ToJSON(list[0])
		fmt.Printf("core events item json \n %s \n", json)
	} else {
		fmt.Printf("core events list count %v\n", len(list))
	}
}
func fetchDoc1() {
	kind := "Event"
	group := "events.k8s.io"
	version := "v1"
	tree := kom.DefaultCluster().Status().Docs().FetchByGVK(fmt.Sprintf("%s/%s", group, version), kind)

	// json := utils.ToJSON(tree)
	fmt.Printf("[%s/%s:%s]%s\n", group, version, kind, tree.ID)
}
func fetchDoc2() {
	kind := "Event"
	group := ""
	version := "v1"
	tree := kom.DefaultCluster().Status().Docs().FetchByGVK(fmt.Sprintf("%s/%s", group, version), kind)

	// json := utils.ToJSON(tree)
	fmt.Printf("[%s/%s:%s]%s\n", group, version, kind, tree.ID)
}
