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
		t.Logf("get pod linked service error %v\n", err.Error())
		return
	}
	for _, service := range services {
		t.Logf("service name %v\n", service.Name)
	}

}
