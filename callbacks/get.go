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
	var err error
	stmt := k.Statement
	gvr := stmt.GVR
	namespaced := stmt.Namespaced
	ns := stmt.Namespace
	name := stmt.Name
	ctx := stmt.Context
	conditions := stmt.Filter.Conditions
	// 如果设置了where条件。那么应该使用List，因为sql查出来的是list，哪怕是只有一个元素
	if len(conditions) > 0 {
		return fmt.Errorf("SQL 查询方式请使用List承载，如需获取单个资源，请从List中获得")
	}
	if name == "" {
		err = fmt.Errorf("获取对象必须指定名称")
		return err
	}

	cacheKey := fmt.Sprintf("%s/%s/%s/%s/%s", ns, name, gvr.Group, gvr.Resource, gvr.Version)
	res, err := utils.GetOrSetCache(stmt.Kubectl.ClusterCache(), cacheKey, stmt.CacheTTL, func() (ret *unstructured.Unstructured, err error) {
		if namespaced {
			if ns == "" {
				ns = "default"
			}
			ret, err = stmt.Kubectl.DynamicClient().Resource(gvr).Namespace(ns).Get(ctx, name, metav1.GetOptions{})
		} else {
			ret, err = stmt.Kubectl.DynamicClient().Resource(gvr).Get(ctx, name, metav1.GetOptions{})
		}
		return
	})

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
