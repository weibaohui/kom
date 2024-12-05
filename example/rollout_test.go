package example

import (
	"testing"

	"github.com/weibaohui/kom/kom"
	"github.com/weibaohui/kom/utils"
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
---
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: random-number-sts
  namespace: default
spec:
  replicas: 3
  selector:
    matchLabels:
      app: random-number
  serviceName: random-number-service
  template:
    metadata:
      labels:
        app: random-number
    spec:
      containers:
      - command:
        - sh
        - -c
        - |
          while true; do
            echo "S1111+$RANDOM";
            sleep 5;
          done;
        image: busybox
        imagePullPolicy: IfNotPresent
        name: random-generator
`
	kom.DefaultCluster().Applier().Apply(yaml)
}

var deployName = "random-number-deployment"
var dsName = "random-d-generator"
var stsName = "random-number-sts"

func TestRollout_Deploy_Undo(t *testing.T) {
	result, err := kom.DefaultCluster().Resource(&v1.Deployment{}).
		Namespace("default").Name(deployName).Ctl().Rollout().Undo()
	if err != nil {
		t.Log(err)
		return
	}
	t.Logf("%s rollout undo %s", deployName, result)
}
func TestRollout_Deploy_History(t *testing.T) {
	result, err := kom.DefaultCluster().Resource(&v1.Deployment{}).
		Namespace("default").Name(deployName).Ctl().Rollout().History()
	if err != nil {
		t.Log(err)
		return
	}
	t.Logf("%s rollout history %s", deployName, utils.ToJSON(result))
}
func TestRollout_Deploy_Status(t *testing.T) {
	result, err := kom.DefaultCluster().Resource(&v1.Deployment{}).
		Namespace("default").Name(deployName).Ctl().Rollout().Status()
	if err != nil {
		t.Log(err)
		return
	}
	t.Logf("%s rollout status %s", deployName, result)
}

func TestRollout_Ds_Undo(t *testing.T) {
	result, err := kom.DefaultCluster().Resource(&v1.DaemonSet{}).
		Namespace("default").Name(dsName).Ctl().Rollout().Undo()
	if err != nil {
		t.Log(err)
		return
	}
	t.Logf("%s rollout undo %s", dsName, result)
}
func TestRollout_Ds_History(t *testing.T) {
	result, err := kom.DefaultCluster().Resource(&v1.DaemonSet{}).
		Namespace("default").Name(dsName).Ctl().Rollout().History()
	if err != nil {
		t.Log(err)
		return
	}
	t.Logf("%s rollout history%s", dsName, utils.ToJSON(result))
}
func TestRollout_Ds_Status(t *testing.T) {
	result, err := kom.DefaultCluster().Resource(&v1.DaemonSet{}).
		Namespace("default").Name(dsName).Ctl().Rollout().Status()
	if err != nil {
		t.Log(err)
		return
	}
	t.Logf("%s rollout status%s", dsName, result)
}

func TestRollout_Sts_Undo(t *testing.T) {
	result, err := kom.DefaultCluster().Resource(&v1.StatefulSet{}).
		Namespace("default").Name(stsName).Ctl().Rollout().Undo()
	if err != nil {
		t.Log(err)
		return
	}
	t.Logf("%s rollout undo %s", stsName, result)
}
func TestRollout_Sts_History(t *testing.T) {
	result, err := kom.DefaultCluster().Resource(&v1.StatefulSet{}).
		Namespace("default").Name(stsName).Ctl().Rollout().History()
	if err != nil {
		t.Log(err)
		return
	}
	t.Logf("%s rollout history%s", stsName, utils.ToJSON(result))
}
func TestRollout_Sts_Status(t *testing.T) {
	result, err := kom.DefaultCluster().Resource(&v1.StatefulSet{}).
		Namespace("default").Name(stsName).Ctl().Rollout().Status()
	if err != nil {
		t.Log(err)
		return
	}
	t.Logf("%s rollout status%s", stsName, result)
}
