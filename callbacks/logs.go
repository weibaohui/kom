package callbacks

import (
	"fmt"
	"reflect"

	"github.com/weibaohui/kom/kom"
)

func GetLogs(k *kom.Kubectl) error {

	stmt := k.Statement
	ns := stmt.Namespace
	name := stmt.Name
	options := stmt.PodLogOptions
	options.Container = stmt.ContainerName
	ctx := stmt.Context

	if stmt.ContainerName == "" {
		return fmt.Errorf("请调用ContainerName()方法设置Pod容器名称")
	}

	// 使用反射获取 dest 的值
	destValue := reflect.ValueOf(stmt.Dest)

	// 确保 dest 是一个指针
	if destValue.Kind() != reflect.Ptr {
		// 处理错误：dest 不是指向切片的指针
		return fmt.Errorf("目标容器必须是指针类型")
	}

	stream, err := k.Client().CoreV1().Pods(ns).GetLogs(name, options).Stream(ctx)
	if err != nil {
		return err
	}
	// 将流赋值给 dest
	destValue.Elem().Set(reflect.ValueOf(stream))
	return nil
}
