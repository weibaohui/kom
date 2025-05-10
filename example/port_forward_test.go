package example

import (
	"fmt"
	"testing"
	"time"

	"github.com/weibaohui/kom/kom"
	v1 "k8s.io/api/core/v1"
)

func TestPortForward(t *testing.T) {
	stopCh := make(chan struct{})
	go func() {
		time.Sleep(time.Minute)
		stopCh <- struct{}{}
	}()
	err := kom.DefaultCluster().Resource(&v1.Pod{}).
		Namespace("default").
		Name("nginx-deployment-f576985cc-7czqr").
		Ctl().Pod().
		ContainerName("nginx").
		PortForward("20088", "80", stopCh).Error

	if err != nil {
		fmt.Println(err)
	}
}
