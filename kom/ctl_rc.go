package kom

import (
	"fmt"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

type replicationController struct {
	kubectl *Kubectl
}

func (r *replicationController) Scale(replicas int32) error {
	var item corev1.ReplicationController

	err := r.kubectl.Get(&item).Error
	if err != nil {
		r.kubectl.Error = errors.Wrapf(err, "ReplicationController Scale Get %s/%s error", r.kubectl.Statement.Namespace, r.kubectl.Statement.Name)
		return r.kubectl.Error
	}
	patchData := fmt.Sprintf("{\"spec\":{\"replicas\":%d}}", replicas)
	err = r.kubectl.Resource(&item).
		Patch(&item, types.MergePatchType, patchData).Error
	if err != nil {
		r.kubectl.Error = errors.Wrapf(err, "ReplicationController Scale %s/%s error", r.kubectl.Statement.Namespace, r.kubectl.Statement.Name)
		return err
	}
	return r.kubectl.Error
}
