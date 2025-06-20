package kom

import (
	"context"
	"io"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/klog/v2"
)

type Statement struct {
	*Kubectl             `json:"Kubectl,omitempty"`                                          // 基础配置
	RowsAffected         int64                        `json:"rowsAffected,omitempty"`        // 返回受影响的行数
	TotalCount           *int64                       `json:"totalCount,omitempty"`          // 返回查询总数，分页使用。只在List查询列表方法中生效。
	AllNamespace         bool                         `json:"allNamespace,omitempty"`        // 所有名空间
	Namespace            string                       `json:"namespace,omitempty"`           // 资源所属命名空间
	NamespaceList        []string                     `json:"namespace_list,omitempty"`      // 多个命名空间，查询列表专用，只有查询列表时会出现跨命名空间查询的情况。在使用时，如果是所有命名空间，就不用NamespaceList
	Name                 string                       `json:"name,omitempty"`                // 资源名称
	GVR                  schema.GroupVersionResource  `json:"GVR"`                           // 资源类型
	GVK                  schema.GroupVersionKind      `json:"GVK"`                           // 资源类型
	Namespaced           bool                         `json:"namespaced,omitempty"`          // 是否是命名空间资源
	ListOptions          []metav1.ListOptions         `json:"listOptions,omitempty"`         // 列表查询参数,作为可变参数使用，默认只取第一个，也只使用一个
	Context              context.Context              `json:"-"`                             // 上下文
	Dest                 interface{}                  `json:"dest,omitempty"`                // 返回结果存放对象，一般为结构体指针
	PatchType            types.PatchType              `json:"patchType,omitempty"`           // PATCH类型
	PatchData            string                       `json:"patchData,omitempty"`           // PATCH数据
	RemoveManagedFields  bool                         `json:"removeManagedFields,omitempty"` // 是否移除管理字段
	useCustomGVK         bool                         `json:"-"`                             // 如果通过CRD方法设置了GVK，那么就强制使用，不再进行GVK的自动解析
	ContainerName        string                       `json:"containerName,omitempty"`       // 容器名称，执行获取容器内日志等操作使用
	Command              string                       `json:"command,omitempty"`             // 容器内执行命令,包括ls、cat以及用户输入的命令
	DocField             string                       `json:"doc_field,omitempty"`           // doc 字段，如spec.spec
	Args                 []string                     `json:"args,omitempty"`                // 容器内执行命令参数
	PodLogOptions        *v1.PodLogOptions            `json:"-" `                            // 获取容器日志使用
	Stdin                io.Reader                    `json:"-" `                            // 设置输入
	StreamOptions        *remotecommand.StreamOptions `json:"-"`                             // 设置Stream 参数
	Filter               Filter                       `json:"filter,omitempty"`
	StdoutCallback       func(data []byte) error      `json:"-"`
	StderrCallback       func(data []byte) error      `json:"-"`
	CacheTTL             time.Duration                `json:"cacheTTL,omitempty"`    // 设置缓存时间
	ForceDelete          bool                         `json:"forceDelete,omitempty"` // 强制删除标志
	PortForwardLocalPort string                       `json:"port_forward_local_port"`
	PortForwardPodPort   string                       `json:"port_forward_pod_port"`
	PortForwardStopCh    chan struct{}                `json:"-"`
}
type Filter struct {
	Columns    []string    `json:"columns,omitempty"`
	Conditions []Condition `json:"condition,omitempty"` // xx=?
	Order      string      `json:"order,omitempty"`
	Limit      int         `json:"limit,omitempty"`
	Offset     int         `json:"offset,omitempty"`
	Sql        string      `json:"sql,omitempty"`    // 原始sql
	Parsed     bool        `json:"parsed,omitempty"` // 是否解析过
	From       string      `json:"from,omitempty"`   // From TableName
}
type Condition struct {
	Depth     int
	AndOr     string
	Field     string
	Operator  string
	Value     interface{} // 通过detectType 赋值为精确类型值，detectType之前都是string
	ValueType string      // number, string, bool, time
}

func (s *Statement) ParseGVKs(gvks []schema.GroupVersionKind, versions ...string) *Statement {

	s.GVR = schema.GroupVersionResource{}
	s.GVK = schema.GroupVersionKind{}
	// 获取单个GVK
	gvk := s.Tools().GetGVK(gvks, versions...)
	s.GVK = gvk
	gvr, namespaced, ok := s.Tools().GetGVRByGVK(gvk)
	if ok {
		s.GVR, s.Namespaced = gvr, namespaced
	} else {
		s.GVR, s.Namespaced = s.Tools().GetGVRByKind(gvk.Kind)
	}
	return s
}

func (s *Statement) ParseNsNameFromRuntimeObj(obj runtime.Object) *Statement {
	// 获取元数据（比如Name和Namespace）
	accessor, err := meta.Accessor(obj)
	if err != nil {
		klog.V(6).Infof("error getting meta data by meta.Accessor : %v", err)
		return s
	}
	if name := accessor.GetName(); name != "" {
		s.Name = name // 获取资源的名称
	}
	if namespace := accessor.GetNamespace(); namespace != "" {
		s.Namespace = namespace // 获取资源的命名空间
	}
	return s
}

func (s *Statement) ParseGVKFromRuntimeObj(obj runtime.Object) *Statement {
	// 使用 scheme.Scheme.ObjectKinds() 获取 Kind
	gvks, _, err := scheme.Scheme.ObjectKinds(obj)
	if err != nil {
		klog.V(6).Infof("error getting kind by scheme.Scheme.ObjectKinds : %v", err)
		return s
	}
	s.ParseGVKs(gvks)
	return s
}

func (s *Statement) ParseFromRuntimeObj(obj runtime.Object) *Statement {
	return s.
		ParseGVKFromRuntimeObj(obj).
		ParseNsNameFromRuntimeObj(obj)
}
