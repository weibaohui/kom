package callbacks

import (
	"fmt"

	"github.com/weibaohui/kom/kom"
	"github.com/weibaohui/kom/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Delete(k *kom.Kubectl) error {
	stmt := k.Statement
	gvr := stmt.GVR
	namespaced := stmt.Namespaced
	ns := stmt.Namespace
	name := stmt.Name
	ctx := stmt.Context
	forceDelete := stmt.ForceDelete // 增加强制删除标志

	// 修改删除选项以支持强制删除
	deleteOptions := metav1.DeleteOptions{}
	if forceDelete {
		background := metav1.DeletePropagationBackground
		deleteOptions.PropagationPolicy = &background
		deleteOptions.GracePeriodSeconds = utils.Int64Ptr(0)
	}

	var err error
	if name == "" {
		err = fmt.Errorf("删除对象必须指定名称")
		return err
	}
	if namespaced {
		if ns == "" {
			ns = metav1.NamespaceDefault
		}

		err = stmt.Kubectl.DynamicClient().Resource(gvr).Namespace(ns).Delete(ctx, name, deleteOptions)
	} else {
		err = stmt.Kubectl.DynamicClient().Resource(gvr).Delete(ctx, name, deleteOptions)
	}

	if err != nil {
		return err
	}
	stmt.RowsAffected = 1
	return nil
}
