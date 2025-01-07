package example

import (
	"slices"
	"testing"
	"time"

	"github.com/weibaohui/kom/kom"
	"github.com/weibaohui/kom/utils"
	v1 "k8s.io/api/core/v1"
)

var podLinkYaml = `
apiVersion: v1
kind: Pod
metadata:
  name: nginx-pod
  labels:
    app: nginx
spec:
  containers:
  - name: nginx
    image: nginx:1.23.3
    ports:
    - containerPort: 80

---
apiVersion: v1
kind: Service
metadata:
  name: nginx-service
spec:
  selector:
    app: nginx
  ports:
  - protocol: TCP
    port: 80
    targetPort: 80
  type: ClusterIP
---
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: nginx-ingress
  annotations:
    nginx.ingress.kubernetes.io/rewrite-target: /
spec:
  rules:
  - host: yourdomain.com
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

func TestPodLinkService(t *testing.T) {

	kom.DefaultCluster().Applier().Apply(podLinkYaml)
	time.Sleep(10 * time.Second)
	services, err := kom.DefaultCluster().Resource(&v1.Pod{}).
		Namespace("default").
		Name("nginx-pod").Ctl().Pod().LinkedService()
	if err != nil {
		t.Logf("get pod linked service error %v\n", err.Error())
		return
	}
	for _, service := range services {
		t.Logf("service name %v\n", service.Name)
	}

	//获取serviceNames列表
	serviceNames := []string{}
	for _, service := range services {
		serviceNames = append(serviceNames, service.Name)
	}
	t.Logf("serviceNames %v\n", serviceNames)
	//检查serviceNames列表是否包含nginx-service
	if !slices.Contains(serviceNames, "nginx-service") {
		t.Errorf("nginx-service not found in serviceNames")
	}
}
func TestPodLinkEndpoints(t *testing.T) {

	kom.DefaultCluster().Applier().Apply(podLinkYaml)
	time.Sleep(10 * time.Second)
	endpoints, err := kom.DefaultCluster().Resource(&v1.Pod{}).
		Namespace("default").
		Name("nginx-pod").Ctl().Pod().LinkedEndpoints()
	if err != nil {
		t.Logf("get pod linked endpoints error %v\n", err.Error())
		return
	}
	for _, endpoint := range endpoints {
		t.Logf("endpoint name %v\n", endpoint.Name)
	}

	//获取serviceNames列表
	names := []string{}
	for _, endpoint := range endpoints {
		names = append(names, endpoint.Name)
	}
	t.Logf("names %v\n", names)
	//检查serviceNames列表是否包含nginx-service
	if !slices.Contains(names, "nginx-service") {
		t.Errorf("nginx-service not found in endpoints")
	}
}
func TestPodLinkIngress(t *testing.T) {
	kom.DefaultCluster().Applier().Apply(podLinkYaml)
	time.Sleep(10 * time.Second)
	ingresses, err := kom.DefaultCluster().Resource(&v1.Pod{}).
		Namespace("default").
		Name("nginx-pod").Ctl().Pod().LinkedIngress()
	if err != nil {
		t.Logf("get pod linked ingress error %v\n", err.Error())
		return
	}
	for _, ingress := range ingresses {
		t.Logf("ingress name %v\n", ingress.Name)
	}

	//获取ingressNames列表
	names := []string{}
	for _, ingress := range ingresses {
		names = append(names, ingress.Name)
	}
	t.Logf("names %v\n", names)
	//检查serviceNames列表是否包含nginx-service
	if !slices.Contains(names, "nginx-ingress") {
		t.Errorf("nginx-ingress not found in ingresses")
	}
}
func TestPodLinkPVC(t *testing.T) {

	yaml := `
	 # 定义 PersistentVolumeClaim
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: my-pod-pvc
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 1Gi
---
# 定义 Pod
apiVersion: v1
kind: Pod
metadata:
  name: my-pod-pvc
spec:
  containers:
  - name: my-container
    image: nginx:1.23.3
    ports:
    - containerPort: 80
    volumeMounts:
    - name: my-volume
      mountPath: /data
  volumes:
  - name: my-volume
    persistentVolumeClaim:
      claimName: my-pod-pvc
`
	kom.DefaultCluster().Applier().Apply(yaml)
	time.Sleep(10 * time.Second)
	pvcs, err := kom.DefaultCluster().Resource(&v1.Pod{}).
		Namespace("default").
		Name("my-pod-pvc").Ctl().Pod().LinkedPVC()
	if err != nil {
		t.Logf("get pod linked pvc error %v\n", err.Error())
		return
	}
	for _, pvc := range pvcs {
		t.Logf("pvc name %v\n", pvc.Name)
	}
	pvcNames := []string{}
	for _, pvc := range pvcs {
		pvcNames = append(pvcNames, pvc.Name)
	}
	//检查pvcs列表是否包含my-pvc
	if !slices.Contains(pvcNames, "my-pod-pvc") {
		t.Errorf("my-pod-pvc not found in pvcs")
	}
}

func TestPodLinkEnv(t *testing.T) {
	kom.DefaultCluster().Applier().Apply(podLinkYaml)
	time.Sleep(10 * time.Second)
	envs, err := kom.DefaultCluster().Resource(&v1.Pod{}).
		Namespace("default").
		Name("nginx-pod").Ctl().Pod().LinkedEnv()
	if err != nil {
		t.Logf("get pod linked env error %v\n", err.Error())
		return
	}

	//json输出
	t.Logf("envs %v\n", utils.ToJSON(envs))
	//逐行输出
	for _, env := range envs {
		t.Logf("env %s %s=%s\n", env.ContainerName, env.EnvName, env.EnvValue)
	}
}

// secret
func TestPodLinkSecret(t *testing.T) {
	yaml := `
	apiVersion: v1
kind: ConfigMap
metadata:
  name: my-configmap
data:
  config.properties: |
    property1=value1
    property2=value2
---
apiVersion: v1
kind: Secret
metadata:
  name: my-secret
type: Opaque
stringData:
  secret.properties: |
    secret_property1=secret/value1
    secret_property2=secret/value2
---
apiVersion: v1
kind: Pod
metadata:
  name: my-pod-secret-configmap
spec:
  containers:
  - name: my-container
    image: nginx:1.23.3
    ports:
    - containerPort: 80
    volumeMounts:
    - name: config-volume
      mountPath: /config
      readOnly: true
    - name: secret-volume
      mountPath: /secret
      readOnly: true
  volumes:
  - name: config-volume
    configMap:
      name: my-configmap
  - name: secret-volume
    secret:
      secretName: my-secret
`
	kom.DefaultCluster().Applier().Apply(yaml)
	time.Sleep(10 * time.Second)
	secrets, err := kom.DefaultCluster().Resource(&v1.Pod{}).
		Namespace("default").
		Name("my-pod-secret-configmap").Ctl().Pod().LinkedSecret()
	if err != nil {
		t.Logf("get pod linked secret error %v\n", err.Error())
		return
	}

	secretNames := []string{}
	for _, secret := range secrets {
		secretNames = append(secretNames, secret.Name)
		t.Logf("secretMounts %s %v\n", secret.Name, secret.Annotations["secretMounts"])
	}
	// 检查secrets列表是否包含my-secret
	if !slices.Contains(secretNames, "my-secret") {
		t.Errorf("my-secret not found in secrets")
	}

}

// secret
func TestPodLinkConfigMap(t *testing.T) {
	yaml := `
	apiVersion: v1
kind: ConfigMap
metadata:
  name: my-configmap
data:
  config.properties: |
    property1=value1
    property2=value2
---
apiVersion: v1
kind: Secret
metadata:
  name: my-secret
type: Opaque
stringData:
  secret.properties: |
    secret_property1=secret/value1
    secret_property2=secret/value2
---
apiVersion: v1
kind: Pod
metadata:
  name: my-pod-secret-configmap
spec:
  containers:
  - name: my-container
    image: nginx:1.23.3
    ports:
    - containerPort: 80
    volumeMounts:
    - name: config-volume
      mountPath: /config
      readOnly: true
    - name: secret-volume
      mountPath: /secret
      readOnly: true
  volumes:
  - name: config-volume
    configMap:
      name: my-configmap
  - name: secret-volume
    secret:
      secretName: my-secret
`
	kom.DefaultCluster().Applier().Apply(yaml)
	time.Sleep(10 * time.Second)
	configMaps, err := kom.DefaultCluster().Resource(&v1.Pod{}).
		Namespace("default").
		Name("my-pod-secret-configmap").Ctl().Pod().LinkedConfigMap()
	if err != nil {
		t.Logf("get pod linked configMap error %v\n", err.Error())
		return
	}

	configMapNames := []string{}
	for _, configMap := range configMaps {
		configMapNames = append(configMapNames, configMap.Name)
		t.Logf("configMapMounts %s %v\n", configMap.Name, configMap.Annotations["configMapMounts"])
		//json configmap
		t.Logf("configMap JSON\n %v\n", utils.ToJSON(configMap))
	}
	// 检查configMaps列表是否包含my-configmap
	if !slices.Contains(configMapNames, "my-configmap") {
		t.Errorf("my-configmap not found in configMaps")
	}

}
