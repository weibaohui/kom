package example

import (
	"testing"

	"github.com/weibaohui/kom/kom"
	v1 "k8s.io/api/networking/v1"
)

func TestIngressClassSetDefault(t *testing.T) {
	scName := "nginx-2"
	err := kom.DefaultCluster().Resource(&v1.IngressClass{}).Name(scName).
		Ctl().IngressClass().SetDefault()
	if err != nil {
		t.Logf("set ingress class default error %v", err)
	}
}
