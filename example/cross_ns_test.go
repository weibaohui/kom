package example

import (
	"fmt"
	"testing"

	"github.com/weibaohui/kom/kom"
	corev1 "k8s.io/api/core/v1"
)

func TestCrossNs(t *testing.T) {

	yaml := `
apiVersion: v1
kind: Pod
metadata:
  name: cross-ns-pod
  namespace: default
spec:
  containers:
    - name: nginx
      image: nginx
---
apiVersion: v1
kind: Pod
metadata:
  name: cross-ns-pod
  namespace: kube-system
spec:
  containers:
    - name: nginx
      image: nginx
---
apiVersion: v1
kind: Pod
metadata:
  name: cross-ns-pod
  namespace: kube-public
spec:
  containers:
    - name: nginx
      image: nginx
---
`
	kom.DefaultCluster().Applier().Apply(yaml)

	var items []corev1.Pod
	var pod corev1.Pod
	err := kom.DefaultCluster().
		Resource(&pod).
		Namespace("kube-system", "kube-public").
		Where("metadata.name = 'cross-ns-pod'").
		List(&items).Error
	if err != nil {
		t.Errorf("List Error %v\n", err)
	}
	t.Log(fmt.Sprintf("pod count %d", len(items)))
	for _, node := range items {
		t.Logf("pde name %s/%s\n", node.Namespace, node.Name)
	}

}
