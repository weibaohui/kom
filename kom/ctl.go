package kom

type ctl struct {
	kubectl *Kubectl
}

func (c *ctl) Deployment() *deploy {
	return &deploy{
		kubectl: c.kubectl,
	}
}
func (c *ctl) ReplicationController() *replicationController {
	return &replicationController{
		kubectl: c.kubectl,
	}
}
func (c *ctl) ReplicaSet() *replicaSet {
	return &replicaSet{
		kubectl: c.kubectl,
	}
}
func (c *ctl) StatefulSet() *statefulSet {
	return &statefulSet{
		kubectl: c.kubectl,
	}
}
func (c *ctl) Pod() *pod {
	return &pod{
		kubectl: c.kubectl,
	}
}
