package applier

import (
	"context"
	"fmt"
	"strings"

	"github.com/weibaohui/kom/kom"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/yaml"
)

type Applier struct {
	Kom *kom.Kom
}

func Instance() *Applier {
	return &Applier{
		Kom: kom.Init(),
	}
}
func Cluster(id string) *Applier {
	return &Applier{
		Kom: kom.Clusters().GetById(id).Kom,
	}
}

func (a *Applier) WithContext(ctx context.Context) *Applier {
	a.Kom = a.Kom.WithContext(ctx)
	return a
}

func (a *Applier) Apply(str string) (result []string) {
	docs := splitYAML(str)

	for _, doc := range docs {
		if strings.TrimSpace(doc) == "" {
			continue
		}
		// 解析 YAML 到 Unstructured 对象
		var obj unstructured.Unstructured
		if err := yaml.Unmarshal([]byte(doc), &obj.Object); err != nil {
			result = append(result, fmt.Sprintf("YAML 解析失败: %v", err))
			continue
		}
		result = append(result, a.createOrUpdateCRD(&obj))
	}

	return result
}
func (a *Applier) Delete(str string) (result []string) {
	docs := splitYAML(str)

	for _, doc := range docs {
		if strings.TrimSpace(doc) == "" {
			continue
		}
		// 解析 YAML 到 Unstructured 对象
		var obj unstructured.Unstructured
		if err := yaml.Unmarshal([]byte(doc), &obj.Object); err != nil {
			result = append(result, fmt.Sprintf("YAML 解析失败: %v", err))
			continue
		}
		result = append(result, a.deleteCRD(&obj))
	}

	return result
}
func (a *Applier) createOrUpdateCRD(obj *unstructured.Unstructured) string {
	// 提取 Group, Version, Kind
	gvk := obj.GroupVersionKind()
	if gvk.Kind == "" || gvk.Version == "" {
		return fmt.Sprintf("YAML 缺少必要的 Group, Version 或 Kind")
	}

	_, namespaced := a.Kom.ParseGVK2GVR([]schema.GroupVersionKind{gvk})

	ns := obj.GetNamespace()
	name := obj.GetName()
	kind := obj.GetKind()

	if ns == "" && namespaced {
		ns = "default" // 默认命名空间
		obj.SetNamespace(ns)
	}
	var cr *unstructured.Unstructured
	err := a.Kom.CRD(gvk.Group, gvk.Version, gvk.Kind).Namespace(ns).Name(name).Get(&cr).Error

	if err == nil && cr != nil && cr.GetName() != "" {
		// 已经存在资源，那么就更新
		obj.SetResourceVersion(cr.GetResourceVersion())
		err = a.Kom.CRD(gvk.Group, gvk.Version, gvk.Kind).Name(name).Namespace(ns).Update(&obj).Error
		if err != nil {
			return fmt.Sprintf("update %s/%s,%s %s/%s error:%v", gvk.Group, gvk.Version, gvk.Kind, ns, name, err)
		}
		return fmt.Sprintf("%s/%s updated", kind, name)
	} else {
		// 不存在，那么就创建
		err = a.Kom.CRD(gvk.Group, gvk.Version, gvk.Kind).Name(name).Namespace(ns).Create(&obj).Error
		if err != nil {
			return fmt.Sprintf("create %s/%s,%s %s/%s error:%v", gvk.Group, gvk.Version, gvk.Kind, ns, name, err)
		}
		return fmt.Sprintf("%s/%s created", kind, name)
	}
}
func (a *Applier) deleteCRD(obj *unstructured.Unstructured) string {
	// 提取 Group, Version, Kind
	gvk := obj.GroupVersionKind()
	if gvk.Kind == "" || gvk.Version == "" {
		return fmt.Sprintf("YAML 缺少必要的 Group, Version 或 Kind")
	}
	ns := obj.GetNamespace()
	name := obj.GetName()
	err := a.Kom.CRD(gvk.Group, gvk.Version, gvk.Kind).Namespace(ns).Name(name).Delete().Error
	if err != nil {
		return fmt.Sprintf("delete %s/%s,%s %s/%s error:%v", gvk.Group, gvk.Version, gvk.Kind, ns, name, err)
	}
	return fmt.Sprintf("%s/%s deleted", gvk.Kind, name)
}

// splitYAML 按 "---" 分割多文档 YAML
func splitYAML(yamlStr string) []string {
	return strings.Split(yamlStr, "\n---\n")
}
