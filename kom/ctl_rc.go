package kom

type replicationController struct {
	kubectl *Kubectl
}

func (r *replicationController) Scale(replicas int32) error {
	return r.kubectl.Ctl().Scale(replicas)
}
func (r *replicationController) Stop() error {
	return r.kubectl.Ctl().Scaler().Stop()
}
func (r *replicationController) Restore() error {
	return r.kubectl.Ctl().Scaler().Restore()
}
