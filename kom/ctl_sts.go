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

func (s *statefulSet) Stop() error {
	return s.kubectl.Ctl().Scaler().Stop()
}
func (s *statefulSet) Restore() error {
	return s.kubectl.Ctl().Scaler().Restore()
}
