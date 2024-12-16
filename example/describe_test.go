package example

import (
	"testing"

	"github.com/weibaohui/kom/kom"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
)

func TestDescribePod(t *testing.T) {

	var describeResult []byte
	err := kom.DefaultCluster().Resource(&v1.Pod{}).Namespace("default").
		Name("nginx-label-5c5444cd8b-pfrzg").Describe(&describeResult).Error
	if err != nil {
		t.Logf("Error DescribePod : %v", err)
	}
	t.Logf("\n%s\n", describeResult)

}
func TestDescribeDeploy(t *testing.T) {

	var describeResult []byte
	err := kom.DefaultCluster().Resource(&appsv1.Deployment{}).Namespace("default").
		Name("nginx-label").Describe(&describeResult).Error
	if err != nil {
		t.Logf("Error DescribeDeploy : %v", err)
	}
	t.Logf("\n%s\n", describeResult)

}
func TestDescribeVM(t *testing.T) {

	var describeResult []byte
	err := kom.DefaultCluster().CRD("kubevirt.io", "v1", "VirtualMachine").Namespace("default").
		Name("testvm").Describe(&describeResult).Error
	if err != nil {
		t.Logf("Error DescribeVM: %v", err)
	}
	t.Logf("\n%s\n", describeResult)

}
