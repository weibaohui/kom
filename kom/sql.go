package kom

import (
	"context"
	"fmt"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
)

func (k8s *Kom) WithContext(ctx context.Context) *Kom {
	tx := k8s.getInstance()
	tx.Statement.Context = ctx
	return tx
}
func (k8s *Kom) Resource(obj runtime.Object) *Kom {
	tx := k8s.getInstance()
	tx.Statement.ParseFromRuntimeObj(obj)
	return tx
}
func (k8s *Kom) Namespace(ns string) *Kom {
	tx := k8s.getInstance()
	tx.Statement.Namespace = ns
	return tx
}
func (k8s *Kom) Name(name string) *Kom {
	tx := k8s.getInstance()
	tx.Statement.Name = name
	return tx
}

func (k8s *Kom) CRD(group string, version string, kind string) *Kom {
	gvk := schema.GroupVersionKind{
		Group:   group,
		Version: version,
		Kind:    kind,
	}
	k8s.Statement.ParseGVKs([]schema.GroupVersionKind{
		gvk,
	})

	return k8s
}

func (k8s *Kom) Get(dest interface{}) *Kom {
	tx := k8s.getInstance()
	// 设置目标对象为 obj 的指针
	tx.Statement.Dest = dest
	tx.Error = tx.Callback().Get().Execute(tx.Statement.Context, tx)
	return tx
}
func (k8s *Kom) List(dest interface{}) *Kom {
	tx := k8s.getInstance()
	tx.Statement.Dest = dest
	tx.Error = tx.Callback().List().Execute(tx.Statement.Context, tx)
	return tx
}
func (k8s *Kom) Create(dest interface{}) *Kom {
	tx := k8s.getInstance()
	tx.Statement.Dest = dest
	tx.Error = tx.Callback().Create().Execute(tx.Statement.Context, tx)
	return tx
}
func (k8s *Kom) Update(dest interface{}) *Kom {
	tx := k8s.getInstance()
	tx.Statement.Dest = dest
	tx.Error = tx.Callback().Update().Execute(tx.Statement.Context, tx)
	return tx
}
func (k8s *Kom) Delete(dest interface{}) *Kom {
	tx := k8s.getInstance()
	tx.Statement.Dest = dest
	tx.Error = tx.Callback().Delete().Execute(tx.Statement.Context, tx)
	return tx
}
func (k8s *Kom) Patch(dest interface{}) *Kom {
	tx := k8s.getInstance()
	tx.Statement.Dest = dest
	tx.Error = tx.Callback().Patch().Execute(tx.Statement.Context, tx)
	return tx
}
func (k8s *Kom) PatchData(data string) *Kom {
	tx := k8s.getInstance()
	tx.Statement.PatchData = data
	return tx
}
func (k8s *Kom) PatchType(t types.PatchType) *Kom {
	tx := k8s.getInstance()
	tx.Statement.PatchType = t
	return tx
}
func (k8s *Kom) Fill(m *unstructured.Unstructured) *Kom {
	tx := k8s.getInstance()
	if tx.Statement.Dest == nil {
		tx.Error = fmt.Errorf("请先执行Get()、List()等方法")
	}
	// 确保将数据填充到传入的 m 中
	if dest, ok := tx.Statement.Dest.(*unstructured.Unstructured); ok {
		*m = *dest
	}
	return k8s
}

func (k8s *Kom) sqlTest() {
	pod := v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      "",
		},
	}
	err := k8s.Resource(&pod).
		Namespace("default").Name("").Get(&pod)
	if err != nil {
		klog.Errorf("k8s.First(&pod) error :%v", err)
	}
}
