package callbacks

import (
	"fmt"
	"reflect"

	"github.com/weibaohui/kom/kom"
	"github.com/weibaohui/kom/kom/describe"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"
)

func Describe(k *kom.Kubectl) error {

	stmt := k.Statement
	ns := stmt.Namespace
	name := stmt.Name
	gvk := k.Statement.GVK

	if stmt.GVK.Empty() {
		return fmt.Errorf("请调用GVK()方法设置GroupVersionKind")
	}

	// 反射检查
	destValue := reflect.ValueOf(stmt.Dest)

	// 确保 dest 是一个指向字节切片的指针
	if !(destValue.Kind() == reflect.Ptr && destValue.Elem().Kind() == reflect.Slice) || destValue.Elem().Type().Elem().Kind() != reflect.Uint8 {
		return fmt.Errorf("请确保dest 是一个指向字节切片的指针。定义var s []byte 使用&s")
	}

	var output string
	var err error
	// 执行describe
	m := k.Status().DescriberMap()
	gk := schema.GroupKind{
		Group: gvk.Group,
		Kind:  gvk.Kind,
	}
	// 先从内置的describerMap中查找
	if d, ok := m[gk]; ok {
		output, err = d.Describe(ns, name, describe.DescriberSettings{
			ShowEvents: true,
		})
		if err != nil {
			return fmt.Errorf("DescriberMap describe %s/%s error: %v", gvk.String(), name, err)
		}
	} else {
		// 没有内置描述器
		mapping := &meta.RESTMapping{
			Resource: k.Statement.GVR,
		}
		if gd, b := describe.GenericDescriberFor(mapping, k.RestConfig()); b {
			output, err = gd.Describe(ns, name, describe.DescriberSettings{
				ShowEvents: true,
			})
			if err != nil {
				return fmt.Errorf("GenericDescriber describe %s/%s error: %v", gvk.String(), name, err)
			}
		}
	}

	// 将结果写入 tx.Statement.Dest
	if destBytes, ok := k.Statement.Dest.(*[]byte); ok {
		// 直接使用 outBuf.Bytes() 赋值
		*destBytes = []byte(output)
		klog.V(8).Infof("Describe result %s", *destBytes)
	} else {
		return fmt.Errorf("dest is not a *[]byte")
	}
	return nil
}
