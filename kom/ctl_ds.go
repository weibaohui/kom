package kom

import (
	"fmt"

	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

type daemonSet struct {
	kubectl *Kubectl
}

func (d *daemonSet) Restart() error {
	return d.kubectl.Ctl().Rollout().Restart()
}
func (d *daemonSet) Stop() error {

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
	err := d.kubectl.Patch(&item, types.MergePatchType, patchData).Error

	if err != nil {
		return fmt.Errorf("stop %s/%s error %v", d.kubectl.Statement.Namespace, d.kubectl.Statement.Name, err)
	}
	return nil
}
func (d *daemonSet) Restore() error {

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
	err := d.kubectl.Patch(&item, types.MergePatchType, patchData).Error

	if err != nil {
		return fmt.Errorf("restore %s/%s error %v", d.kubectl.Statement.Namespace, d.kubectl.Statement.Name, err)
	}
	return nil
}

func (d *daemonSet) ManagedPods() ([]*corev1.Pod, error) {
	//先找到ds
	var ds v1.DaemonSet
	err := d.kubectl.WithCache(d.kubectl.Statement.CacheTTL).Resource(&ds).Get(&ds).Error

	if err != nil {
		return nil, err
	}
	// 通过ds 获取pod
	var podList []*corev1.Pod
	err = d.kubectl.newInstance().WithCache(d.kubectl.Statement.CacheTTL).Resource(&corev1.Pod{}).
		Namespace(d.kubectl.Statement.Namespace).
		Where(fmt.Sprintf("metadata.ownerReferences.name='%s' and metadata.ownerReferences.kind='%s'", ds.GetName(), "DaemonSet")).
		List(&podList).Error
	return podList, err
}
func (d *daemonSet) ManagedPod() (*corev1.Pod, error) {
	podList, err := d.ManagedPods()
	if err != nil {
		return nil, err
	}
	if len(podList) > 0 {
		return podList[0], nil
	}
	return nil, fmt.Errorf("未发现DaemonSet[%s]下的Pod", d.kubectl.Statement.Name)
}
