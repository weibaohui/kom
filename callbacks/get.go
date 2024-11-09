package callbacks

import (
	"fmt"

	"github.com/weibaohui/kom/kom"
	"github.com/weibaohui/kom/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

func Get(k *kom.Kubectl) error {

	stmt := k.Statement
	gvr := stmt.GVR
	namespaced := stmt.Namespaced
	ns := stmt.Namespace
	name := stmt.Name
	ctx := stmt.Context

	var res *unstructured.Unstructured
	var err error
	if name == "" {
		err = fmt.Errorf("获取对象必须指定名称")
		return err
	}
	if namespaced {
		if ns == "" {
			ns = "default"
		}
		res, err = stmt.Kubectl.DynamicClient().Resource(gvr).Namespace(ns).Get(ctx, name, metav1.GetOptions{})
	} else {
		res, err = stmt.Kubectl.DynamicClient().Resource(gvr).Get(ctx, name, metav1.GetOptions{})
	}

	if err != nil {
		return err
	}
	stmt.RowsAffected = 1
	if stmt.RemoveManagedFields {
		utils.RemoveManagedFields(res)
	}
	// 将 unstructured 转换回原始对象
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(res.Object, stmt.Dest)
	if err != nil {
		return err
	}
	return nil
}
