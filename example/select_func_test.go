package example

import (
	"fmt"
	"testing"

	"github.com/weibaohui/kom/kom"
	corev1 "k8s.io/api/core/v1"
)

func TestSelectEqual(t *testing.T) {
	var list []corev1.Pod
	err := kom.DefaultCluster().From("pod").
		Where("metadata.namespace ='?'  or metadata.namespace='?'", "kube-system", "default").
		Order("metadata.creationTimestamp` desc").
		List(&list).Error

	if err != nil {
		fmt.Printf("List error %v\n", err)
	}
	fmt.Printf("Count %d\n", len(list))
	for _, d := range list {
		fmt.Printf("List Item  %s\t %s  \t %s \n", d.GetNamespace(), d.GetName(), d.GetCreationTimestamp())
	}
}
func TestSelectLike(t *testing.T) {
	var list []corev1.Pod
	err := kom.DefaultCluster().From("pod").
		Where("metadata.name like '?'", "%random%").
		Order("metadata.creationTimestamp` desc").
		List(&list).Error

	if err != nil {
		fmt.Printf("List error %v\n", err)
	}
	fmt.Printf("Count %d\n", len(list))
	for _, d := range list {
		fmt.Printf("List Item  %s\t %s  \t %s \n", d.GetNamespace(), d.GetName(), d.GetCreationTimestamp())
	}
}
