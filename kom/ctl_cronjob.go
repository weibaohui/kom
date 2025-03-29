package kom

import (
	"k8s.io/apimachinery/pkg/types"
)

type cronJob struct {
	kubectl *Kubectl
}

func (c *cronJob) Pause() error {
	var item interface{}
	patchData := `{"spec":{"suspend":true}}`
	err := c.kubectl.Patch(&item, types.StrategicMergePatchType, patchData).Error
	return err
}
func (c *cronJob) Resume() error {
	var item interface{}
	patchData := `{"spec":{"suspend":false}}`
	err := c.kubectl.Patch(&item, types.StrategicMergePatchType, patchData).Error
	return err
}
