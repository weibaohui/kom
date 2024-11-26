package kom

import (
	"context"
	"fmt"
	"io"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
)

func (k *Kubectl) WithContext(ctx context.Context) *Kubectl {
	tx := k.getInstance()
	tx.Statement.Context = ctx
	return tx
}
func (k *Kubectl) Resource(obj runtime.Object) *Kubectl {
	tx := k.getInstance()
	tx.Statement.ParseFromRuntimeObj(obj)
	return tx
}
func (k *Kubectl) Namespace(ns string) *Kubectl {
	tx := k.getInstance()
	tx.Statement.Namespace = ns
	return tx
}

// ContainerName
// Deprecated: use Ctl().Pod().ContainerName() instead.
func (k *Kubectl) ContainerName(c string) *Kubectl {
	tx := k.getInstance()
	tx.Statement.ContainerName = c
	return tx
}
func (k *Kubectl) Name(name string) *Kubectl {
	tx := k.getInstance()
	tx.Statement.Name = name
	return tx
}

func (k *Kubectl) CRD(group string, version string, kind string) *Kubectl {
	return k.GVK(group, version, kind)
}
func (k *Kubectl) GVK(group string, version string, kind string) *Kubectl {
	gvk := schema.GroupVersionKind{
		Group:   group,
		Version: version,
		Kind:    kind,
	}
	tx := k.getInstance()
	tx.Statement.useCustomGVK = true
	tx.Statement.ParseGVKs([]schema.GroupVersionKind{
		gvk,
	})

	return tx
}

// Deprecated: use Ctl().Pod().Command() instead.
func (k *Kubectl) Command(command string, args ...string) *Kubectl {
	tx := k.getInstance()
	tx.Statement.Command = command
	tx.Statement.Args = args
	return tx
}

// Deprecated: use Ctl().Pod().Stdin() instead.
func (k *Kubectl) Stdin(reader io.Reader) *Kubectl {
	tx := k.getInstance()
	tx.Statement.Stdin = reader
	return tx
}

// Deprecated: use Ctl().Pod().GetLogs() instead.
func (k *Kubectl) GetLogs(requestPtr interface{}, opt *v1.PodLogOptions) *Kubectl {
	tx := k.getInstance()
	tx.Statement.PodLogOptions = opt
	tx.Statement.PodLogOptions.Container = tx.Statement.ContainerName
	tx.Statement.Dest = requestPtr
	tx.Error = tx.Callback().Logs().Execute(tx)
	return tx
}

func (k *Kubectl) Get(dest interface{}) *Kubectl {
	tx := k.getInstance()
	tx.Statement.Dest = dest
	tx.Error = tx.Callback().Get().Execute(tx)
	return tx
}
func (k *Kubectl) List(dest interface{}, opt ...metav1.ListOptions) *Kubectl {
	tx := k.getInstance()

	// 之前步骤可能使用WithLabelSelector 设置了option
	defaultOpts := tx.Statement.ListOptions
	if len(defaultOpts) == 0 {
		tx.Statement.ListOptions = opt
	} else {
		defaultOpt := tx.Statement.ListOptions[0]
		// 用户提供的 ListOptions（从 opt 中取第一个）
		userOpt := metav1.ListOptions{}
		if len(opt) > 0 {
			userOpt = opt[0]
		}
		// 合并默认选项和用户选项
		finalOpt := mergeListOptions(defaultOpt, userOpt)
		tx.Statement.ListOptions = []metav1.ListOptions{finalOpt}
	}

	tx.Statement.Dest = dest
	tx.Error = tx.Callback().List().Execute(tx)
	return tx
}
func mergeListOptions(opt1, opt2 metav1.ListOptions) metav1.ListOptions {
	return metav1.ListOptions{
		LabelSelector: mergeSelectors(opt1.LabelSelector, opt2.LabelSelector),
		FieldSelector: mergeSelectors(opt1.FieldSelector, opt2.FieldSelector),
	}
}

// 合并两个选择器，使用逗号分隔
func mergeSelectors(selector1, selector2 string) string {
	if selector1 != "" && selector2 != "" {
		return fmt.Sprintf("%s,%s", selector1, selector2)
	}
	if selector1 != "" {
		return selector1
	}
	return selector2
}
func (k *Kubectl) Create(dest interface{}) *Kubectl {
	tx := k.getInstance()
	tx.Statement.Dest = dest
	tx.Error = tx.Callback().Create().Execute(tx)
	return tx
}
func (k *Kubectl) Watch(dest interface{}, opt ...metav1.ListOptions) *Kubectl {
	tx := k.getInstance()
	tx.Statement.ListOptions = opt
	tx.Statement.Dest = dest
	tx.Error = tx.Callback().Watch().Execute(tx)
	return tx
}
func (k *Kubectl) Update(dest interface{}) *Kubectl {
	tx := k.getInstance()
	tx.Statement.Dest = dest
	tx.Error = tx.Callback().Update().Execute(tx)
	return tx
}
func (k *Kubectl) Delete() *Kubectl {
	tx := k.getInstance()
	tx.Error = tx.Callback().Delete().Execute(tx)
	return tx
}
func (k *Kubectl) Patch(dest interface{}, pt types.PatchType, data string) *Kubectl {
	tx := k.getInstance()
	tx.Statement.Dest = dest
	tx.Statement.PatchData = data
	tx.Statement.PatchType = pt
	tx.Error = tx.Callback().Patch().Execute(tx)
	return tx
}

// Execute 请确保dest 是一个指向字节切片的指针。定义var s []byte 使用&s
// Deprecated: use Ctl().Pod().Command().Execute() instead.
func (k *Kubectl) Execute(dest interface{}) *Kubectl {
	tx := k.getInstance()
	tx.Statement.Dest = dest
	tx.Error = tx.Callback().Exec().Execute(tx)
	return tx
}

func (k *Kubectl) WithLabelSelector(labelSelector string) *Kubectl {
	tx := k.getInstance()
	options := tx.Statement.ListOptions

	// 如果 ListOptions 为空，则初始化
	if options == nil || len(options) == 0 {
		tx.Statement.ListOptions = []metav1.ListOptions{
			{
				LabelSelector: labelSelector,
			},
		}
		return tx
	}

	// 合并 LabelSelector
	opt := options[0]
	if opt.LabelSelector != "" {
		opt.LabelSelector += "," + labelSelector
	} else {
		opt.LabelSelector = labelSelector
	}
	tx.Statement.ListOptions = options
	return tx
}

func (k *Kubectl) WithFieldSelector(fieldSelector string) *Kubectl {
	tx := k.getInstance()
	options := tx.Statement.ListOptions

	// 如果 ListOptions 为空，则初始化
	if options == nil || len(options) == 0 {
		tx.Statement.ListOptions = []metav1.ListOptions{
			{
				FieldSelector: fieldSelector,
			},
		}
		return tx
	}

	// 合并 FieldSelector
	opt := options[0]
	if opt.FieldSelector != "" {
		opt.FieldSelector += "," + fieldSelector
	} else {
		opt.FieldSelector = fieldSelector
	}

	return tx
}
