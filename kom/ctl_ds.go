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

func (r *daemonSet) ManagedPods() ([]*corev1.Pod, error) {
	//先找到ds
	var ds v1.DaemonSet
	err := r.kubectl.Resource(&ds).Get(&ds).Error

	if err != nil {
		return nil, err
	}
	// 通过ds 获取pod
	var podList []*corev1.Pod
	err = r.kubectl.newInstance().Resource(&corev1.Pod{}).
		Namespace(r.kubectl.Statement.Namespace).
		Where(fmt.Sprintf("metadata.ownerReferences.name='%s' and metadata.ownerReferences.kind='%s'", ds.GetName(), "DaemonSet")).
		List(&podList).Error
	return podList, err
}
func (r *daemonSet) ManagedPod() (*corev1.Pod, error) {
	podList, err := r.ManagedPods()
	if err != nil {
		return nil, err
	}
	if len(podList) > 0 {
		return podList[0], nil
	}
	return nil, fmt.Errorf("未发现DaemonSet[%s]下的Pod", r.kubectl.Statement.Name)
}
