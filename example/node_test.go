package example

import (
	"testing"

	"github.com/weibaohui/kom/kom"
	"github.com/weibaohui/kom/utils"
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
func TestNodeTaint(t *testing.T) {
	err := kom.DefaultCluster().Resource(&v1.Node{}).
		Name(nodeName).Ctl().Node().Taint("dedicated2=special-user:NoSchedule")
	if err != nil {
		t.Logf("Node Cordon %s error:%v", nodeName, err.Error())
		return
	}
}
func TestNodeUnTaint(t *testing.T) {
	err := kom.DefaultCluster().Resource(&v1.Node{}).
		Name(nodeName).Ctl().Node().UnTaint("dedicated2=special-user:NoSchedule")
	if err != nil {
		t.Logf("Node Cordon %s error:%v", nodeName, err.Error())
		return
	}
}
func TestNodeUnTaint2(t *testing.T) {
	err := kom.DefaultCluster().Resource(&v1.Node{}).
		Name(nodeName).Ctl().Node().UnTaint("dedicated2=s:NoSchedule")
	if err != nil {
		t.Logf("Node Cordon %s error:%v", nodeName, err.Error())
		return
	}
}
func TestNodeUnTaint3(t *testing.T) {
	err := kom.DefaultCluster().Resource(&v1.Node{}).
		Name(nodeName).Ctl().Node().UnTaint("dedicated2:NoSchedule")
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
func TestNodeLabels(t *testing.T) {
	labels, err := kom.DefaultCluster().Resource(&v1.Node{}).Ctl().Node().AllNodeLabels()
	if err != nil {
		t.Logf("Node get labels error:%v", err.Error())
		return
	}
	t.Logf("Node Labels %s", utils.ToJSON(labels))
}

func TestNodeShell(t *testing.T) {
	name := "kind-control-plane"
	ns, pod, container, err := kom.DefaultCluster().Resource(&v1.Node{}).Name(name).Ctl().Node().CreateNodeShell()

	if err != nil {
		t.Logf("Node Shell error:%v", err.Error())
		return
	}
	t.Logf("Node Shell ns=%s podName=%s containerName=%s", ns, pod, container)

}
