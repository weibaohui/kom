package example

import (
	"testing"
	"time"

	"github.com/weibaohui/kom/kom"
	v1 "k8s.io/api/core/v1"
)

func TestDeployManagedRsMuiltiple(t *testing.T) {

	rs, err := kom.DefaultCluster().Namespace("k8m").
		Name("k8m").
		Ctl().Deployment().
		ManagedLatestReplicaSet()
	if err != nil {
		t.Logf("ManagedLatestReplicaSet error: %v", err)
	}
	if rs == nil {
		t.Logf("ManagedLatestReplicaSet error: %v", err)
		return
	}
	t.Logf("ManagedLatestReplicaSet: %v", rs.Name)
}
func TestDeployManagedRs(t *testing.T) {
	yaml := `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: managed-pods
  labels:
    app: nginx
spec:
  replicas: 2
  selector:
    matchLabels:
      app: nginx
  revisionHistoryLimit: 0
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - name: nginx
        image: nginx:alpine
`
	kom.DefaultCluster().Applier().Apply(yaml)

	rs, err := kom.DefaultCluster().Namespace("default").
		Name("managed-pods").
		Ctl().Deployment().
		ManagedLatestReplicaSet()
	if err != nil {
		t.Logf("ManagedLatestReplicaSet error: %v", err)
	}
	if rs == nil {
		t.Logf("ManagedLatestReplicaSet error: %v", err)
		return
	}
	t.Logf("ManagedLatestReplicaSet: %v", rs.Name)
}
func TestDeployManagedPods(t *testing.T) {
	yaml := `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: managed-pods
  labels:
    app: nginx
spec:
  replicas: 2
  selector:
    matchLabels:
      app: nginx
  revisionHistoryLimit: 0
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - name: nginx
        image: nginx:alpine
`
	kom.DefaultCluster().Applier().Apply(yaml)

	list, err := kom.DefaultCluster().Namespace("default").
		Name("managed-pods").
		Ctl().Deployment().
		ManagedPods()
	if err != nil {
		t.Logf("ManagedPods error: %v", err)
	}
	t.Logf("ManagedPods Count %d", len(list))
	if len(list) == 0 {
		t.Logf("ManagedPods error: %v", err)
		return
	}
	for _, pod := range list {
		t.Logf("ManagedPods: %v", pod.Name)
	}
}
func TestDeployManagedPodsOpenKruise(t *testing.T) {
	time.Sleep(10 * time.Second)
	list, err := kom.DefaultCluster().
		Resource(&v1.Pod{}).
		Namespace("default").
		Name("nginx-statefulset").
		Ctl().StatefulSet().
		ManagedPods()
	if err != nil {
		t.Logf("ManagedPods error: %v", err)
	}
	t.Logf("ManagedPods Count %d", len(list))
	if len(list) == 0 {
		t.Logf("ManagedPods error: %v", err)
		return
	}
	for _, pod := range list {
		t.Logf("ManagedPods: %v", pod.Name)
	}
}
func TestDeployManagedPod(t *testing.T) {
	yaml := `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: managed-pods
  labels:
    app: nginx
spec:
  replicas: 2
  selector:
    matchLabels:
      app: nginx
  revisionHistoryLimit: 0
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - name: nginx
        image: nginx:alpine
`
	kom.DefaultCluster().Applier().Apply(yaml)

	item, err := kom.DefaultCluster().Namespace("default").
		Name("managed-pods").
		Ctl().Deployment().
		ManagedPod()
	if err != nil {
		t.Logf("ManagedPod error: %v", err)
	}
	if item == nil {
		t.Logf("ManagedPod error: %v", err)
		return
	}
	t.Logf("ManagedPod: %v", item.Name)
}

func TestStatefulSetManagedPod(t *testing.T) {

	item, err := kom.DefaultCluster().Namespace("default").
		Name("svc-not-exists-web").
		Ctl().StatefulSet().
		ManagedPod()
	if err != nil {
		t.Logf("ManagedPod error: %v", err)
	}
	if item == nil {
		t.Logf("ManagedPod error: %v", err)
		return
	}
	if err != nil {
		t.Logf("ManagedPod error: %v", err)
	}
	t.Logf("ManagedPod: %v", item.Name)
}
func TestDaemonSetManagedPod(t *testing.T) {
	item, err := kom.DefaultCluster().Namespace("default").
		Name("nginx-daemonset").
		Ctl().DaemonSet().
		ManagedPod()
	if err != nil {
		t.Logf("ManagedPod error: %v", err)
	}
	if item == nil {
		t.Logf("ManagedPod error: %v", err)
		return
	}
	t.Logf("ManagedPod: %v", item.Name)
}
func TestReplicaSetManagedPod(t *testing.T) {
	item, err := kom.DefaultCluster().Namespace("default").
		Name("managed-pods-9c65654f4").
		Ctl().ReplicaSet().
		ManagedPod()
	if err != nil {
		t.Logf("ManagedPod error: %v", err)
	}
	if item == nil {
		t.Logf("ManagedPod error: %v", err)
		return
	}
	t.Logf("ManagedPod: %v", item.Name)
}
