package kom

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
)

type Statement struct {
	*Kubectl
	Error        error
	RowsAffected int64
	Namespace    string
	Name         string
	GVR          schema.GroupVersionResource
	GVK          schema.GroupVersionKind
	Namespaced   bool
	ListOptions  []metav1.ListOptions
	Context      context.Context `json:"-"`
	Dest         interface{}
	PatchType    types.PatchType
	PatchData    string
	clean        bool // 移除管理字段
}

func (s *Statement) SetNamespace(ns string) *Statement {
	s.Namespace = ns
	return s
}
func (s *Statement) SetName(name string) *Statement {
	s.Name = name
	return s
}

func (s *Statement) setGVR(gvr schema.GroupVersionResource) *Statement {
	s.GVR = gvr
	return s
}

func (s *Statement) SetDest(dest interface{}) *Statement {
	s.Dest = dest
	return s
}

func (s *Statement) ParseGVKs(gvks []schema.GroupVersionKind, versions ...string) *Statement {

	s.GVR = schema.GroupVersionResource{}
	s.GVK = schema.GroupVersionKind{}
	// 获取单个GVK
	gvk := getParsedGVK(gvks, versions...)
	s.GVK = gvk

	// 获取GVR
	if s.Kubectl.isBuiltinResource(gvk.Kind) {
		// 内置资源
		s.GVR, s.Namespaced = s.Kubectl.getGVR(gvk.Kind)

	} else {
		crd, err := s.Kubectl.getCRD(gvk.Kind, gvk.Group)
		if err != nil {
			s.Error = err
			return s
		}
		// 检查CRD是否是Namespaced
		s.Namespaced = crd.Object["spec"].(map[string]interface{})["scope"].(string) == "Namespaced"
		s.GVR = getGRVFromCRD(crd)

	}

	return s
}

func (s *Statement) ParseNsNameFromRuntimeObj(obj runtime.Object) *Statement {
	// 获取元数据（比如Name和Namespace）
	accessor, err := meta.Accessor(obj)
	if err != nil {
		s.Error = err
		return s
	}
	s.Name = accessor.GetName()           // 获取资源的名称
	s.Namespace = accessor.GetNamespace() // 获取资源的命名空间
	return s
}

func (s *Statement) ParseGVKFromRuntimeObj(obj runtime.Object) *Statement {
	// 使用 scheme.Scheme.ObjectKinds() 获取 Kind
	gvks, _, err := scheme.Scheme.ObjectKinds(obj)
	if err != nil {
		s.Error = fmt.Errorf("error getting kind by scheme.Scheme.ObjectKinds : %v", err)
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
