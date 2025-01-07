package example

import (
	"slices"
	"testing"
	"time"

	"github.com/weibaohui/kom/kom"
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
		Name("nginx-pod").Ctl().Pod().LinksEndpoints()
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
