package example

import (
	"fmt"
	"testing"

	"github.com/weibaohui/kom/kom"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/watch"
)

func TestPodWatch(t *testing.T) {

	var watcher watch.Interface

	var pod corev1.Pod
	err := kom.DefaultCluster().Resource(&pod).
		AllNamespace().Watch(&watcher).Error
	if err != nil {
		fmt.Printf("Create Watcher Error %v", err)
		return
	}
	defer watcher.Stop()

	for event := range watcher.ResultChan() {
		err := kom.DefaultCluster().Tools().ConvertRuntimeObjectToTypedObject(event.Object, &pod)
		if err != nil {
			fmt.Printf("无法将对象转换为 *v1.Pod 类型: %v", err)
			return
		}
		// 处理事件
		switch event.Type {
		case watch.Added:
			fmt.Printf("Added Pod [ %s/%s ]\n", pod.Namespace, pod.Name)
		case watch.Modified:
			fmt.Printf("Modified Pod [ %s/%s ]\n", pod.Namespace, pod.Name)
		case watch.Deleted:
			fmt.Printf("Deleted Pod [ %s/%s ]\n", pod.Namespace, pod.Name)
		}
	}

}
