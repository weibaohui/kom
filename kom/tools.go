package kom

import (
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

func Tools() *tools {
	return &tools{}
}

type tools struct {
}

// ConvertRuntimeObjectToTypedObject 是一个通用的转换函数，将 runtime.Object 转换为指定的目标类型
func (u *tools) ConvertRuntimeObjectToTypedObject(obj runtime.Object, target interface{}) error {
	// 将 obj 断言为 *unstructured.Unstructured 类型
	unstructuredObj, ok := obj.(*unstructured.Unstructured)
	if !ok {
		return fmt.Errorf("无法将对象转换为 *unstructured.Unstructured 类型")
	}

	// 使用 DefaultUnstructuredConverter 将 unstructured 数据转换为具体类型
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(unstructuredObj.Object, target)
	if err != nil {
		return fmt.Errorf("无法将对象转换为目标类型: %v", err)
	}

	return nil
}
func (u *tools) ConvertRuntimeObjectToUnstructuredObject(obj runtime.Object) (*unstructured.Unstructured, error) {
	// 将 obj 断言为 *unstructured.Unstructured 类型
	unstructuredObj, ok := obj.(*unstructured.Unstructured)
	if !ok {
		return nil, fmt.Errorf("无法将对象转换为 *unstructured.Unstructured 类型")
	}

	return unstructuredObj, nil
}
