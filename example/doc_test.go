package example

import (
	"testing"

	"github.com/weibaohui/kom/kom"
	"github.com/weibaohui/kom/utils"
)

func TestDoc(t *testing.T) {
	docs := kom.DefaultCluster().Status().Docs()
	pc := docs.FetchByGVK("v1", "Pod")
	t.Log(utils.ToJSON(pc))
}
