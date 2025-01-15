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
