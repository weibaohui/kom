package example

import (
	"bufio"
	"context"
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
	"k8s.io/klog/v2"
)

func Example() {
	doc()
	builtInExample()
	crdExample()
	yamlApplyDelete()
	podLogs()
	multiCluster()
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
	result := kom.DefaultCluster().Applier().Apply(yaml)
	for _, r := range result {
		fmt.Println(r)
	}
	result = kom.DefaultCluster().Applier().Apply(yaml)
	for _, r := range result {
		fmt.Println(r)
	}
	result = kom.DefaultCluster().Applier().Delete(yaml)
	for _, r := range result {
		fmt.Println(r)
	}
}
func crdExample() {
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
	result := kom.Cluster("default").Applier().Apply(yaml)
	for _, str := range result {
		fmt.Println(str)
	}
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

	err := kom.DefaultCluster().
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
	item := v1.Deployment{}
	err := kom.DefaultCluster().
		Resource(&item).
		Namespace("default").
		Name("nginx").
		Get(&item).Error
	if err != nil {
		klog.Errorf("Deployment Get(&item) error :%v", err)
	}
	fmt.Printf("Get Item %s\n", item.Spec.Template.Spec.Containers[0].Image)
	kom.DefaultCluster().Applier().Delete(yaml)
	createItem := v1.Deployment{

		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-deploy",
			Namespace: "default",
		},
		Spec: v1.DeploymentSpec{
			Replicas: utils.Int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "test",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "test",
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
	err = kom.DefaultCluster().
		Resource(&createItem).
		Create(&createItem).Error
	if err != nil {
		klog.Errorf("Deployment Create(&item) error :%v", err)
	}
	err = kom.DefaultCluster().
		Resource(&createItem).
		Namespace(createItem.Namespace).
		Name(createItem.Name).
		Get(&createItem).Error
	if err != nil {
		klog.Errorf("Deployment Get(&item) error :%v", err)
	}
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
            "new-label": "new-value"
        }
    }
}`
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
	err = kom.DefaultCluster().
		Resource(&createItem).
		Namespace(createItem.Namespace).
		Name(createItem.Name).
		Delete().Error

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

	err = kom.DefaultCluster().
		Resource(&item).
		Namespace("default").
		List(&items, metav1.ListOptions{LabelSelector: "app=nginx"}).Error
	if err != nil {
		fmt.Printf("List Error %v\n", err)
	}
	fmt.Printf("List Deployment WithLabelSelector app=nginx count =%v \n", len(items))
	for _, d := range items {
		fmt.Printf("List Deployment WithLabelSelector Items foreach %s,%s\n", d.Namespace, d.Spec.Template.Spec.Containers[0].Image)
	}
}

func podLogs() {
	yaml := `apiVersion: v1
kind: Pod
metadata:
  name: random-char-pod-1
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
    name: container-b
`
	result := kom.DefaultCluster().Applier().Delete(yaml)
	for _, str := range result {
		fmt.Println(str)
	}
	time.Sleep(time.Second * 5)
	result = kom.DefaultCluster().Applier().Apply(yaml)
	for _, str := range result {
		fmt.Println(str)
	}
	time.Sleep(time.Second * 5)
	options := corev1.PodLogOptions{
		Container: "container-b",
	}
	podLogs := kom.DefaultCluster().Poder().
		Namespace("default").
		Name("random-char-pod-1").
		GetLogs("random-char-pod-1", &options)
	logStream, err := podLogs.Stream(context.TODO())
	if err != nil {
		fmt.Println("Error getting pod logs:", err)
	}
	// 逐行读取日志并发送到 Channel
	reader := bufio.NewReader(logStream)
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
