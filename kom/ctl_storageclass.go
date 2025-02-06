package kom

import v1 "k8s.io/api/core/v1"

type storageClass struct {
	kubectl *Kubectl
}

// PVCCount 统计PVC数量
func (s *storageClass) PVCCount() (int, error) {
	var pvcList []*v1.PersistentVolumeClaim
	err := s.kubectl.newInstance().
		WithContext(s.kubectl.Statement.Context).
		WithCache(s.kubectl.Statement.CacheTTL).
		Resource(&v1.PersistentVolumeClaim{}).
		AllNamespace().
		Where("spec.storageClassName=?", s.kubectl.Statement.Name).
		List(&pvcList).Error
	if err != nil {
		return 0, err
	}
	return len(pvcList), nil
}
