package callbacks

import (
	"fmt"
	"reflect"

	"github.com/weibaohui/kom/kom"
	"github.com/weibaohui/kom/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

// List todo 删除这个ctx参数，ctx从statement中获取
func List(k *kom.Kubectl) error {

	stmt := k.Statement
	gvr := stmt.GVR
	namespaced := stmt.Namespaced
	ns := stmt.Namespace
	ctx := stmt.Context

	opts := stmt.ListOptions
	listOptions := metav1.ListOptions{}
	if len(opts) > 0 {
		listOptions = opts[0]
	}

	// 使用反射获取 dest 的值
	destValue := reflect.ValueOf(stmt.Dest)

	// 确保 dest 是一个指向切片的指针
	if destValue.Kind() != reflect.Ptr || destValue.Elem().Kind() != reflect.Slice {
		// 处理错误：dest 不是指向切片的指针
		return fmt.Errorf("请传入数组类型")
	}
	// 获取切片的元素类型
	elemType := destValue.Elem().Type().Elem()

	var list *unstructured.UnstructuredList
	var err error

	if namespaced {
		if ns == "" {
			ns = "default"
		}
		list, err = stmt.Kubectl.DynamicClient().Resource(gvr).Namespace(ns).List(ctx, listOptions)
	} else {
		list, err = stmt.Kubectl.DynamicClient().Resource(gvr).List(ctx, listOptions)
	}
	if err != nil {
		return err
	}

	utils.SortByCreationTime(list.Items)
	// 先清空之前的值
	destValue.Elem().Set(reflect.MakeSlice(destValue.Elem().Type(), 0, 0))

	for _, item := range list.Items {
		obj := item.DeepCopy()
		if stmt.RemoveManagedFields {
			utils.RemoveManagedFields(obj)
		}
		// 创建新的指向元素类型的指针
		newElemPtr := reflect.New(elemType)
		// unstructured 转换为原始目标类型
		err = runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Object, newElemPtr.Interface())
		// 将指针的值添加到切片中
		destValue.Elem().Set(reflect.Append(destValue.Elem(), newElemPtr.Elem()))

	}
	stmt.RowsAffected = int64(len(list.Items))

	if err != nil {
		return err
	}
	return nil
}
