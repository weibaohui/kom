package kom

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
)

func (k *Kom) WithContext(ctx context.Context) *Kom {
	tx := k.getInstance()
	tx.Statement.Context = ctx
	return tx
}
func (k *Kom) Resource(obj runtime.Object) *Kom {
	tx := k.getInstance()
	tx.Statement.ParseFromRuntimeObj(obj)
	return tx
}
func (k *Kom) Namespace(ns string) *Kom {
	tx := k.getInstance()
	tx.Statement.Namespace = ns
	return tx
}
func (k *Kom) Name(name string) *Kom {
	tx := k.getInstance()
	tx.Statement.Name = name
	return tx
}

func (k *Kom) CRD(group string, version string, kind string) *Kom {
	gvk := schema.GroupVersionKind{
		Group:   group,
		Version: version,
		Kind:    kind,
	}
	k.Statement.ParseGVKs([]schema.GroupVersionKind{
		gvk,
	})

	return k
}

func (k *Kom) Get(dest interface{}) *Kom {
	tx := k.getInstance()
	// 设置目标对象为 obj 的指针
	tx.Statement.Dest = dest
	tx.Error = tx.Callback().Get().Execute(tx)
	return tx
}
func (k *Kom) List(dest interface{}, opt ...metav1.ListOptions) *Kom {
	tx := k.getInstance()
	tx.Statement.ListOptions = opt
	tx.Statement.Dest = dest
	tx.Error = tx.Callback().List().Execute(tx)
	return tx
}
func (k *Kom) Create(dest interface{}) *Kom {
	tx := k.getInstance()
	tx.Statement.Dest = dest
	tx.Error = tx.Callback().Create().Execute(tx)
	return tx
}
func (k *Kom) Update(dest interface{}) *Kom {
	tx := k.getInstance()
	tx.Statement.Dest = dest
	tx.Error = tx.Callback().Update().Execute(tx)
	return tx
}
func (k *Kom) Delete() *Kom {
	tx := k.getInstance()
	tx.Error = tx.Callback().Delete().Execute(tx)
	return tx
}
func (k *Kom) Patch(dest interface{}) *Kom {
	tx := k.getInstance()
	tx.Statement.Dest = dest
	tx.Error = tx.Callback().Patch().Execute(tx)
	return tx
}
func (k *Kom) PatchData(data string) *Kom {
	tx := k.getInstance()
	tx.Statement.PatchData = data
	return tx
}
func (k *Kom) PatchType(t types.PatchType) *Kom {
	tx := k.getInstance()
	tx.Statement.PatchType = t
	return tx
}
func (k *Kom) fill(m *unstructured.Unstructured) *Kom {
	tx := k.getInstance()
	if tx.Statement.Dest == nil {
		tx.Error = fmt.Errorf("请先执行Get()、List()等方法")
	}
	// 确保将数据填充到传入的 m 中
	if dest, ok := tx.Statement.Dest.(*unstructured.Unstructured); ok {
		*m = *dest
	}
	return k
}
