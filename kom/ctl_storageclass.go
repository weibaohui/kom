package kom

import (
	"fmt"

	v1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	"k8s.io/kubectl/pkg/util/storage"
)

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
		RemoveManagedFields().
		List(&pvcList).Error
	if err != nil {
		return 0, err
	}
	return len(pvcList), nil
}

// PVCount 统计PV数量
func (s *storageClass) PVCount() (int, error) {
	var pvList []*v1.PersistentVolume
	err := s.kubectl.newInstance().
		WithContext(s.kubectl.Statement.Context).
		WithCache(s.kubectl.Statement.CacheTTL).
		Resource(&v1.PersistentVolume{}).
		AllNamespace().
		Where("spec.storageClassName=?", s.kubectl.Statement.Name).
		RemoveManagedFields().
		List(&pvList).Error
	if err != nil {
		return 0, err
	}
	return len(pvList), nil
}

// SetDefault 设置为默认存储类
func (s *storageClass) SetDefault() error {
	var scList []*storagev1.StorageClass
	err := s.kubectl.newInstance().
		WithContext(s.kubectl.Statement.Context).
		Resource(&storagev1.StorageClass{}).
		List(&scList).Error
	if err != nil {
		return err
	}
	if len(scList) == 0 {
		return fmt.Errorf("storageclass %s not found", s.kubectl.Statement.Name)
	}
	for _, sc := range scList {
		patchData := ""
		// 如果注解中包含默认的注解
		if storage.IsDefaultAnnotationText(sc.ObjectMeta) == "Yes" {
			patchData = fmt.Sprintf(`{"metadata": {"annotations": {"%s": null}}}`, storage.IsDefaultStorageClassAnnotation)
		}

		// 如果名字相符，增加注解
		if sc.Name == s.kubectl.Statement.Name {
			patchData = fmt.Sprintf(`{"metadata": {"annotations": {"%s": "true"}}}`, storage.IsDefaultStorageClassAnnotation)
		}
		if patchData == "" {
			continue
		}
		var item interface{}
		err = s.kubectl.
			newInstance().
			WithContext(s.kubectl.Statement.Context).
			Name(sc.Name).
			Resource(&storagev1.StorageClass{}).
			Patch(&item, types.MergePatchType, patchData).Error
		if err != nil {
			klog.V(6).Infof("set %s/%s/%s annotation error %v", s.kubectl.Statement.Namespace, s.kubectl.Statement.Name, sc.Name, err.Error())
			return err
		}
	}
	return nil
}
