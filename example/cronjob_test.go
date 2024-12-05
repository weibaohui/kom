package example

import (
	"testing"

	"github.com/weibaohui/kom/kom"
	v1 "k8s.io/api/batch/v1"
)

var cronJobName = "hello"

func TestCronJob_Pause(t *testing.T) {
	err := kom.DefaultCluster().Resource(&v1.CronJob{}).
		Namespace("default").Name(cronJobName).Ctl().CronJob().Pause()
	if err != nil {
		t.Log(err)
		return
	}
	t.Logf(" CronJob %s Pause", cronJobName)
}
func TestCronJob_Resume(t *testing.T) {
	err := kom.DefaultCluster().Resource(&v1.CronJob{}).
		Namespace("default").Name(cronJobName).Ctl().CronJob().Resume()
	if err != nil {
		t.Log(err)
		return
	}
	t.Logf(" CronJob %s Resume", cronJobName)
}
