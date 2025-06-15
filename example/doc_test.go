package example

import (
	"testing"
	"time"

	"github.com/weibaohui/kom/kom"
	v1 "k8s.io/api/apps/v1"
)

func TestGetDoc(t *testing.T) {
	var docResult []byte
	item := v1.Deployment{}
	field := "spec.replicas"
	field = "spec.template.spec.containers.name"
	field = "spec.template.spec.containers.imagePullPolicy"
	field = "spec.template.spec.containers.livenessProbe.successThreshold"
	err := kom.DefaultCluster().WithCache(time.Second * 50).
		Resource(&item).DocField(field).Doc(&docResult).Error
	if err != nil {
		t.Errorf("Get Deployment Doc [%s]error :%v", field, err)
	}
	t.Logf("Get Deployment Doc [%s] :%s", field, string(docResult))
	err = kom.DefaultCluster().WithCache(time.Second * 50).
		Resource(&item).DocField(field).Doc(&docResult).Error
	if err != nil {
		t.Errorf("Get Deployment Doc [%s]error :%v", field, err)
	}
	t.Logf("Get Deployment Doc [%s] :%s", field, string(docResult))

}
