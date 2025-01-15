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
func (r *replicaSet) Stop() error {
	return r.kubectl.Ctl().Scaler().Stop()
}
func (r *replicaSet) Restore() error {
	return r.kubectl.Ctl().Scaler().Restore()
}
