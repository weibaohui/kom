package example

import (
	"testing"

	"github.com/weibaohui/kom/kom"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestSQL(t *testing.T) {
	sql := "select * from deploy where id!=1 and x.y.z.name!='xxx' and (yyy.uuu.ii.sage>80 and sex=0) and (x='ttt' or yyy='ttt') order by id desc"
	sql = "select * from deploy where (metadata.namespace='kube-system' or metadata.namespace='default' ) and  metadata.name!='hello-28890812-wj2dh' order by id desc limit 2"
	sql = "select * from deploy where (metadata.namespace='kube-system' or metadata.namespace='default' ) and  spec.replicas>=1 order by id desc limit 2"
	sql = "select * from deploy where (metadata.namespace='kube-system' or metadata.namespace='default' ) and  spec.replicas>=2 order by id desc limit 2"
	sql = "select * from deploy where (metadata.namespace='kube-system' or metadata.namespace='default' ) And  metadata.name Like '%dns%' order by id desc limit 2"
	sql = "select * from deploy where (metadata.namespace='kube-system' or metadata.namespace='default' ) and  metadata.name LIKE '%dns' order by id desc limit 2"
	sql = "select * from deploy where (metadata.namespace='kube-system' or metadata.namespace='default' ) and  metadata.name LIKE 'dns%' order by id desc limit 2"
	sql = "select * from deploy where (metadata.namespace='kube-system' or metadata.namespace='default' ) and  spec.replicas<>2 order by id desc limit 2"
	sql = "select * from deploy where (metadata.namespace='kube-system' or metadata.namespace='default' ) and  spec.replicas!=2 order by id desc limit 2"
	sql = "select * from deploy where (metadata.namespace='kube-system' or metadata.namespace='default' ) and  spec.replicas>1 order by id desc limit 2"
	sql = "select * from deploy where (metadata.namespace='kube-system' or metadata.namespace='default' ) and  spec.replicas<1 order by id desc limit 2"
	sql = "select * from deploy where (metadata.namespace='kube-system' or metadata.namespace='default' ) and  spec.replicas<3 order by id desc limit 2"
	sql = "select * from deploy where (metadata.namespace='kube-system' or metadata.namespace='default' ) and  spec.replicas=3 order by id desc limit 2"
	sql = "select * from deploy where (metadata.namespace='kube-system' or metadata.namespace='default' ) and  spec.replicas=2 order by id desc limit 2"
	sql = "select * from deploy where (metadata.namespace='kube-system' or metadata.namespace='default' ) and  spec.replicasin (2,1) order by id desc limit 2"
	sql = "select * from deploy where (metadata.namespace='kube-system' or metadata.namespace='default' ) and  spec.replicas not in (2,3) order by id desc limit 2"
	sql = "select * from deploy where (metadata.namespace='kube-system' or metadata.namespace='default' ) and  spec.replicas between 3 and 4 order by id desc limit 2"
	sql = "select * from deploy where (metadata.namespace='kube-system' or metadata.namespace='default' ) and  spec.replicas not between 3 and 4 order by id desc limit 2"
	sql = "select * from deploy where (metadata.namespace='kube-system' or metadata.namespace='default' ) and  spec.replicas not between 1 and 2 order by id desc limit 2"
	sql = "select * from deploy where (metadata.namespace='kube-system' or metadata.namespace='default' ) and  spec.replicas not between 2 and 3 order by id desc limit 2"
	sql = "select * from deploy where (metadata.namespace='kube-system' or metadata.namespace='default' ) and  metadata.creationTimestamp  between '2024-11-09 00:00:00' and '2024-11-09 23:59:39' order by id desc limit 2"
	sql = "select * from deploy where (metadata.namespace='kube-system' or metadata.namespace='default' ) and  metadata.creationTimestamp not between '2024-11-08' and '2024-11-10' order by id desc limit 2"
	// sql = "select * from deploy where (metadata.namespace='kube-system' or metadata.namespace='default' ) and  metadata.creationTimestamp > '2024-11-08' order by id desc limit 2"
	// sql = "select * from deploy where (metadata.namespace='kube-system' or metadata.namespace='default' ) and  spec.replicas in (2,1) order by id desc limit 2"
	// sql = "select * from virtualmachine where (metadata.namespace='kube-system' or metadata.namespace='default' )  order by id desc"
	// sql = "select * from deploy where (metadata.namespace='kube-system' or metadata.namespace='default' ) and  metadata.creationTimestamp in ('2024-11-08','2024-11-09','2024-11-10') order by id desc limit 2"

	var list []unstructured.Unstructured
	err := kom.DefaultCluster().Sql(sql).List(&list).Error
	if err != nil {
		t.Logf("List error %v", err)
	}
	t.Logf("Count %d", len(list))
	for _, d := range list {
		t.Logf("List Items foreach %s,%s\n", d.GetNamespace(), d.GetName())
	}
}
func TestCRDSQL(t *testing.T) {
	sql := "select * from vm where (metadata.namespace='kube-system' or metadata.namespace='default' )  "

	var list []unstructured.Unstructured
	err := kom.DefaultCluster().Sql(sql).List(&list).Error
	if err != nil {
		t.Logf("List error %v", err)
	}
	t.Logf("Count %d", len(list))
	for _, d := range list {
		t.Logf("List Items foreach %s,%s\n", d.GetNamespace(), d.GetName())
	}
}
func TestPodSQL(t *testing.T) {
	sql := "select * from pod where metadata.namespace='kube-system' or metadata.namespace='default'  order by metadata.name desc  "

	var list []v1.Pod
	err := kom.DefaultCluster().Sql(sql).List(&list).Error
	if err != nil {
		t.Logf("List error %v", err)
	}
	t.Logf("Count %d", len(list))
	for _, d := range list {
		t.Logf("List Items foreach %s,%s\n", d.GetNamespace(), d.GetName())
	}
}
func TestNodeIPSQL(t *testing.T) {
	sql := "select * from node where status.addresses[type=InternalIP].address like '%10%'  "
	// sql = "select * from node where status.addresses.address like '%10%'  "

	var list []v1.Node
	err := kom.DefaultCluster().Sql(sql).List(&list).Error
	if err != nil {
		t.Logf("List error %v", err)
	}
	t.Logf("Count %d", len(list))
	for _, d := range list {
		t.Logf("List Items foreach %s,%s\n", d.GetNamespace(), d.GetName())
	}
}
