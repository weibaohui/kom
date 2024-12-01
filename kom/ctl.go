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
func (c *ctl) DaemonSet() *daemonSet {
	return &daemonSet{
		kubectl: c.kubectl,
	}
}
func (c *ctl) Pod() *pod {
	return &pod{
		kubectl: c.kubectl,
	}
}
func (c *ctl) Node() *node {
	return &node{
		kubectl: c.kubectl,
	}
}
func (c *ctl) Rollout() *rollout {
	return &rollout{
		kubectl: c.kubectl,
	}
}
func (c *ctl) Scale(replicas int32) error {
	item := &scale{
		kubectl: c.kubectl,
	}
	return item.Scale(replicas)
}

func isSupportedKind(kind string, supportedKinds []string) bool {
	for _, k := range supportedKinds {
		if kind == k {
			return true
		}
	}
	return false
}
