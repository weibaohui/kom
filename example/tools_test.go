package example

import (
	"testing"

	"github.com/weibaohui/kom/kom"
)

func TestToolCacheClear(t *testing.T) {

	kom.DefaultCluster().Tools().ClearCache()

}
