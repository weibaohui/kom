package callbacks

import (
	"fmt"

	"github.com/weibaohui/kom/kom"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

func Delete(kom *kom.Kom) error {
	if klog.V(8).Enabled() {
		json := kom.Statement.String()
		klog.V(8).Infof("DefaultCB Delete %s", json)
	}

	stmt := kom.Statement
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
		err = stmt.DynamicClient.Resource(gvr).Namespace(ns).Delete(ctx, name, metav1.DeleteOptions{})
	} else {
		err = stmt.DynamicClient.Resource(gvr).Delete(ctx, name, metav1.DeleteOptions{})
	}

	if err != nil {
		return err
	}

	return nil
}
