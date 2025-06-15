package kom

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

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

// Namespace 设置命名空间 *
// 传入*代表取所有的命名空间，等同于调用AllNamespace()方法
func (k *Kubectl) Namespace(namespaces ...string) *Kubectl {
	tx := k.getInstance()
	if tx.Statement.NamespaceList == nil {
		tx.Statement.NamespaceList = make([]string, 0)
	}
	if len(namespaces) == 1 {
		ns := namespaces[0]
		if ns == "*" {
			// all namespaces
			tx.Statement.AllNamespace = true
			tx.Statement.Namespace = metav1.NamespaceAll
		} else if ns == "" {
			// 与kubectl 保持一致，不传表示限制在default 命名空间中
			tx.Statement.AllNamespace = false
			tx.Statement.Namespace = metav1.NamespaceDefault
		} else {
			tx.Statement.AllNamespace = false
			tx.Statement.Namespace = ns
		}
		return tx
	}
	for _, ns := range namespaces {
		if ns == "*" {
			// 只要出现了*，那么就是所有命名空间
			// 在使用时，如果是所有，就不用NamespaceList
			tx.Statement.AllNamespace = true
			tx.Statement.NamespaceList = make([]string, 0)
			break
		}
		tx.Statement.NamespaceList = append(tx.Statement.NamespaceList, ns)
	}

	var parts []string
	for _, ns := range tx.Statement.NamespaceList {
		parts = append(parts, fmt.Sprintf("metadata.namespace='%s'", ns))
	}
	result := strings.Join(parts, " or ")
	result = fmt.Sprintf("(%s)", result)
	if result != "()" {
		k.Where(result)
	}
	return tx
}
func (k *Kubectl) AllNamespace() *Kubectl {
	tx := k.getInstance()
	tx.Statement.AllNamespace = true
	return tx
}
func (k *Kubectl) RemoveManagedFields() *Kubectl {
	tx := k.getInstance()
	tx.Statement.RemoveManagedFields = true
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
func (k *Kubectl) WithCache(ttl time.Duration) *Kubectl {
	tx := k.getInstance()
	tx.Statement.CacheTTL = ttl
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

// DocField 获取字段信息,如spec.spec
func (k *Kubectl) DocField(field string) *Kubectl {
	tx := k.getInstance()
	tx.Statement.DocField = field
	return tx
}
func (k *Kubectl) Doc(dest interface{}) *Kubectl {
	tx := k.getInstance()
	tx.Statement.Dest = dest
	tx.Error = tx.Callback().Doc().Execute(tx)
	return tx
}
func (k *Kubectl) Describe(dest interface{}) *Kubectl {
	tx := k.getInstance()
	tx.Statement.Dest = dest
	tx.Error = tx.Callback().Describe().Execute(tx)
	return tx
}
func (k *Kubectl) List(dest interface{}, opt ...metav1.ListOptions) *Kubectl {
	tx := k.getInstance()

	// 先判断opt是否有值，没有值，不用处理了。
	// 如果opt没有值，那么前面步骤使用WithLabelSelector，那就沿用，没用就为空。
	// 如果opt有值，使用 opt进行合并
	if opt != nil && len(opt) >= 0 {
		// 之前步骤可能使用WithLabelSelector 设置了option
		if len(tx.Statement.ListOptions) == 0 {
			// 之前也没有设置值，那么直接使用opt
			tx.Statement.ListOptions = opt
		} else {
			// 之前有值，需要合并值
			// 之前的值只可能是在selector，所以应该以现在的opt为基准，合并之前opt的selector
			preOpt := tx.Statement.ListOptions[0]
			currentOpt := opt[0]
			currentOpt.LabelSelector = mergeSelectors(preOpt.LabelSelector, currentOpt.LabelSelector)
			currentOpt.FieldSelector = mergeSelectors(preOpt.FieldSelector, currentOpt.FieldSelector)

			tx.Statement.ListOptions = []metav1.ListOptions{currentOpt}
		}
	}

	tx.Statement.Dest = dest
	tx.Error = tx.Callback().List().Execute(tx)
	return tx
}

// mergeSelectors 合并两个选择器字符串，若两者均非空则用逗号连接，否则返回非空的选择器。
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
func (k *Kubectl) ForceDelete() *Kubectl {
	tx := k.getInstance()
	tx.Statement.ForceDelete = true
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
	tx.Statement.ListOptions = []metav1.ListOptions{opt}

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
	tx.Statement.ListOptions = []metav1.ListOptions{opt}
	return tx
}

func (k *Kubectl) FillTotalCount(total *int64) *Kubectl {
	tx := k.getInstance()
	tx.Statement.TotalCount = total
	return tx
}
