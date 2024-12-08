package kom

import (
	"context"
	"io"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/klog/v2"
)

type Statement struct {
	*Kubectl                                        // 基础配置
	RowsAffected        int64                       // 返回受影响的行数
	AllNamespace        bool                        // 所有名空间
	Namespace           string                      // 资源所属命名空间
	Name                string                      // 资源名称
	GVR                 schema.GroupVersionResource // 资源类型
	GVK                 schema.GroupVersionKind     // 资源类型
	Namespaced          bool                        // 是否是命名空间资源
	ListOptions         []metav1.ListOptions        // 列表查询参数,作为可变参数使用，默认只取第一个，也只使用一个
	Context             context.Context             `json:"-"` // 上下文
	Dest                interface{}                 // 返回结果存放对象，一般为结构体指针
	PatchType           types.PatchType             // PATCH类型
	PatchData           string                      // PATCH数据
	RemoveManagedFields bool                        // 是否移除管理字段
	useCustomGVK        bool                        // 如果通过CRD方法设置了GVK，那么就强制使用，不再进行GVK的自动解析
	ContainerName       string                      // 容器名称，执行获取容器内日志等操作使用
	Command             string                      // 容器内执行命令,包括ls、cat以及用户输入的命令
	Args                []string                    // 容器内执行命令参数
	PodLogOptions       *v1.PodLogOptions           `json:"-"` // 获取容器日志使用
	Stdin               io.Reader                   `json:"-"` // 设置输入
}

func (s *Statement) ParseGVKs(gvks []schema.GroupVersionKind, versions ...string) *Statement {

	s.GVR = schema.GroupVersionResource{}
	s.GVK = schema.GroupVersionKind{}
	// 获取单个GVK
	gvk := s.Tools().GetGVK(gvks, versions...)
	s.GVK = gvk

	// 获取GVR
	if s.Tools().IsBuiltinResource(gvk.Kind) {
		// 内置资源
		if s.useCustomGVK {
			// 设置了CRD，带有version
			s.GVR, s.Namespaced = s.Tools().GetGVRByGVK(gvk)
		} else {
			s.GVR, s.Namespaced = s.Tools().GetGVRByKind(gvk.Kind)
		}
		klog.V(6).Infof("useCustomGVK=%v \n GVR=%v \n GVK=%v", s.useCustomGVK, s.GVR, s.GVK)
	} else {
		crd, err := s.Tools().GetCRD(gvk.Kind, gvk.Group)
		if err != nil {
			return s
		}
		// 检查CRD是否是Namespaced
		s.Namespaced = crd.Object["spec"].(map[string]interface{})["scope"].(string) == "Namespaced"
		s.GVR = s.Tools().GetGVRFromCRD(crd)

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
