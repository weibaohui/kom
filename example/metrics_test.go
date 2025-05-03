package example

import (
	"testing"

	"github.com/weibaohui/kom/kom"
	"github.com/weibaohui/kom/utils"
	v1 "k8s.io/api/core/v1"
)

func TestPrintTopPodList(t *testing.T) {

	result, err := kom.DefaultCluster().
		Resource(&v1.Pod{}).Namespace("*").
		// Name("cpu-memory-fluctuation-advanced").
		Ctl().Pod().Top()

	if err != nil {
		t.Log(err.Error())
	}

	t.Logf("\n%s\n", utils.ToJSON(result))

}
func TestPrintTopSinglePod(t *testing.T) {

	result, err := kom.DefaultCluster().
		Resource(&v1.Pod{}).Namespace("default").
		Name("cpu-memory-fluctuation-advanced").
		Ctl().Pod().Top()

	if err != nil {
		t.Log(err.Error())
	}

	t.Logf("\n%s\n", utils.ToJSON(result))

}
func TestPrintPodMetrics(t *testing.T) {

	result, err := kom.DefaultCluster().
		Resource(&v1.Pod{}).Namespace("kube-system").
		Name("kube-apiserver-kind-cluster-control-plane").
		Ctl().Pod().ResourceUsageTable()

	if err != nil {
		t.Log(err.Error())
	}

	t.Logf("\n%s\n", utils.ToJSON(result))

}

func TestPrintNodeMetrics(t *testing.T) {

	result, err := kom.DefaultCluster().
		Resource(&v1.Node{}).
		Name("kind-cluster-control-plane").
		Ctl().Node().ResourceUsageTable()

	if err != nil {
		t.Log(err.Error())
	}

	t.Logf("\n%s\n", utils.ToJSON(result))

}

func TestPrintTopNodeList(t *testing.T) {

	result, err := kom.DefaultCluster().
		Resource(&v1.Node{}).
		Ctl().Node().Top()

	if err != nil {
		t.Log(err.Error())
	}

	t.Logf("\n%s\n", utils.ToJSON(result))

}
func TestPrintTopSingleNode(t *testing.T) {

	result, err := kom.DefaultCluster().
		Resource(&v1.Node{}).
		Name("kind-cluster-control-plane").
		Ctl().Node().Top()

	if err != nil {
		t.Log(err.Error())
	}

	t.Logf("\n%s\n", utils.ToJSON(result))

}
