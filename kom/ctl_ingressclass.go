package kom

import (
	"fmt"

	v1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
)

type ingressClass struct {
	kubectl *Kubectl
}

// SetDefault 设置为默认ingress类
func (i *ingressClass) SetDefault() error {
	var scList []*v1.IngressClass
	err := i.kubectl.newInstance().
		WithContext(i.kubectl.Statement.Context).
		Resource(&v1.IngressClass{}).
		List(&scList).Error
	if err != nil {
		return err
	}
	if len(scList) == 0 {
		return fmt.Errorf("IngressClass %s not found", i.kubectl.Statement.Name)
	}
	for _, sc := range scList {
		patchData := ""
		// 如果注解中包含默认的注解
		if sc.Annotations != nil && sc.Annotations[v1.AnnotationIsDefaultIngressClass] != "" {
			patchData = fmt.Sprintf(`{"metadata": {"annotations": {"%s": null}}}`, v1.AnnotationIsDefaultIngressClass)
		}

		// 如果名字相符，增加注解
		if sc.Name == i.kubectl.Statement.Name {
			patchData = fmt.Sprintf(`{"metadata": {"annotations": {"%s": "true"}}}`, v1.AnnotationIsDefaultIngressClass)
		}
		if patchData == "" {
			continue
		}
		var item interface{}
		err = i.kubectl.
			newInstance().
			WithContext(i.kubectl.Statement.Context).
			Name(sc.Name).
			Resource(&v1.IngressClass{}).
			Patch(&item, types.StrategicMergePatchType, patchData).Error
		if err != nil {
			klog.V(6).Infof("set %s/%s/%s annotation error %v", i.kubectl.Statement.Namespace, i.kubectl.Statement.Name, sc.Name, err.Error())
			return err
		}
	}
	return nil
}
