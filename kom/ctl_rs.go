package kom

type replicaSet struct {
	kubectl *Kubectl
}

func (r *replicaSet) Restart() error {
	return r.kubectl.Ctl().Rollout().Restart()
}
func (r *replicaSet) Scale(replicas int32) error {
	return r.kubectl.Ctl().Scale(replicas)
}
