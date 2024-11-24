package kom

import (
	"fmt"

	"github.com/pkg/errors"
	v1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/types"
)

type replicaSet struct {
	kubectl *Kubectl
}

func (r *replicaSet) Scale(replicas int32) error {
	var item v1.ReplicaSet

	err := r.kubectl.Get(&item).Error
	if err != nil {
		r.kubectl.Error = errors.Wrapf(err, "ReplicaSet Scale Get %s/%s error", r.kubectl.Statement.Namespace, r.kubectl.Statement.Name)
		return r.kubectl.Error
	}
	patchData := fmt.Sprintf("{\"spec\":{\"replicas\":%d}}", replicas)
	err = r.kubectl.Resource(&item).
		Patch(&item, types.MergePatchType, patchData).Error
	if err != nil {
		r.kubectl.Error = errors.Wrapf(err, "ReplicaSet Scale %s/%s error", r.kubectl.Statement.Namespace, r.kubectl.Statement.Name)
		return err
	}
	return r.kubectl.Error
}
