package callbacks

import (
	"fmt"
	"reflect"

	"github.com/weibaohui/kom/kom"
	"github.com/weibaohui/kom/kom/doc"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"
)

func Doc(k *kom.Kubectl) error {

	stmt := k.Statement
	gvk := k.Statement.GVK
	field := stmt.DocField

	if stmt.GVK.Empty() {
		return fmt.Errorf("请调用GVK()方法设置GroupVersionKind")
	}

	// 反射检查
	destValue := reflect.ValueOf(stmt.Dest)

	// 确保 dest 是一个指向字节切片的指针
	if !(destValue.Kind() == reflect.Ptr && destValue.Elem().Kind() == reflect.Slice) || destValue.Elem().Type().Elem().Kind() != reflect.Uint8 {
		return fmt.Errorf("请确保dest 是一个指向字节切片的指针。定义var s []byte 使用&s")
	}

	apiDoc := doc.DocField{
		Kind: gvk.Kind,
		ApiVersion: schema.GroupVersion{
			Group:   gvk.Group,
			Version: gvk.Version,
		},
		OpenapiSchema: k.Status().OpenAPISchema(),
	}
	result := apiDoc.GetApiDocV2(field)

	// 将结果写入 tx.Statement.Dest
	if destBytes, ok := k.Statement.Dest.(*[]byte); ok {
		*destBytes = []byte(result)
		klog.V(8).Infof("Doc result %s", *destBytes)
	} else {
		return fmt.Errorf("dest is not a *[]byte")
	}
	return nil
}
