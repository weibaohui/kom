package example

import (
	"testing"

	"github.com/weibaohui/kom/kom"
	v1 "k8s.io/api/core/v1"
)

func TestPodLink(t *testing.T) {
	services, err := kom.DefaultCluster().Resource(&v1.Pod{}).
		Namespace("k8m").
		Name("k8m-6b56d66cbf-cf222").Ctl().Pod().LinkedService()
	if err != nil {
		t.Errorf("get pod linked service error %v\n", err.Error())
	}
	t.Logf("pod linked service %v\n", services)
}
