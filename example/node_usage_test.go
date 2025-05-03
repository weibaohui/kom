package example

import (
	"fmt"
	"testing"
	"time"

	"github.com/weibaohui/kom/kom"
	"github.com/weibaohui/kom/utils"
	corev1 "k8s.io/api/core/v1"
)

func TestNodePodCount(t *testing.T) {
	nodeName := "kind-control-plane"
	total, used, available := kom.DefaultCluster().Resource(&corev1.Node{}).
		Name(nodeName).Ctl().Node().PodCount()
	t.Logf("Total %d, Used %d, Available %d\n", total, used, available)
}

func TestNodeIPUsage(t *testing.T) {
	nodeName := "kind-control-plane"
	total, used, available := kom.DefaultCluster().Resource(&corev1.Node{}).
		Name(nodeName).Ctl().Node().IPUsage()
	t.Logf("Total %d, Used %d, Available %d\n", total, used, available)
}
func TestNodeResourceUsageTable(t *testing.T) {
	// 打印开始时间
	startTime := time.Now()
	nodeName := "kind-control-plane"
	usage, err := kom.DefaultCluster().Resource(&corev1.Node{}).
		Name(nodeName).WithCache(5 * time.Second).Ctl().Node().ResourceUsageTable()
	if err != nil {
		fmt.Print(err.Error())
		return
	}
	t.Logf("Node Usage %s\n", utils.ToJSON(usage))
	// 打印结束时间
	endTime := time.Now()
	// 计算耗时
	duration := endTime.Sub(startTime)
	t.Logf("Node 统计 耗时  %d  毫秒\n", duration.Milliseconds())

}
