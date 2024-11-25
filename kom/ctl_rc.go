package kom

type replicationController struct {
	kubectl *Kubectl
}

func (r *replicationController) Scale(replicas int32) error {
	return r.kubectl.Ctl().Scale(replicas)
}
