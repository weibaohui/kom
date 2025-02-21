package callbacks

import (
	"fmt"
	"reflect"

	"github.com/weibaohui/kom/kom"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
)

func Watch(k *kom.Kubectl) error {

	stmt := k.Statement
	gvr := stmt.GVR
	namespaced := stmt.Namespaced
	ns := stmt.Namespace
	ctx := stmt.Context
	namespaceList := stmt.NamespaceList

	opts := stmt.ListOptions
	listOptions := metav1.ListOptions{}
	if len(opts) > 0 {
		listOptions = opts[0]
	}

	destValue := reflect.ValueOf(stmt.Dest)

	// 确保 dest 是一个指向接口的指针
	if destValue.Kind() != reflect.Ptr || destValue.Elem().Kind() != reflect.Interface {
		return fmt.Errorf("stmt.Dest 必须是指向 watch.Interface 的指针")
	}

	// 确保 dest 的实际类型实现了 watch.Interface 接口
	if !destValue.Elem().Type().Implements(reflect.TypeOf((*watch.Interface)(nil)).Elem()) {
		return fmt.Errorf("stmt.Dest 必须实现 watch.Interface 接口")
	}

	var watcher watch.Interface
	var err error

	if namespaced {
		if stmt.AllNamespace || len(namespaceList) > 1 {
			// 全部命名空间 或者  传入多个命名空间
			// client-go 不支持跨命名空间查询，就全部查出来，后面再过滤
			ns = metav1.NamespaceAll
		} else {
			// 不是全部，也没有传多个命名空间
			if ns == "" {
				ns = metav1.NamespaceDefault
			}
		}

		watcher, err = stmt.Kubectl.DynamicClient().Resource(gvr).Namespace(ns).Watch(ctx, listOptions)
	} else {
		watcher, err = stmt.Kubectl.DynamicClient().Resource(gvr).Watch(ctx, listOptions)
	}
	if err != nil {
		return err
	}

	// 将 watch 赋值给 dest
	destValue.Elem().Set(reflect.ValueOf(watcher))

	return nil
}
