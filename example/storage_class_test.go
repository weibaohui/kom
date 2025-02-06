package example

import (
	"testing"

	"github.com/weibaohui/kom/kom"
	v1 "k8s.io/api/storage/v1"
)

func TestStorageClassPVCCount(t *testing.T) {
	scName := "hostpath"
	count, err := kom.DefaultCluster().Resource(&v1.StorageClass{}).Name(scName).
		Ctl().StorageClass().PVCCount()
	if err != nil {
		t.Logf("get stroage class pvc count error %v", err)
	}

	t.Logf("pvc count %d\n", count)
}
