package callbacks

import (
	"fmt"

	"github.com/weibaohui/kom/kom"
	"github.com/weibaohui/kom/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

func Patch(kom *kom.Kom) error {
	stmt := kom.Statement
	gvr := stmt.GVR
	namespaced := stmt.Namespaced
	ns := stmt.Namespace
	name := stmt.Name
	ctx := stmt.Context
	patchType := stmt.PatchType
	patchData := stmt.PatchData

	var res *unstructured.Unstructured
	var err error
	if name == "" {
		err = fmt.Errorf("patch对象必须指定名称")
		return err
	}
	if namespaced {
		res, err = stmt.Kom.DynamicClient().Resource(gvr).Namespace(ns).Patch(ctx, name, patchType, []byte(patchData), metav1.PatchOptions{})
	} else {
		res, err = stmt.Kom.DynamicClient().Resource(gvr).Patch(ctx, name, patchType, []byte(patchData), metav1.PatchOptions{})
	}
	if err != nil {
		return err
	}

	utils.RemoveManagedFields(res)

	// 将 unstructured 转换回原始对象
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(res.Object, stmt.Dest)
	if err != nil {
		return err
	}

	return nil
}
