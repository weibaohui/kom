package kom

import (
	"fmt"
	"time"

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

	if err := d.checkResourceKind(kind, rolloutSupportedKinds); err != nil {
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

	if err := d.checkResourceKind(kind, rolloutSupportedKinds); err != nil {
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
