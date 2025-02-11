package example

import (
	"testing"

	"github.com/weibaohui/kom/kom"
	v1 "k8s.io/api/storage/v1"
)

func TestStorageClassPVCCount(t *testing.T) {
	scName := "standard"
	count, err := kom.DefaultCluster().Resource(&v1.StorageClass{}).Name(scName).
		Ctl().StorageClass().PVCCount()
	if err != nil {
		t.Logf("get stroage class pvc count error %v", err)
	}

	t.Logf("pvc count %d\n", count)
}
func TestStorageClassPVCount(t *testing.T) {
	scName := "standard"
	count, err := kom.DefaultCluster().Resource(&v1.StorageClass{}).Name(scName).
		Ctl().StorageClass().PVCount()
	if err != nil {
		t.Logf("get stroage class pv count error %v", err)
	}

	t.Logf("pv count %d\n", count)
}
func TestStorageClassSetDefault(t *testing.T) {
	scName := "hostpath"
	err := kom.DefaultCluster().Resource(&v1.StorageClass{}).Name(scName).
		Ctl().StorageClass().SetDefault()
	if err != nil {
		t.Logf("set stroage class default error %v", err)
	}
}
