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

	// 获取serviceNames列表
	serviceNames := []string{}
	for _, service := range services {
		serviceNames = append(serviceNames, service.Name)
	}
	t.Logf("serviceNames %v\n", serviceNames)
	// 检查serviceNames列表是否包含nginx-service
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

	// 获取serviceNames列表
	names := []string{}
	for _, endpoint := range endpoints {
		names = append(names, endpoint.Name)
	}
	t.Logf("names %v\n", names)
	// 检查serviceNames列表是否包含nginx-service
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

	// 获取ingressNames列表
	names := []string{}
	for _, ingress := range ingresses {
		names = append(names, ingress.Name)
	}
	t.Logf("names %v\n", names)
	// 检查serviceNames列表是否包含nginx-service
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
	// 检查pvcs列表是否包含my-pvc
	if !slices.Contains(pvcNames, "my-pod-pvc") {
		t.Errorf("my-pod-pvc not found in pvcs")
	}
}
func TestPodLinkPV(t *testing.T) {

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
	pvs, err := kom.DefaultCluster().Resource(&v1.Pod{}).
		Namespace("default").
		Name("my-pod-pvc").Ctl().Pod().LinkedPV()
	if err != nil {
		t.Logf("get pod linked pv error %v\n", err.Error())
		return
	}
	for _, pv := range pvs {
		t.Logf("pv name %v\n", pv.Name)
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

	// json输出
	t.Logf("envs %v\n", utils.ToJSON(envs))
	// 逐行输出
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
  name: my-configmap-multiple-files
data:
  file1.properties: |
    property1=value1
    property2=value2
  file2.properties: |
    propertyA=valueA
    propertyB=valueB
  file3.properties: |
    propertyX=valueX
    propertyY=valueY
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: another-configmap
data:
  config1.properties: |
    another_property1=another_value1
    another_property2=another_value2
  config2.properties: |
    another_propertyA=another_valueA
    another_propertyB=another_valueB
---
apiVersion: v1
kind: Secret
metadata:
  name: my-secret-multiple-files
type: Opaque
stringData:
  file1.secret: |
    secret_property1=secret/value1
    secret_property2=secret/value2
  file2.secret: |
    secret_propertyA=secret/valueA
    secret_propertyB=secret/valueB
  file3.secret: |
    secret_propertyX=secret/valueX
    secret_propertyY=secret/valueY
---
apiVersion: v1
kind: Secret
metadata:
  name: another-secret
type: Opaque
stringData:
  secret1.secret: |
    another_secret_property1=secret/another_value1
    another_secret_property2=secret/another_value2
  secret2.secret: |
    another_secret_propertyA=secret/another_valueA
    another_secret_propertyB=secret/another_valueB
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
    - name: config-volume-1
      mountPath: /config1
      readOnly: true
    - name: config-volume-2
      mountPath: /config2
      readOnly: true
    - name: config-volume-1
      mountPath: /config3from1
      readOnly: true
    - name: config-volume-2
      mountPath: /config4from2
      readOnly: true	  	  
    - name: secret-volume-1
      mountPath: /secret1
      readOnly: true
    - name: secret-volume-2
      mountPath: /secret2
      readOnly: true
    - name: secret-volume-1
      mountPath: /secret3from1
      subPath: secret3from1-subpath
      readOnly: true
    - name: secret-volume-2
      mountPath: /secret4from2
      subPath: secret4from2-subpath
      readOnly: true	  
  volumes:
  - name: config-volume-1
    configMap:
      name: my-configmap-multiple-files
  - name: config-volume-2
    configMap:
      name: another-configmap
  - name: secret-volume-1
    secret:
      secretName: my-secret-multiple-files
  - name: secret-volume-2
    secret:
      secretName: another-secret

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
	if !slices.Contains(secretNames, "my-secret-multiple-files") {
		t.Errorf("my-secret-multiple-files not found in secrets")
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
		// json configmap
		t.Logf("configMap JSON\n %v\n", utils.ToJSON(configMap))
	}
	// 检查configMaps列表是否包含my-configmap
	if !slices.Contains(configMapNames, "my-configmap") {
		t.Errorf("my-configmap not found in configMaps")
	}

}

func TestPodLinkEnvFrom(t *testing.T) {
	yaml := `
# ConfigMap 定义
apiVersion: v1
kind: ConfigMap
metadata:
  name: app-config
data:
  APP_COLOR: blue
  APP_MODE: prod
---
# Secret 定义
apiVersion: v1
kind: Secret
metadata:
  name: db-credentials
type: Opaque
data:
  DB_USERNAME: dXNlcm5hbWU=  # 这是 "username" 的 Base64 编码
  DB_PASSWORD: cGFzc3dvcmQ=  # 这是 "password" 的 Base64 编码
---
# 用于从文件加载环境变量的 ConfigMap
apiVersion: v1
kind: ConfigMap
metadata:
  name: my-env-configmap
data:
  env.list: "FROM_FILE_VARIABLE=This is a variable from file"	
---
# Pod 定义
apiVersion: v1
kind: Pod
metadata:
  name: multi-env-pod-1
spec:
  containers:
  - name: multi-env-container
    image: busybox:1.35.0
    command: ["tail", "-f", "/dev/null"]
    env:
      # 直接定义环境变量
      - name: DIRECT_VARIABLE
        value: "This is a direct variable"
      # 从文件加载环境变量
      - name: FROM_FILE_VARIABLE
        valueFrom:
          configMapKeyRef:
            name: my-env-configmap
            key: env.list
      # 从 ConfigMap 中获取环境变量
      - name: CONFIGMAP_VARIABLE
        valueFrom:
          configMapKeyRef:
            name: app-config
            key: APP_COLOR
      # 从 Secret 中单个获取环境变量
      - name: DB_USERNAME
        valueFrom:
          secretKeyRef:
            name: db-credentials
            key: DB_USERNAME
      - name: DB_PASSWORD
        valueFrom:
          secretKeyRef:
            name: db-credentials
            key: DB_PASSWORD
      # 从 Pod 字段中获取环境变量
      - name: POD_NAME
        valueFrom:
          fieldRef:
            fieldPath: metadata.name
      - name: POD_NAMESPACE
        valueFrom:
          fieldRef:
            fieldPath: metadata.namespace
      # 从 Downward API 中通过 resourceFieldRef 获取环境变量
      - name: CPU_REQUEST
        valueFrom:
          resourceFieldRef:
            containerName: multi-env-container
            resource: requests.cpu
      - name: MEMORY_LIMIT
        valueFrom:
          resourceFieldRef:
            containerName: multi-env-container
            resource: limits.memory
    envFrom:
      # 从文件加载环境变量的另一种方式
      - configMapRef:
          name: my-env-configmap
      # 从 Secret 中一次性获取多个环境变量
      - secretRef:
          name: db-credentials
    resources:
      requests:
        cpu: "200m"
        memory: "128Mi"
      limits:
        cpu: "400m"
        memory: "256Mi"
  volumes:
    - name: envfile-volume
      configMap:
        name: my-env-configmap

`
	kom.DefaultCluster().Applier().Apply(yaml)
	time.Sleep(10 * time.Second)
	envs, err := kom.DefaultCluster().Resource(&v1.Pod{}).
		Namespace("default").
		Name("multi-env-pod-1").Ctl().Pod().LinkedEnvFromPod()
	if err != nil {
		t.Logf("get pod linked env error %v\n", err.Error())
		return
	}
	t.Logf("envs %v\n", utils.ToJSON(envs))
}
func TestPodLinkEnvFrom2(t *testing.T) {
	yaml := `
apiVersion: v1
kind: Pod
metadata:
  name: multi-env-pod-2
spec:
  containers:
  - name: multi-env-container
    image: busybox:1.35.0
    command: ["tail", "-f", "/dev/null"]
    env:
      # 直接定义环境变量
      - name: DIRECT_VARIABLE
        value: "This is a direct variable"
      # 从 Pod 字段中获取环境变量
      - name: POD_NAME
        valueFrom:
          fieldRef:
            fieldPath: metadata.name
      - name: POD_NAMESPACE
        valueFrom:
          fieldRef:
            fieldPath: metadata.namespace
      # 从 Downward API 中通过 resourceFieldRef 获取环境变量
      - name: CPU_REQUEST
        valueFrom:
          resourceFieldRef:
            containerName: multi-env-container
            resource: requests.cpu
      - name: MEMORY_LIMIT
        valueFrom:
          resourceFieldRef:
            containerName: multi-env-container
            resource: limits.memory
    resources:
      requests:
        cpu: "200m"
        memory: "128Mi"
      limits:
        cpu: "400m"
        memory: "256Mi"
`
	kom.DefaultCluster().Applier().Apply(yaml)
	time.Sleep(10 * time.Second)
	envs, err := kom.DefaultCluster().Resource(&v1.Pod{}).
		Namespace("default").
		Name("multi-env-pod-2").Ctl().Pod().LinkedEnvFromPod()
	if err != nil {
		t.Logf("get pod linked env error %v\n", err.Error())
		return
	}
	t.Logf("envs %v\n", utils.ToJSON(envs))
}

func TestPodLinkEnvFrom3(t *testing.T) {
	yaml := `
apiVersion: v1
kind: Pod
metadata:
  name: multi-env-pod-3
spec:
  containers:
  - name: c4
    image: busybox:1.35.0
    command: ["tail", "-f", "/dev/null"]
    env:
      # 直接定义环境变量
      - name: DIRECT_VARIABLE
        value: "This is a direct variable"
      # 从 Pod 字段中获取环境变量
      - name: POD_NAME
        valueFrom:
          fieldRef:
            fieldPath: metadata.name
      - name: POD_NAMESPACE
        valueFrom:
          fieldRef:
            fieldPath: metadata.namespace
      # 从 Downward API 中通过 resourceFieldRef 获取环境变量
      - name: CPU_REQUEST
        valueFrom:
          resourceFieldRef:
            containerName: multi-env-container
            resource: requests.cpu
      - name: MEMORY_LIMIT
        valueFrom:
          resourceFieldRef:
            containerName: multi-env-container
            resource: limits.memory  
  - name: multi-env-container
    image: busybox:1.35.0
    command: ["tail", "-f", "/dev/null"]
    env:
      # 直接定义环境变量
      - name: DIRECT_VARIABLE
        value: "This is a direct variable"
      # 从 Pod 字段中获取环境变量
      - name: POD_NAME
        valueFrom:
          fieldRef:
            fieldPath: metadata.name
      - name: POD_NAMESPACE
        valueFrom:
          fieldRef:
            fieldPath: metadata.namespace
      # 从 Downward API 中通过 resourceFieldRef 获取环境变量
      - name: CPU_REQUEST
        valueFrom:
          resourceFieldRef:
            containerName: c4
            resource: requests.cpu
      - name: MEMORY_LIMIT
        valueFrom:
          resourceFieldRef:
            containerName: c4
            resource: limits.memory
    resources:
      requests:
        cpu: "200m"
        memory: "128Mi"
      limits:
        cpu: "400m"
        memory: "256Mi"
`
	kom.DefaultCluster().Applier().Apply(yaml)
	time.Sleep(10 * time.Second)
	envs, err := kom.DefaultCluster().Resource(&v1.Pod{}).
		Namespace("default").
		Name("multi-env-pod-3").Ctl().Pod().LinkedEnvFromPod()
	if err != nil {
		t.Logf("get pod linked env error %v\n", err.Error())
		return
	}
	t.Logf("envs %v\n", utils.ToJSON(envs))
}
