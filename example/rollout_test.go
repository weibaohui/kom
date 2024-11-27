package example

import (
	"testing"

	"github.com/weibaohui/kom/kom"
	v1 "k8s.io/api/apps/v1"
)

func TestRollout_Undo(t *testing.T) {
	result, err := kom.DefaultCluster().Resource(&v1.Deployment{}).
		Namespace("default").Name("random-number-deployment").Ctl().Rollout().Undo(5)
	if err != nil {
		t.Log(err)
		return
	}
	t.Logf("random-number-deployment undo %s", result)
}
func TestRollout_History(t *testing.T) {
	result, err := kom.DefaultCluster().Resource(&v1.Deployment{}).
		Namespace("default").Name("random-number-deployment").Ctl().Rollout().History()
	if err != nil {
		t.Error(err)
		return
	}
	t.Logf("random-number-deployment undo %s", result)
}
