package kom

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/duke-git/lancet/v2/slice"
	v1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
)

// rollout支持的资源类型
var rolloutSupportedKinds = []string{"Deployment", "StatefulSet", "DaemonSet", "ReplicaSet"}

type rollout struct {
	kubectl *Kubectl
}

func (d *rollout) logInfo(action string) {
	kind := d.kubectl.Statement.GVK.Kind
	resource := d.kubectl.Statement.GVR.Resource
	namespace := d.kubectl.Statement.Namespace
	name := d.kubectl.Statement.Name
	klog.V(8).Infof("%s Kind=%s", action, kind)
	klog.V(8).Infof("%s Resource=%s", action, resource)
	klog.V(8).Infof("%s %s/%s", action, namespace, name)
}
func (d *rollout) handleError(kind string, namespace string, name string, action string, err error) error {
	if err != nil {
		d.kubectl.Error = fmt.Errorf("%s %s/%s %s error %v", kind, namespace, name, action, err)
		return err
	}
	return nil
}
func (d *rollout) checkResourceKind(kind string, supportedKinds []string) error {
	if !isSupportedKind(kind, supportedKinds) {
		d.kubectl.Error = fmt.Errorf("%s %s/%s operation is not supported", kind, d.kubectl.Statement.Namespace, d.kubectl.Statement.Name)
		return d.kubectl.Error
	}
	return nil
}

func (d *rollout) Restart() error {

	kind := d.kubectl.Statement.GVK.Kind
	d.logInfo("Restart")

	if err := d.checkResourceKind(kind, rolloutSupportedKinds); err != nil {
		return err
	}

	var item interface{}
	patchData := fmt.Sprintf(`{"spec":{"template":{"metadata":{"annotations":{"kom.kubernetes.io/restartedAt":"%s"}}}}}`, time.Now().Format(time.DateTime))
	err := d.kubectl.Patch(&item, types.MergePatchType, patchData).Error
	return d.handleError(kind, d.kubectl.Statement.Namespace, d.kubectl.Statement.Name, "restarting", err)
}
func (d *rollout) Pause() error {
	kind := d.kubectl.Statement.GVK.Kind
	d.logInfo("Pause")

	if err := d.checkResourceKind(kind, []string{"Deployment"}); err != nil {
		return err
	}

	var item interface{}
	patchData := `{"spec":{"paused":true}}`
	err := d.kubectl.Patch(&item, types.MergePatchType, patchData).Error
	return d.handleError(kind, d.kubectl.Statement.Namespace, d.kubectl.Statement.Name, "pause", err)

}
func (d *rollout) Resume() error {
	kind := d.kubectl.Statement.GVK.Kind
	d.logInfo("Resume")

	if err := d.checkResourceKind(kind, []string{"Deployment"}); err != nil {
		return err
	}

	var item interface{}
	patchData := `{"spec":{"paused":null}}`
	err := d.kubectl.Patch(&item, types.MergePatchType, patchData).Error
	return d.handleError(kind, d.kubectl.Statement.Namespace, d.kubectl.Statement.Name, "resume", err)

}

// Status
// 通用字段提取：
//
// spec.replicas：期望的副本数。
// status.replicas：当前实际运行的副本数。
// status.updatedReplicas：已更新为最新版本的副本数。
// status.readyReplicas：通过健康检查并准备好服务的副本数。
// status.unavailableReplicas：当前不可用的副本数。
// Deployment:
//
// 完成条件：updatedReplicas == spec.replicas，且 readyReplicas == spec.replicas，并且 unavailableReplicas == 0。
// StatefulSet:
//
// 完成条件：updatedReplicas == spec.replicas 且 readyReplicas == spec.replicas。
// DaemonSet:
//
// 特有字段：
// status.desiredNumberScheduled：期望调度的节点数。
// status.updatedNumberScheduled：已更新的节点数。
// status.numberReady：健康且就绪的节点数。
// status.numberUnavailable：不可用的节点数。
// 完成条件：updatedNumberScheduled == desiredNumberScheduled，且 numberReady == desiredNumberScheduled，并且 numberUnavailable == 0。
// ReplicaSet:
//
// 完成条件：readyReplicas == spec.replicas。
// 返回状态：
//
// 滚动更新完成时，返回成功消息。
// 更新中返回进度信息。
func (d *rollout) Status() (string, error) {
	kind := d.kubectl.Statement.GVK.Kind
	d.logInfo("Status")

	if err := d.checkResourceKind(kind, rolloutSupportedKinds); err != nil {
		return "", err
	}

	var item unstructured.Unstructured
	err := d.kubectl.Get(&item).Error
	if err != nil {
		return "", d.handleError(kind, d.kubectl.Statement.Namespace, d.kubectl.Statement.Name, "status", err)
	}

	// 提取 replicas 配置
	specReplicas, _, _ := unstructured.NestedInt64(item.Object, "spec", "replicas")
	updatedReplicas, _, _ := unstructured.NestedInt64(item.Object, "status", "updatedReplicas")
	readyReplicas, _, _ := unstructured.NestedInt64(item.Object, "status", "readyReplicas")
	unavailableReplicas, _, _ := unstructured.NestedInt64(item.Object, "status", "unavailableReplicas")

	switch kind {
	case "Deployment":
		// 判断 Deployment 是否完成滚动更新
		if updatedReplicas == specReplicas && readyReplicas == specReplicas && unavailableReplicas == 0 {
			return "Deployment successfully rolled out", nil
		}
		return fmt.Sprintf("Deployment rollout in progress: %d of %d updated, %d ready", updatedReplicas, specReplicas, readyReplicas), nil

	case "StatefulSet":
		// 判断 StatefulSet 是否完成滚动更新
		if updatedReplicas == specReplicas && readyReplicas == specReplicas {
			return "StatefulSet successfully rolled out", nil
		}
		return fmt.Sprintf("StatefulSet rollout in progress: %d of %d updated, %d ready", updatedReplicas, specReplicas, readyReplicas), nil

	case "DaemonSet":
		desiredNumberScheduled, _, _ := unstructured.NestedInt64(item.Object, "status", "desiredNumberScheduled")
		updatedNumberScheduled, _, _ := unstructured.NestedInt64(item.Object, "status", "updatedNumberScheduled")
		numberReady, _, _ := unstructured.NestedInt64(item.Object, "status", "numberReady")
		numberUnavailable, _, _ := unstructured.NestedInt64(item.Object, "status", "numberUnavailable")

		// 判断 DaemonSet 是否完成滚动更新
		if updatedNumberScheduled == desiredNumberScheduled && numberReady == desiredNumberScheduled && numberUnavailable == 0 {
			return "DaemonSet successfully rolled out", nil
		}
		return fmt.Sprintf("DaemonSet rollout in progress: %d of %d updated, %d ready", updatedNumberScheduled, desiredNumberScheduled, numberReady), nil

	case "ReplicaSet":
		// 判断 ReplicaSet 是否完成滚动更新
		if readyReplicas == specReplicas {
			return "ReplicaSet successfully rolled out", nil
		}
		return fmt.Sprintf("ReplicaSet rollout in progress: %d of %d ready", readyReplicas, specReplicas), nil

	default:
		return "", fmt.Errorf("unsupported kind: %s", kind)
	}
}
func (d *rollout) History() (string, error) {
	kind := d.kubectl.Statement.GVK.Kind
	name := d.kubectl.Statement.Name
	ns := d.kubectl.Statement.Namespace
	d.logInfo("History")

	// 校验是否是支持的资源类型
	if err := d.checkResourceKind(kind, []string{"Deployment", "StatefulSet", "DaemonSet"}); err != nil {
		return "", err
	}

	var item unstructured.Unstructured
	err := d.kubectl.Get(&item).Error
	if err != nil {
		return "", d.handleError(kind, d.kubectl.Statement.Namespace, d.kubectl.Statement.Name, "history", err)
	}

	switch kind {
	case "Deployment":
		// 获取 Deployment 的 spec.selector.matchLabels
		labels, found, err := unstructured.NestedMap(item.Object, "spec", "selector", "matchLabels")
		if err != nil || !found {
			return "", fmt.Errorf("failed to get matchLabels from Deployment: %v", err)
		}

		// 构造 labelSelector 字符串，将所有标签拼接起来
		labelSelector := ""
		for key, value := range labels {
			labelSelector += fmt.Sprintf("%s=%s,", key, value)
		}
		// 去除最后一个逗号
		if len(labelSelector) > 0 {
			labelSelector = labelSelector[:len(labelSelector)-1]
		}

		// 查询与 Deployment 关联的所有 ReplicaSet
		var rsList []*v1.ReplicaSet

		err = d.kubectl.newInstance().Resource(&v1.ReplicaSet{}).WithLabelSelector(labelSelector).List(&rsList).Error
		if err != nil {
			return "", fmt.Errorf("failed to list ReplicaSets for Deployment: %v", err)
		}
		rsList = slice.Filter(rsList, func(index int, item *v1.ReplicaSet) bool {
			for _, owner := range item.OwnerReferences {
				if owner.Kind == kind && owner.Name == name {
					return true
				}
			}
			return false
		})
		// 如果没有 ReplicaSet，则没有历史
		if len(rsList) == 0 {
			return "No ReplicaSets found for Deployment", nil
		}

		// 格式化历史记录
		historyStr := "deployment/" + name + " history:\n"
		for _, rs := range rsList {
			rsName := rs.GetName()
			if len(rs.Annotations) > 0 {
				rsRevision := rs.Annotations["deployment.kubernetes.io/revision"]
				historyStr += fmt.Sprintf("ReplicaSet: %s, Revision: %s\n", rsName, rsRevision)
			}

		}
		return historyStr, nil

	case "StatefulSet":

		var versionList []*v1.ControllerRevision
		err = d.kubectl.newInstance().Resource(&v1.ControllerRevision{}).
			Namespace(ns).
			List(&versionList).Error
		if err != nil {
			return "", fmt.Errorf("failed to get controllerrevisions for StatefulSet: %s/%s %v", ns, name, err)
		}

		versionList = d.filterByOwner(versionList, kind, name)

		if len(versionList) == 0 {
			return "No history found for StatefulSet", nil
		}

		// 格式化历史记录
		historyStr := "StatefulSet/" + name + " history:\n"
		for _, rv := range versionList {
			rvName := rv.GetName()

			historyStr += fmt.Sprintf("ControllerRevision: %s, Revision: %d\n", rvName, rv.Revision)
		}
		return historyStr, nil
	case "DaemonSet":

		var versionList []*v1.ControllerRevision
		err = d.kubectl.newInstance().Resource(&v1.ControllerRevision{}).
			Namespace(ns).
			List(&versionList).Error
		if err != nil {
			return "", fmt.Errorf("failed to get controllerrevisions for DaemonSet: %s/%s %v", ns, name, err)
		}
		versionList = d.filterByOwner(versionList, kind, name)
		if len(versionList) == 0 {
			return "No history found for DaemonSet", nil
		}

		// 格式化历史记录
		historyStr := "DaemonSet/" + name + " history:\n"
		for _, rv := range versionList {
			rvName := rv.GetName()

			historyStr += fmt.Sprintf("ControllerRevision: %s, Revision: %d\n", rvName, rv.Revision)
		}
		return historyStr, nil

	default:
		return "", fmt.Errorf("unsupported kind: %s", kind)
	}
}

func (d *rollout) filterByOwner(versionList []*v1.ControllerRevision, kind string, name string) []*v1.ControllerRevision {
	versionList = slice.Filter(versionList, func(index int, item *v1.ControllerRevision) bool {
		for _, owner := range item.OwnerReferences {
			if owner.Kind == kind && owner.Name == name {
				return true
			}
		}
		return false
	})
	return versionList
}
func (d *rollout) Undo(toVersions ...int) (string, error) {
	kind := d.kubectl.Statement.GVK.Kind
	name := d.kubectl.Statement.Name
	namespace := d.kubectl.Statement.Namespace
	toVersion := 0
	if len(toVersions) > 0 {
		toVersion = toVersions[0]
	}
	d.logInfo("Undo")

	// 校验是否是支持的资源类型
	if err := d.checkResourceKind(kind, []string{"Deployment", "DaemonSet", "StatefulSet"}); err != nil {
		return "", err
	}

	var item unstructured.Unstructured
	err := d.kubectl.Get(&item).Error
	if err != nil {
		return "", d.handleError(kind, namespace, name, "Undo", err)
	}

	// 根据资源类型调用不同的回滚方法
	switch kind {
	case "Deployment":
		err = d.rollbackDeployment(toVersion)
	case "StatefulSet":
		err = d.rollbackStatefulSet(toVersion)
	case "DaemonSet":
		err = d.rollbackDaemonSet(toVersion)
	default:
		return "", fmt.Errorf("unsupported kind: %s", kind)
	}

	if err != nil {
		return "", d.handleError(kind, namespace, name, "Undo", err)
	}

	return fmt.Sprintf("%s/%s rolled back successfully", kind, name), nil
}

func (d *rollout) rollbackDeployment(toVersion int) error {
	kind := d.kubectl.Statement.GVK.Kind
	name := d.kubectl.Statement.Name
	ns := d.kubectl.Statement.Namespace
	var deploy v1.Deployment
	err := d.kubectl.Resource(&deploy).
		WithLabelSelector("app=" + name).
		Get(&deploy).Error
	if err != nil {
		return fmt.Errorf(" rollbackDeployment get deployment  err %v ", err)
	}

	if toVersion == 0 {
		// 没有指定版本，则回滚到上一个版本
		revision, err := ExtractDeploymentRevision(deploy.Annotations)
		if err != nil {
			return fmt.Errorf(" rollbackDeployment get deployment revision err %v ", err)
		}
		toVersion = revision - 1
	}

	var rsList []v1.ReplicaSet
	err = d.kubectl.newInstance().Resource(&v1.ReplicaSet{}).
		Namespace(ns).
		List(&rsList).Error
	if err != nil {
		return fmt.Errorf(" rollbackDeployment get rs list err %v ", err)
	}
	var vrs *v1.ReplicaSet
	for _, rs := range rsList {
		owners := rs.OwnerReferences
		if owners != nil && len(owners) > 0 {
			for _, owner := range owners {
				if owner.Kind == kind && owner.Name == name {

					if v, err := ExtractDeploymentRevision(rs.Annotations); err == nil && v == toVersion {
						vrs = &rs
						break
					}
				}
			}
		}
		if vrs != nil {
			break
		}
	}
	if vrs == nil {
		return fmt.Errorf("rollbackDeployment get rs [%s %s] err : not found ", kind, name)
	}
	spec := vrs.Spec.Template.Spec

	deploy.Spec.Template.Spec = spec
	err = d.kubectl.Resource(&deploy).Update(&deploy).Error
	if err != nil {
		return fmt.Errorf(" rollbackDeployment rollout undo deployment  err %v ", err)
	}

	return nil
}

// ExtractDeploymentRevision 从 annotations 中提取 deployment.kubernetes.io/revision 的值，并转换为 int
func ExtractDeploymentRevision(annotations map[string]string) (int, error) {
	const revisionKey = "deployment.kubernetes.io/revision"

	// 检查 annotations 是否为空
	if annotations == nil {
		return 0, errors.New("annotations is nil")
	}

	// 获取 revision 的值
	revisionStr, exists := annotations[revisionKey]
	if !exists {
		return 0, fmt.Errorf("annotation %q not found", revisionKey)
	}

	// 转换为 int
	revision, err := strconv.Atoi(revisionStr)
	if err != nil {
		return 0, fmt.Errorf("failed to convert %q to int: %v", revisionStr, err)
	}

	return revision, nil
}

func (d *rollout) rollbackDaemonSet(toVersion int) error {
	// 从ControllerVersion列表中找指定版本的ControllerVersion
	// 将Revision.Data.Raw转换为DaemonSet
	// 提取DaemonSet的Spec.Template.Spec，赋值到原先的DaemonSet上，更新
	// 完成回滚

	kind := d.kubectl.Statement.GVK.Kind
	name := d.kubectl.Statement.Name
	ns := d.kubectl.Statement.Namespace

	var ds v1.DaemonSet
	err := d.kubectl.Resource(&ds).
		WithLabelSelector("app=" + name).
		Get(&ds).Error
	if err != nil {
		return fmt.Errorf("rollbackDaemonSet get daemonset err %v", err)
	}
	var versionList []*v1.ControllerRevision
	err = d.kubectl.newInstance().Resource(&v1.ControllerRevision{}).
		Namespace(ns).
		List(&versionList).Error
	if err != nil {
		return fmt.Errorf("rollbackDaemonSet list controllerrevisions err %v", err)
	}
	// 如果没有指定版本，则回滚到上一个版本
	if toVersion == 0 {
		// 查找最新的 ControllerRevision 来确定版本
		// 找到最大的version
		var latestRevision int64 = 0
		for _, revision := range versionList {
			for _, owner := range revision.OwnerReferences {
				if owner.Kind == kind && owner.Name == name {
					// 选择目标版本为最新版本
					// 确定最新的版本
					if revision.Revision > latestRevision {
						latestRevision = revision.Revision
					}
				}
			}
		}

		toVersion = int(latestRevision - 1)
		if toVersion <= 0 {
			// 做一个防护，只要是有变更，那么版本号必大于0，最小为1
			toVersion = 1
		}
	}

	// 获取目标版本的 ControllerRevision 并提取 PodTemplateSpec

	// 查找目标版本的 ControllerRevision
	var targetRevision *v1.ControllerRevision

	for _, revision := range versionList {
		for _, owner := range revision.OwnerReferences {
			if owner.Kind == kind && owner.Name == name {
				if int(revision.Revision) == toVersion {
					targetRevision = revision
					break
				}
			}
		}
		if targetRevision != nil {
			break
		}
	}

	if targetRevision == nil {
		return fmt.Errorf("rollbackDaemonSet get target revision %d for %s not found", toVersion, name)
	}

	// 提取目标版本的 PodTemplateSpec
	var dsTemplate v1.DaemonSet
	err = json.Unmarshal(targetRevision.Data.Raw, &dsTemplate)
	if err != nil {
		return fmt.Errorf("rollbackDaemonSet unmarshal controllerrevision data err %v", err)
	}

	// 使用目标版本的模板更新当前 DaemonSet
	ds.Spec.Template.Spec = dsTemplate.Spec.Template.Spec

	// 更新 DaemonSet
	err = d.kubectl.Resource(&ds).Update(&ds).Error
	if err != nil {
		return fmt.Errorf("rollbackDaemonSet update daemonset err %v", err)
	}

	return nil
}
func (d *rollout) rollbackStatefulSet(toVersion int) error {
	// 从ControllerVersion列表中找指定版本的ControllerVersion
	// 将Revision.Data.Raw转换为DaemonSet
	// 提取DaemonSet的Spec.Template.Spec，赋值到原先的DaemonSet上，更新
	// 完成回滚

	kind := d.kubectl.Statement.GVK.Kind
	name := d.kubectl.Statement.Name
	ns := d.kubectl.Statement.Namespace

	var sts v1.StatefulSet
	err := d.kubectl.Resource(&sts).
		WithLabelSelector("app=" + name).
		Get(&sts).Error
	if err != nil {
		return fmt.Errorf("rollbackStatefulSet get StatefulSet err %v", err)
	}
	var versionList []*v1.ControllerRevision
	err = d.kubectl.newInstance().Resource(&v1.ControllerRevision{}).
		Namespace(ns).
		List(&versionList).Error
	if err != nil {
		return fmt.Errorf("rollbackStatefulSet list controllerrevisions err %v", err)
	}
	// 如果没有指定版本，则回滚到上一个版本
	if toVersion == 0 {
		// 查找最新的 ControllerRevision 来确定版本
		// 找到最大的version
		var latestRevision int64 = 0
		for _, revision := range versionList {
			for _, owner := range revision.OwnerReferences {
				if owner.Kind == kind && owner.Name == name {
					// 选择目标版本为最新版本
					// 确定最新的版本
					if revision.Revision > latestRevision {
						latestRevision = revision.Revision
					}
				}
			}
		}

		toVersion = int(latestRevision - 1)
		if toVersion <= 0 {
			// 做一个防护，只要是有变更，那么版本号必大于0，最小为1
			toVersion = 1
		}
	}

	// 获取目标版本的 ControllerRevision 并提取 PodTemplateSpec

	// 查找目标版本的 ControllerRevision
	var targetRevision *v1.ControllerRevision

	for _, revision := range versionList {
		for _, owner := range revision.OwnerReferences {
			if owner.Kind == kind && owner.Name == name {
				if int(revision.Revision) == toVersion {
					targetRevision = revision
					break
				}
			}
		}
		if targetRevision != nil {
			break
		}
	}

	if targetRevision == nil {
		return fmt.Errorf("rollbackStatefulSet get target revision %d for %s not found", toVersion, name)
	}

	// 提取目标版本的 PodTemplateSpec
	var stsTemplate v1.StatefulSet
	err = json.Unmarshal(targetRevision.Data.Raw, &stsTemplate)
	if err != nil {
		return fmt.Errorf("rollbackStatefulSet unmarshal controllerrevision data err %v", err)
	}

	// 使用目标版本的模板更新当前 DaemonSet
	sts.Spec.Template.Spec = stsTemplate.Spec.Template.Spec

	// 更新 DaemonSet
	err = d.kubectl.Resource(&sts).Update(&sts).Error
	if err != nil {
		return fmt.Errorf("rollbackStatefulSet update daemonset err %v", err)
	}

	return nil
}
