package example

import (
	"testing"

	"github.com/weibaohui/kom/kom"
)

func TestCRDList(t *testing.T) {
	list := kom.DefaultCluster().Status().CRDList()
	for _, crd := range list {
		t.Logf("%s", crd.GetName())
	}
	supported := kom.DefaultCluster().Status().IsGatewayAPISupported()
	t.Logf("gateway api supported: %v", supported)
}
