package callbacks

import (
	"fmt"

	"github.com/weibaohui/kom/kom"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Delete(k *kom.Kubectl) error {

	stmt := k.Statement
	gvr := stmt.GVR
	namespaced := stmt.Namespaced
	ns := stmt.Namespace
	name := stmt.Name
	ctx := stmt.Context

	var err error
	if name == "" {
		err = fmt.Errorf("删除对象必须指定名称")
		return err
	}
	if namespaced {
		err = stmt.Kubectl.DynamicClient().Resource(gvr).Namespace(ns).Delete(ctx, name, metav1.DeleteOptions{})
	} else {
		err = stmt.Kubectl.DynamicClient().Resource(gvr).Delete(ctx, name, metav1.DeleteOptions{})
	}

	if err != nil {
		return err
	}

	return nil
}
