package callbacks

import (
	"context"

	"github.com/weibaohui/kom/kom"
	"github.com/weibaohui/kom/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
)

func Create(ctx context.Context, k8s *kom.Kom) error {
	if klog.V(8).Enabled() {
		json := k8s.Statement.String()
		klog.V(8).Infof("DefaultCB Create %s", json)
	}

	stmt := k8s.Statement
	gvr := stmt.GVR
	namespaced := stmt.Namespaced
	ns := stmt.Namespace
	ctx = stmt.Context

	// 将 obj 转换为 Unstructured
	unstructuredObj := &unstructured.Unstructured{}
	unstructuredData, err := runtime.DefaultUnstructuredConverter.ToUnstructured(stmt.Dest)
	if err != nil {
		return err // 处理转换错误
	}
	unstructuredObj.SetUnstructuredContent(unstructuredData)
	var res *unstructured.Unstructured

	if namespaced {
		res, err = stmt.DynamicClient.Resource(gvr).Namespace(ns).Create(ctx, unstructuredObj, metav1.CreateOptions{})
	} else {
		res, err = stmt.DynamicClient.Resource(gvr).Create(ctx, unstructuredObj, metav1.CreateOptions{})
	}

	if err != nil {
		return err
	}
	utils.RemoveManagedFields(res)
	return nil
}
