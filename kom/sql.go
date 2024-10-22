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

func (kom *Kom) WithContext(ctx context.Context) *Kom {
	tx := kom.getInstance()
	tx.Statement.Context = ctx
	return tx
}
func (kom *Kom) Resource(obj runtime.Object) *Kom {
	tx := kom.getInstance()
	tx.Statement.ParseFromRuntimeObj(obj)
	return tx
}
func (kom *Kom) Namespace(ns string) *Kom {
	tx := kom.getInstance()
	tx.Statement.Namespace = ns
	return tx
}
func (kom *Kom) Name(name string) *Kom {
	tx := kom.getInstance()
	tx.Statement.Name = name
	return tx
}

func (kom *Kom) CRD(group string, version string, kind string) *Kom {
	gvk := schema.GroupVersionKind{
		Group:   group,
		Version: version,
		Kind:    kind,
	}
	kom.Statement.ParseGVKs([]schema.GroupVersionKind{
		gvk,
	})

	return kom
}

func (kom *Kom) Get(dest interface{}) *Kom {
	tx := kom.getInstance()
	// 设置目标对象为 obj 的指针
	tx.Statement.Dest = dest
	tx.Error = tx.Callback().Get().Execute(tx.Statement.Context, tx)
	return tx
}
func (kom *Kom) List(dest interface{}, opt ...metav1.ListOptions) *Kom {
	tx := kom.getInstance()
	tx.Statement.ListOptions = opt
	tx.Statement.Dest = dest
	tx.Error = tx.Callback().List().Execute(tx.Statement.Context, tx)
	return tx
}
func (kom *Kom) Create(dest interface{}) *Kom {
	tx := kom.getInstance()
	tx.Statement.Dest = dest
	tx.Error = tx.Callback().Create().Execute(tx.Statement.Context, tx)
	return tx
}
func (kom *Kom) Update(dest interface{}) *Kom {
	tx := kom.getInstance()
	tx.Statement.Dest = dest
	tx.Error = tx.Callback().Update().Execute(tx.Statement.Context, tx)
	return tx
}
func (kom *Kom) Delete() *Kom {
	tx := kom.getInstance()
	tx.Error = tx.Callback().Delete().Execute(tx.Statement.Context, tx)
	return tx
}
func (kom *Kom) Patch(dest interface{}) *Kom {
	tx := kom.getInstance()
	tx.Statement.Dest = dest
	tx.Error = tx.Callback().Patch().Execute(tx.Statement.Context, tx)
	return tx
}
func (kom *Kom) PatchData(data string) *Kom {
	tx := kom.getInstance()
	tx.Statement.PatchData = data
	return tx
}
func (kom *Kom) PatchType(t types.PatchType) *Kom {
	tx := kom.getInstance()
	tx.Statement.PatchType = t
	return tx
}
func (kom *Kom) fill(m *unstructured.Unstructured) *Kom {
	tx := kom.getInstance()
	if tx.Statement.Dest == nil {
		tx.Error = fmt.Errorf("请先执行Get()、List()等方法")
	}
	// 确保将数据填充到传入的 m 中
	if dest, ok := tx.Statement.Dest.(*unstructured.Unstructured); ok {
		*m = *dest
	}
	return kom
}
