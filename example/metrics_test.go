package example

import (
	"testing"

	"github.com/weibaohui/kom/kom"
	"github.com/weibaohui/kom/utils"
	v1 "k8s.io/api/core/v1"
)

func TestPrintTopPodList(t *testing.T) {

	result, err := kom.DefaultCluster().
		Resource(&v1.Pod{}).Namespace("default").
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
        t.Fail() // 添加失败处理
    }

    t.Logf("\n%s\n", utils.ToJSON(result))
    // 添加断言验证结果
    if err == nil {
        if len(result) != 1 {
            t.Errorf("预期结果应该只包含1个Pod，但获得了%d个", len(result))
        }
        if result[0].Name != "cpu-memory-fluctuation-advanced" {
            t.Errorf("预期Pod名称为cpu-memory-fluctuation-advanced，但获得了%s", result[0].Name)
        }
    }
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
