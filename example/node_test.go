package example

import (
	"testing"

	"github.com/weibaohui/kom/kom"
	v1 "k8s.io/api/core/v1"
)

var nodeName = "kind-control-plane"

func TestNodeCordon(t *testing.T) {
	err := kom.DefaultCluster().Resource(&v1.Node{}).
		Name(nodeName).Ctl().Node().Cordon()
	if err != nil {
		t.Logf("Node Cordon %s error:%v", nodeName, err.Error())
		return
	}
}

func TestNodeUnCordon(t *testing.T) {
	err := kom.DefaultCluster().Resource(&v1.Node{}).
		Name(nodeName).Ctl().Node().UnCordon()
	if err != nil {
		t.Logf("Node UnCordon %s error:%v", nodeName, err.Error())
		return
	}
}
func TestNodeDrain(t *testing.T) {
	err := kom.DefaultCluster().Resource(&v1.Node{}).
		Name(nodeName).Ctl().Node().Drain()
	if err != nil {
		t.Logf("Node Drain %s error:%v", nodeName, err.Error())
		return
	}
}
