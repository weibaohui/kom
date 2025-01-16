package kom

import (
	"fmt"

	"k8s.io/apimachinery/pkg/types"
)

type daemonSet struct {
	kubectl *Kubectl
}

func (r *daemonSet) Restart() error {
	return r.kubectl.Ctl().Rollout().Restart()
}
func (r *daemonSet) Stop() error {

	patchData := `{
  "spec": {
    "template": {
      "spec": {
        "nodeSelector": {
          "kubernetes.io/hostname": "non-existent-node"
        }
      }
    }
  }
}`
	var item interface{}
	err := r.kubectl.Patch(&item, types.MergePatchType, patchData).Error

	if err != nil {
		return fmt.Errorf("stop %s/%s error %v", r.kubectl.Statement.Namespace, r.kubectl.Statement.Name, err)
	}
	return nil
}
func (r *daemonSet) Restore() error {

	patchData := `{
  "spec": {
    "template": {
      "spec": {
        "nodeSelector": {
          "kubernetes.io/hostname": null
        }
      }
    }
  }
}`
	var item interface{}
	err := r.kubectl.Patch(&item, types.MergePatchType, patchData).Error

	if err != nil {
		return fmt.Errorf("restore %s/%s error %v", r.kubectl.Statement.Namespace, r.kubectl.Statement.Name, err)
	}
	return nil
}
