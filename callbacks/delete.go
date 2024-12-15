package callbacks

import (
	"fmt"

	"github.com/weibaohui/kom/kom"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Delete(k *kom.Kubectl) error {
	// 前序步骤有任何Error及时终止
	if k.Error != nil {
		return k.Error
	}
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
		if ns == "" {
			ns = "default"
		}
		err = stmt.Kubectl.DynamicClient().Resource(gvr).Namespace(ns).Delete(ctx, name, metav1.DeleteOptions{})
	} else {
		err = stmt.Kubectl.DynamicClient().Resource(gvr).Delete(ctx, name, metav1.DeleteOptions{})
	}

	if err != nil {
		return err
	}
	stmt.RowsAffected = 1
	return nil
}
