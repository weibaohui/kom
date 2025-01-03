package callbacks

import (
	"github.com/weibaohui/kom/kom"
	"github.com/weibaohui/kom/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

func Create(k *kom.Kubectl) error {

	stmt := k.Statement
	gvr := stmt.GVR
	namespaced := stmt.Namespaced
	ns := stmt.Namespace
	ctx := stmt.Context

	// 将 obj 转换为 Unstructured
	unstructuredObj := &unstructured.Unstructured{}
	unstructuredData, err := runtime.DefaultUnstructuredConverter.ToUnstructured(stmt.Dest)
	if err != nil {
		return err // 处理转换错误
	}
	unstructuredObj.SetUnstructuredContent(unstructuredData)
	var res *unstructured.Unstructured

	if namespaced {
		if ns == "" {
			ns = metav1.NamespaceDefault
			unstructuredObj.SetNamespace(ns)
		}
		res, err = stmt.Kubectl.DynamicClient().Resource(gvr).Namespace(ns).Create(ctx, unstructuredObj, metav1.CreateOptions{})
	} else {
		res, err = stmt.Kubectl.DynamicClient().Resource(gvr).Create(ctx, unstructuredObj, metav1.CreateOptions{})
	}

	if err != nil {
		return err
	}
	stmt.RowsAffected = 1
	if stmt.RemoveManagedFields {
		utils.RemoveManagedFields(res)
	}
	return nil
}
