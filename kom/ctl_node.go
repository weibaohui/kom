package kom

import (
	"fmt"
	"html/template"
	"strings"
	"time"

	"github.com/duke-git/lancet/v2/maputil"
	"github.com/duke-git/lancet/v2/random"
	"github.com/duke-git/lancet/v2/slice"
	"github.com/weibaohui/kom/utils"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
)

type node struct {
	kubectl *Kubectl
}

// Cordon node
// cordon 命令的核心功能是将节点标记为 Unschedulable。在此状态下，调度器（Scheduler）将不会向该节点分配新的 Pod。
func (d *node) Cordon() error {
	var item interface{}
	patchData := `{"spec":{"unschedulable":true}}`
	err := d.kubectl.Patch(&item, types.StrategicMergePatchType, patchData).Error
	return err
}

// UnCordon node
// uncordon 命令是 cordon 的逆操作，用于将节点从不可调度状态恢复为可调度状态。
func (d *node) UnCordon() error {
	var item interface{}
	patchData := `{"spec":{"unschedulable":null}}`
	err := d.kubectl.Patch(&item, types.StrategicMergePatchType, patchData).Error
	return err
}

// Taint node
// taint 命令用于给节点打标签，以表示节点上某些 Pod 不应该运行。
// taint 命令的语法格式为：kubectl taint node <node-name> <key>=<value>:<effect>
// 其中，<key>、<value> 和 <effect> 分别表示标签的键、值和作用域。
// effect 的值可以是 NoSchedule、PreferNoSchedule 或 NoExecute。
// taint 命令会向节点的 metadata.annotations 字段中添加一个名为 taints 的键值对，
// Example
// Taint("dedicated2=special-user:NoSchedule")
// Taint("dedicated2:NoSchedule")
func (d *node) Taint(str string) error {
	taint, err := parseTaint(str)
	if err != nil {
		return err
	}
	var original *corev1.Node
	err = d.kubectl.Get(&original).Error
	if err != nil {
		return err
	}
	taints := original.Spec.Taints
	if taints == nil || len(taints) == 0 {
		taints = []corev1.Taint{*taint}
	} else {
		taints = append(taints, *taint)
	}

	var item interface{}
	patchData := fmt.Sprintf(`{"spec":{"taints":%s}}`, utils.ToJSON(taints))
	err = d.kubectl.Patch(&item, types.StrategicMergePatchType, patchData).Error
	return err
}
func (d *node) UnTaint(str string) error {
	taint, err := parseTaint(str)
	if err != nil {
		return err
	}
	var original *corev1.Node
	err = d.kubectl.Get(&original).Error
	if err != nil {
		return err
	}
	taints := original.Spec.Taints
	if taints == nil || len(taints) == 0 {
		return fmt.Errorf("taint %s not found", str)
	}

	taints = slice.Filter(taints, func(index int, item corev1.Taint) bool {
		return item.Key != taint.Key
	})
	var item interface{}
	patchData := fmt.Sprintf(`{"spec":{"taints":%s}}`, utils.ToJSON(taints))
	err = d.kubectl.Patch(&item, types.StrategicMergePatchType, patchData).Error
	return err
}

// AllNodeLabels 获取所有节点的标签
func (d *node) AllNodeLabels() (map[string]string, error) {

	var list []*corev1.Node
	err := d.kubectl.newInstance().WithContext(d.kubectl.Statement.Context).Resource(&corev1.Node{}).WithCache(d.kubectl.Statement.CacheTTL).
		List(&list).Error
	if err != nil {
		return nil, err
	}
	var labels map[string]string
	for _, n := range list {
		if len(n.Labels) > 0 {
			labels = maputil.Merge(labels, n.Labels)
		}
	}
	return labels, nil
}

// Drain node
// drain 通常在节点需要进行维护时使用。它不仅会标记节点为不可调度，还会逐一驱逐（Evict）该节点上的所有 Pod。
func (d *node) Drain() error {
	// todo 增加--force的处理，也就强制驱逐所有pod，即便是不满足PDB
	name := d.kubectl.Statement.Name

	// Step 1: 将节点标记为不可调度
	klog.V(8).Infof("node/%s  cordoned\n", name)
	err := d.Cordon()
	if err != nil {
		klog.V(8).Infof("node/%s  cordon error %v\n", name, err.Error())
		return err
	}

	// Step 2: 获取节点上的所有 Pod
	// 列出节点上的pod
	var podList []*corev1.Pod
	err = d.kubectl.newInstance().WithContext(d.kubectl.Statement.Context).Resource(&corev1.Pod{}).
		WithFieldSelector(fmt.Sprintf("spec.nodeName=%s", name)).
		List(&podList).Error
	if err != nil {
		klog.V(8).Infof("list pods in node/%s  error %v\n", name, err.Error())
		return err
	}

	// Step 3: 驱逐所有可驱逐的 Pod
	for _, pod := range podList {
		if isDaemonSetPod(pod) || isMirrorPod(pod) {
			// 忽略 DaemonSet 和 Mirror Pod
			klog.V(8).Infof("ignore evict pod  %s/%s  \n", pod.Namespace, pod.Name)
			continue
		}
		klog.V(8).Infof("pod/%s eviction started", pod.Name)

		// 驱逐 Pod
		err := d.evictPod(pod)
		if err != nil {
			klog.V(8).Infof("failed to evict pod %s: %v", pod.Name, err)
			return fmt.Errorf("failed to evict pod %s: %v", pod.Name, err)
		}
		klog.V(8).Infof("pod/%s evictied", pod.Name)
	}

	// Step 4: 等待所有 Pod 被驱逐
	err = wait.PollImmediate(2*time.Second, 5*time.Minute, func() (bool, error) {
		var podList []*corev1.Pod
		err = d.kubectl.newInstance().WithContext(d.kubectl.Statement.Context).Resource(&corev1.Pod{}).
			WithFieldSelector(fmt.Sprintf("spec.nodeName=%s", name)).
			List(&podList).Error
		if err != nil {
			klog.V(8).Infof("list pods in node/%s  error %v\n", name, err.Error())
			return false, err
		}
		for _, pod := range podList {
			if isDaemonSetPod(pod) || isMirrorPod(pod) {
				// 忽略 DaemonSet 和 Mirror Pod
				klog.V(8).Infof("ignore evict pod  %s/%s  \n", pod.Namespace, pod.Name)
				continue
			}
			klog.V(8).Infof("pod/%s eviction started", pod.Name)

			// 驱逐 Pod
			err := d.evictPod(pod)
			if err != nil {
				return false, fmt.Errorf("failed to evict pod %s: %v", pod.Name, err)
			}
			klog.V(8).Infof("pod/%s evictied", pod.Name)
		}
		return true, nil
	})
	if err != nil {
		return fmt.Errorf("timeout waiting for pods to be evicted: %w", err)
	}

	klog.V(8).Infof("node/%s drained", name)
	return nil
}

// CreateNodeShell 获取节点NodeShell
// 要求容器内必须含有nsenter
func (d *node) CreateNodeShell(image ...string) (namespace, podName, containerName string, err error) {
	// 获取节点
	runImage := "alpine:latest"
	if len(image) > 0 {
		runImage = image[0]
	}
	namespace = "kube-system"
	containerName = "shell"
	podName = fmt.Sprintf("node-shell-%s", strings.ToLower(random.RandString(8)))
	var yaml = `
apiVersion: v1
kind: Pod
metadata:
  name: %s
  namespace: %s
spec:
  containers:
  - args:
    - -t
    - "1"
    - -m
    - -u
    - -i
    - -n
    - sleep
    - "14000"
    command:
    - nsenter
    image: %s
    name: %s
    imagePullPolicy: IfNotPresent
    securityContext:
      privileged: true
  hostIPC: true
  hostNetwork: true
  hostPID: true
  restartPolicy: Never
  nodeName: %s	
  tolerations:
  - operator: Exists
`
	yaml = fmt.Sprintf(yaml, podName, namespace, runImage, containerName, d.kubectl.Statement.Name)

	ret := d.kubectl.Applier().Apply(yaml)
	// [Pod/node-shell-xqrbqqvt created]
	// 检查是否包含 created
	klog.V(6).Infof("%s Node Shell 创建 结果 %s", d.kubectl.Statement.Name, ret)

	// 创建成功
	if len(ret) > 0 && strings.Contains(ret[0], "created") {
		// 等待启动或者超时,超时采用默认的超时时间
		err = d.waitPodReady(namespace, podName, d.kubectl.Statement.CacheTTL)
		return
	}

	// 创建失败
	err = fmt.Errorf("node shell 创建失败 %s", ret)
	return
}
func (d *node) waitPodReady(ns, podName string, ttl time.Duration) error {
	if ttl < time.Second {
		klog.Warningf("传入的 ttl=%v 太小，强制设置为默认 30s", ttl)
		ttl = 30 * time.Second
	}
	timeout := time.After(ttl)
	start := time.Now()
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			return fmt.Errorf("等待 Pod %s/%s 启动超时，实际等待了 %v", ns, podName, time.Since(start))
		case <-ticker.C:
			var p *v1.Pod
			err := d.kubectl.newInstance().
				WithContext(d.kubectl.Statement.Context).
				Resource(&v1.Pod{}).
				Name(podName).
				Namespace(ns).
				Get(&p).Error

			if err != nil {
				klog.V(6).Infof("等待 Pod %s/%s 创建中...", ns, podName)
				continue
			}

			if p == nil {
				klog.V(6).Infof("Pod %s/%s 未创建", ns, podName)
				continue
			}

			// 确保容器状态可用
			if len(p.Status.ContainerStatuses) == 0 {
				klog.V(6).Infof("Pod %s/%s 容器状态为空，等待中...", ns, podName)
				continue
			}

			// 检查所有容器是否 ready
			allReady := true
			for _, cs := range p.Status.ContainerStatuses {
				if !cs.Ready {
					klog.V(6).Infof("容器 %s 在 Pod %s/%s 中未就绪", cs.Name, ns, podName)
					allReady = false
					break
				}
			}

			if allReady {
				klog.V(6).Infof("Pod %s/%s 所有容器已就绪", ns, podName)
				return nil
			}

			klog.V(6).Infof("Pod %s/%s 容器未全部就绪，继续等待...", ns, podName)
		}
	}
}

// CreateKubectlShell kubectl 操作shell
// 要求容器内必须含有nsenter
// CreateKubectlShell 创建一个用于运行 kubectl 的 Pod，并传入 kubeconfig 配置内容
func (d *node) CreateKubectlShell(kubeconfig string, image ...string) (namespace, podName, containerName string, err error) {
	// 默认的 kubectl 镜像
	runImage := "bitnami/kubectl:latest"
	if len(image) > 0 {
		runImage = image[0]
	}

	namespace = "kube-system"
	containerName = "shell"
	podName = fmt.Sprintf("kubectl-shell-%s", strings.ToLower(random.RandString(8)))

	// 将 kubeconfig 字符串中的换行符替换为 \n
	kubeconfigEscaped := strings.Replace(kubeconfig, "\n", `\n`, -1)
	// 使用模板字符串来创建 YAML 配置
	podTemplate := `
apiVersion: v1
kind: Pod
metadata:
  name: {{.PodName}}
  namespace: {{.Namespace}}
spec:
  initContainers:
  - name: init-container
    image: {{.RunImage}}
    imagePullPolicy: IfNotPresent
    command: ['sh', '-c', 'echo -e "{{.Kubeconfig}}" > /.kube/config || (echo "Failed to write kubeconfig" && exit 1)']
    volumeMounts:
     - name: kube-config
       mountPath: /.kube
  containers:
  - name: {{.ContainerName}}
    image: {{.RunImage}}
    command: ['tail', '-f', '/dev/null']
    env:
     - name: KUBECONFIG
       value: /.kube/config
    imagePullPolicy: IfNotPresent
    volumeMounts:
    - name: kube-config
      mountPath: /.kube
  nodeName: {{.NodeName}}	
  tolerations:
  - operator: Exists	  
  volumes:
  - name: kube-config
    emptyDir: {}
`

	// 准备模板数据
	data := map[string]interface{}{
		"PodName":       podName,
		"Namespace":     namespace,
		"RunImage":      runImage,
		"Kubeconfig":    kubeconfigEscaped,
		"ContainerName": containerName,
		"NodeName":      d.kubectl.Statement.Name, // 使用 node 上的 kubectl name
	}

	// 创建模板并执行填充
	tmpl, err := template.New("podConfig").Parse(podTemplate)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to parse template: %w", err)
	}

	// 生成最终的 YAML 配置
	var yaml string
	var buf strings.Builder
	err = tmpl.Execute(&buf, data)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to execute template: %w", err)
	}
	yaml = buf.String()

	klog.V(6).Infof("Generated YAML:\n%s", yaml)
	// 调用 kubectl 的 Applier 方法来应用生成的 YAML 配置
	ret := d.kubectl.Applier().Apply(yaml)

	// 检查是否创建成功
	klog.V(6).Infof("%s kubectl Shell 创建结果 %s", d.kubectl.Statement.Name, ret)

	// 如果返回结果中包含 "created" 字符串，则认为创建成功
	if len(ret) > 0 && strings.Contains(ret[0], "created") {
		// 等待启动或者超时,超时采用默认的超时时间
		err = d.waitPodReady(namespace, podName, d.kubectl.Statement.CacheTTL)
		return
	}

	// 创建失败
	err = fmt.Errorf("kubectl shell 创建失败 %s", ret)
	return
}

// 检查是否为 DaemonSet 创建的 Pod
func isDaemonSetPod(pod *corev1.Pod) bool {
	for _, owner := range pod.OwnerReferences {
		if owner.Kind == "DaemonSet" {
			return true
		}
	}
	return false
}

// 检查是否为 Mirror Pod
func isMirrorPod(pod *corev1.Pod) bool {
	_, exists := pod.Annotations[corev1.MirrorPodAnnotationKey]
	return exists
}

// 驱逐 Pod
func (d *node) evictPod(pod *corev1.Pod) error {
	klog.V(8).Infof("evicting pod %s/%s \n", pod.Namespace, pod.Name)
	eviction := &policyv1.Eviction{
		ObjectMeta: metav1.ObjectMeta{
			Name:      pod.Name,
			Namespace: pod.Namespace,
		},
	}
	err := d.kubectl.Client().PolicyV1().Evictions(pod.Namespace).Evict(d.kubectl.Statement.Context, eviction)

	// err := d.kubectl.newInstance().Resource(eviction).Create(eviction).Error
	if err != nil {
		return err
	}
	klog.V(8).Infof(" pod %s/%s evicted\n", pod.Namespace, pod.Name)
	return nil
}

// ParseTaint parses a taint string into a corev1.Taint structure.
func parseTaint(taintStr string) (*corev1.Taint, error) {
	// Split the input string into key-value-effect
	var key, value, effect string
	parts := strings.Split(taintStr, ":")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid taint format: %s", taintStr)
	}
	keyValue := parts[0]
	effect = parts[1]

	// Check the effect
	if effect != string(corev1.TaintEffectNoSchedule) &&
		effect != string(corev1.TaintEffectPreferNoSchedule) &&
		effect != string(corev1.TaintEffectNoExecute) {
		return nil, fmt.Errorf("invalid taint effect: %s", effect)
	}

	// Parse the key and value
	keyValueParts := strings.SplitN(keyValue, "=", 2)
	key = keyValueParts[0]
	if len(keyValueParts) == 2 {
		value = keyValueParts[1]
	}

	// Return the Taint structure
	return &corev1.Taint{
		Key:    key,
		Value:  value,
		Effect: corev1.TaintEffect(effect),
	}, nil
}
func (d *node) RunningPods() ([]*corev1.Pod, error) {
	cacheTime := d.getCacheTTL()

	var podList []*corev1.Pod
	// status.phase!=Succeeded,status.phase!=Failed
	err := d.kubectl.newInstance().WithContext(d.kubectl.Statement.Context).Resource(&corev1.Pod{}).
		AllNamespace().
		Where("spec.nodeName=? and 'status.phase'!='Succeeded' and 'status.phase'!='Failed'", d.kubectl.Statement.Name).
		WithCache(cacheTime).List(&podList).Error
	if err != nil {
		klog.V(6).Infof("list pods in node/%s  error %v\n", d.kubectl.Statement.Name, err.Error())
		return nil, err
	}
	return podList, nil
}

// IPUsage 计算节点上IP数量状态，返回节点IP总数，已用数量，可用数量
func (d *node) IPUsage() (total, used, available int) {
	cacheTime := d.getCacheTTL()
	n, err := d.getNodeWithCache(cacheTime)
	if err != nil {
		klog.V(6).Infof("Get ResourceUsage in node/%s  error %v\n", d.kubectl.Statement.Name, err.Error())
		return 0, 0, 0
	}
	// 计算总数
	cidr := n.Spec.PodCIDR
	count, err := utils.CidrTotalIPs(cidr)
	if err != nil {
		klog.V(6).Infof("Get ResourceUsage in node/%s  error %v\n", d.kubectl.Statement.Name, err.Error())
		return 0, 0, 0
	}
	total = count

	// 计算PodIP数量，
	var podList []*corev1.Pod
	err = d.kubectl.newInstance().WithContext(d.kubectl.Statement.Context).Resource(&corev1.Pod{}).
		AllNamespace().
		Where("spec.nodeName=? and 'status.podIP' != '' ", d.kubectl.Statement.Name).
		WithCache(cacheTime).List(&podList).Error
	if err != nil {
		klog.V(6).Infof("list pods in node/%s  error %v\n", d.kubectl.Statement.Name, err.Error())
		return 0, 0, 0
	}

	used = len(podList)
	available = total - used
	return

}

// getCacheTTL 获取缓存时间
// 默认5秒
func (d *node) getCacheTTL(defaultCacheTime ...time.Duration) time.Duration {
	cacheTime := d.kubectl.Statement.CacheTTL

	if cacheTime == 0 {
		if len(defaultCacheTime) > 0 {
			return defaultCacheTime[0]
		}
		return 10 * time.Second
	}
	return cacheTime
}

// PodCount 计算节点上Pod数量，已经节点Pod数上限
func (d *node) PodCount() (total, used, available int) {
	cacheTime := d.getCacheTTL()
	n, err := d.getNodeWithCache(cacheTime)
	if err != nil {
		klog.V(6).Infof("Get PodCount in node/%s  error %v\n", d.kubectl.Statement.Name, err.Error())
		return 0, 0, 0
	}

	total = int(n.Status.Allocatable.Pods().Value())

	// 计算PodIP数量，
	var podList []*corev1.Pod
	err = d.kubectl.newInstance().WithContext(d.kubectl.Statement.Context).Resource(&corev1.Pod{}).
		AllNamespace().
		Where("spec.nodeName=? ", d.kubectl.Statement.Name).
		WithCache(cacheTime).List(&podList).Error
	if err != nil {
		klog.V(6).Infof("list pods in node/%s  error %v\n", d.kubectl.Statement.Name, err.Error())
		return 0, 0, 0
	}

	used = len(podList)
	available = total - used
	return

}

// getNodeWithCache 获取节点的方法，带缓存
func (d *node) getNodeWithCache(cacheTime time.Duration) (*corev1.Node, error) {
	node, err := utils.GetOrSetCache(
		d.kubectl.ClusterCache(),
		fmt.Sprintf("getNodeWithCache/%s", d.kubectl.Statement.Name),
		d.getCacheTTL(10*time.Second),
		func() (*corev1.Node, error) {
			var n *corev1.Node
			err := d.kubectl.newInstance().WithContext(d.kubectl.Statement.Context).Resource(&corev1.Node{}).
				Name(d.kubectl.Statement.Name).WithCache(cacheTime).Get(&n).Error
			if err != nil {
				klog.V(6).Infof("Get ResourceUsage in node/%s  error %v\n", d.kubectl.Statement.Name, err.Error())
				return nil, err
			}
			return n, nil
		},
	)
	return node, err

}
