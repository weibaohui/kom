package example

import (
	"testing"

	"github.com/weibaohui/kom/kom"
	v1 "k8s.io/api/apps/v1"
)

func TestDeployStop(t *testing.T) {
	yaml := `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  labels:
    app: nginx
spec:
  replicas: 3
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
        image: nginx
`
	result := kom.DefaultCluster().Applier().Apply(yaml)
	for _, r := range result {
		t.Log(r)
	}
	var deploy v1.Deployment
	err := kom.DefaultCluster().Resource(&deploy).
		Namespace("default").
		Name("nginx-deployment").
		Ctl().Deployment().Stop()
	if err != nil {
		t.Log(err)
	}
}

func TestDeployRestore(t *testing.T) {
	var deploy v1.Deployment
	err := kom.DefaultCluster().Resource(&deploy).
		Namespace("default").
		Name("nginx-deployment").
		Ctl().Deployment().Restore()
	if err != nil {
		t.Log(err)
	}
}
func TestDaemonSetRestore(t *testing.T) {
	var ds v1.DaemonSet
	err := kom.DefaultCluster().Resource(&ds).
		Namespace("default").
		Name("nginx-daemonset-test").
		Ctl().DaemonSet().Restore()
	if err != nil {
		t.Log(err)
	}
}
func TestDaemonSetStop(t *testing.T) {
	yaml := `apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: nginx-daemonset-test
  namespace: default
spec:
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - image: nginx:alpine
        name: nginx
       `
	kom.DefaultCluster().Applier().Apply(yaml)
	var ds v1.DaemonSet
	err := kom.DefaultCluster().Resource(&ds).
		Namespace("default").
		Name("nginx-daemonset-test").
		Ctl().DaemonSet().Stop()
	if err != nil {
		t.Log(err)
	}
}
