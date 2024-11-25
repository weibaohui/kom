package kom

type daemonSet struct {
	kubectl *Kubectl
}

func (r *daemonSet) Restart() error {
	return r.kubectl.Ctl().Rollout().Restart()
}
