package example

import (
	"testing"

	"github.com/weibaohui/kom/kom"
	v1 "k8s.io/api/apps/v1"
)

func prepareDeploy() {
	var yaml = `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: random-number-deployment
  namespace: default
spec:
  selector:
    matchLabels:
      app: random-number
  template:
    metadata:
      labels:
        app: random-number
    spec:
      containers:
      - args:
        - |
          while true; do
            echo "A+$RANDOM";
            sleep 2;
          done;
        command:
        - /bin/sh
        - -c
        image: busybox
        imagePullPolicy: Always
        name: random-number
        resources:
          limits:
            cpu: 50m
            memory: 16Mi
          requests:
            cpu: 10m
            memory: 8Mi
---
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: random-d-generator
  namespace: default
spec:
  selector:
    matchLabels:
      app: random-d-generator
  template:
    metadata:
      labels:
        app: random-d-generator
    spec:
      containers:
      - command:
        - /bin/sh
        - -c
        - |
          while true; do
            echo "D$(head /dev/urandom | tr -dc '0-9' | head -c 5)";
            sleep 2;
          done
        image: busybox
        imagePullPolicy: Always
        name: busybox
        resources:
          limits:
            cpu: 50m
            memory: 16Mi
          requests:
            cpu: 10m
            memory: 8Mi
`
	kom.DefaultCluster().Applier().Apply(yaml)
}

var deployName = "random-number-deployment"
var dsName = "random-d-generator"

func TestRollout_Deploy_Undo(t *testing.T) {
	result, err := kom.DefaultCluster().Resource(&v1.Deployment{}).
		Namespace("default").Name(deployName).Ctl().Rollout().Undo()
	if err != nil {
		t.Log(err)
		return
	}
	t.Logf("%s undo %s", deployName, result)
}
func TestRollout_Deploy_History(t *testing.T) {
	result, err := kom.DefaultCluster().Resource(&v1.Deployment{}).
		Namespace("default").Name(deployName).Ctl().Rollout().History()
	if err != nil {
		t.Log(err)
		return
	}
	t.Logf("%s rollout history %s", deployName, result)
}
func TestRollout_Ds_Undo(t *testing.T) {
	result, err := kom.DefaultCluster().Resource(&v1.DaemonSet{}).
		Namespace("default").Name(dsName).Ctl().Rollout().Undo()
	if err != nil {
		t.Log(err)
		return
	}
	t.Logf("%s undo %s", dsName, result)
}
func TestRollout_Ds_History(t *testing.T) {
	result, err := kom.DefaultCluster().Resource(&v1.DaemonSet{}).
		Namespace("default").Name(dsName).Ctl().Rollout().History()
	if err != nil {
		t.Log(err)
		return
	}
	t.Logf("%s rollout %s", dsName, result)
}
func TestRollout_Ds_Status(t *testing.T) {
	result, err := kom.DefaultCluster().Resource(&v1.DaemonSet{}).
		Namespace("default").Name(dsName).Ctl().Rollout().Status()
	if err != nil {
		t.Log(err)
		return
	}
	t.Logf("%s rollout %s", dsName, result)
}
