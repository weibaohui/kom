package kom

type statefulSet struct {
	kubectl *Kubectl
}

func (s *statefulSet) Restart() error {
	return s.kubectl.Ctl().Rollout().Restart()
}
func (s *statefulSet) Scale(replicas int32) error {
	return s.kubectl.Ctl().Scale(replicas)
}
