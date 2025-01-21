package kom

import (
	"fmt"
	"slices"
	"strings"

	"github.com/duke-git/lancet/v2/slice"
	"github.com/weibaohui/kom/utils"
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/klog/v2"
)

// 获取pod相关的Service
func (p *pod) LinkedService() ([]*v1.Service, error) {
	// 	查询流程
	// 获取目标 Pod 的详细信息：

	// 使用 Pod 的 API 获取其 metadata.labels。
	// 确定 Pod 所在的 Namespace。
	// 获取 Namespace 内的所有 Services：

	// 使用 kubectl get services -n {namespace} 或调用 API /api/v1/namespaces/{namespace}/services。
	// 逐个匹配 Service 的 selector：

	// 对每个 Service：
	// 提取其 spec.selector。
	// 遍历 selector 的所有键值对，检查 Pod 是否包含这些标签且值相等。
	// 如果所有标签条件都满足，将此 Service 记录为与该 Pod 关联。
	// 返回结果：

	// 将所有匹配的 Service 名称及相关信息返回。
	var pod *v1.Pod
	err := p.kubectl.Get(&pod).Error
	if err != nil {
		return nil, fmt.Errorf("get pod %s/%s error %v", p.kubectl.Statement.Namespace, p.kubectl.Statement.Name, err.Error())
	}
	if pod == nil {
		return nil, fmt.Errorf("get pod %s/%s error %v", p.kubectl.Statement.Namespace, p.kubectl.Statement.Name, "pod is nil")
	}
	podLabels := pod.GetLabels()

	if len(podLabels) == 0 {
		return nil, nil
	}

	var services []*v1.Service
	err = p.kubectl.newInstance().WithContext(p.kubectl.Statement.Context).
		Resource(&v1.Service{}).
		Namespace(p.kubectl.Statement.Namespace).
		RemoveManagedFields().
		List(&services).Error

	if err != nil {
		return nil, fmt.Errorf("get service error %v", err.Error())
	}

	var result []*v1.Service
	for _, svc := range services {
		serviceLabels := svc.Spec.Selector
		// 遍历selector
		// serviceLabels中所有的kv,都必须在podLabels中存在,且值相等
		if utils.CompareMapContains(serviceLabels, podLabels) {
			result = append(result, svc)
		}
	}
	return result, nil
}

func (p *pod) LinkedEndpoints() ([]*v1.Endpoints, error) {

	services, err := p.LinkedService()
	if err != nil {
		return nil, err
	}
	if len(services) == 0 {
		return nil, nil
	}
	// endpoints 与 svc 同名
	// 1.获取service 名称
	// 2.获取endpoints
	// 3.返回endpoints

	var names []string
	for _, svc := range services {
		names = append(names, svc.Name)
	}

	var endpoints []*v1.Endpoints

	err = p.kubectl.newInstance().
		WithContext(p.kubectl.Statement.Context).
		Resource(&v1.Endpoints{}).
		Namespace(p.kubectl.Statement.Namespace).
		Where("metadata.name in " + utils.StringListToSQLIn(names)).
		RemoveManagedFields().
		List(&endpoints).Error
	if err != nil {
		return nil, err
	}
	return endpoints, nil
}

func (p *pod) LinkedPVC() ([]*v1.PersistentVolumeClaim, error) {

	var pod v1.Pod
	err := p.kubectl.Get(&pod).Error
	if err != nil {
		return nil, err
	}
	// 找打pvc 名称列表
	var pvcNames []string
	for _, volume := range pod.Spec.Volumes {
		if volume.PersistentVolumeClaim != nil {
			pvcNames = append(pvcNames, volume.PersistentVolumeClaim.ClaimName)
		}
	}

	if len(pvcNames) == 0 {
		return nil, nil
	}

	// 找出同ns下pvc的列表，过滤pvcNames
	var pvcList []*v1.PersistentVolumeClaim
	err = p.kubectl.newInstance().WithContext(p.kubectl.Statement.Context).
		Resource(&v1.PersistentVolumeClaim{}).
		Namespace(p.kubectl.Statement.Namespace).
		Where("metadata.name in " + utils.StringListToSQLIn(pvcNames)).
		RemoveManagedFields().
		List(&pvcList).Error
	if err != nil {
		return nil, err
	}

	return pvcList, nil
}
func (p *pod) LinkedPV() ([]*v1.PersistentVolume, error) {

	pvcList, err := p.LinkedPVC()
	if err != nil {
		return nil, err
	}
	var pvNames []string
	for _, pvc := range pvcList {
		pvNames = append(pvNames, pvc.Spec.VolumeName)
	}
	// 找出同ns下pvc的列表，过滤pvcNames
	var pvList []*v1.PersistentVolume
	err = p.kubectl.newInstance().WithContext(p.kubectl.Statement.Context).
		Resource(&v1.PersistentVolume{}).
		Namespace(p.kubectl.Statement.Namespace).
		Where("metadata.name in " + utils.StringListToSQLIn(pvNames)).
		RemoveManagedFields().
		List(&pvList).Error
	if err != nil {
		return nil, err
	}
	return pvList, nil
}

func (p *pod) LinkedIngress() ([]*networkingv1.Ingress, error) {

	var pod v1.Pod
	err := p.kubectl.Get(&pod).Error
	if err != nil {
		return nil, err
	}
	services, err := p.LinkedService()
	if err != nil {
		return nil, err
	}

	if len(services) == 0 {
		return nil, nil
	}

	var servicesName []string
	for _, svc := range services {
		servicesName = append(servicesName, svc.Name)
	}

	// 获取ingress
	// Ingress 通过 spec.rules 或 spec.defaultBackend 中的 service.name 指定关联的 Service。
	// 遍历services，获取ingress
	var ingressList []networkingv1.Ingress
	err = p.kubectl.newInstance().WithContext(p.kubectl.Statement.Context).
		Resource(&networkingv1.Ingress{}).
		Namespace(p.kubectl.Statement.Namespace).
		WithCache(p.kubectl.Statement.CacheTTL).
		RemoveManagedFields().
		List(&ingressList).Error
	if err != nil {
		return nil, err
	}

	// 过滤ingressList，只保留与services关联的ingress
	var result []*networkingv1.Ingress
	for _, ingress := range ingressList {
		if slices.Contains(servicesName, ingress.Spec.Rules[0].Host) {
			result = append(result, &ingress)
		}
	}
	// 遍历 Ingress 检查关联
	for _, ingress := range ingressList {
		if ingress.Spec.DefaultBackend != nil {
			if ingress.Spec.DefaultBackend.Service != nil && ingress.Spec.DefaultBackend.Service.Name != "" {
				if slices.Contains(servicesName, ingress.Spec.DefaultBackend.Service.Name) {
					result = append(result, &ingress)
				}
			}
		}

		for _, rule := range ingress.Spec.Rules {
			if rule.HTTP != nil {
				for _, path := range rule.HTTP.Paths {
					if path.Backend.Service != nil && path.Backend.Service.Name != "" {

						backName := path.Backend.Service.Name
						if slices.Contains(servicesName, backName) {
							result = append(result, &ingress)
						}
					}
				}

			}

		}
	}

	return result, nil
}

// PodMount 是Pod的挂载信息
// 挂载的类型有：configMap、secret
type PodMount struct {
	Name      string `json:"name",omitempty`
	MountPath string `json:"mountPath",omitempty`
	SubPath   string `json:"subPath",omitempty`
	Mode      *int32 `json:"mode",omitempty`
	ReadOnly  bool   `json:"readOnly",omitempty`
}

// LinkedConfigMap 获取Pod相关的ConfigMap
func (p *pod) LinkedConfigMap() ([]*v1.ConfigMap, error) {
	var item *v1.Pod
	err := p.kubectl.Get(&item).Error
	if err != nil {
		return nil, err
	}
	// 找打configmap 名称列表
	var configMapNames []string
	for _, volume := range item.Spec.Volumes {
		if volume.ConfigMap != nil {
			configMapNames = append(configMapNames, volume.ConfigMap.Name)
		}
	}
	if len(configMapNames) == 0 {
		return nil, nil
	}
	// 找出同ns下configmap的列表，过滤configMapNames
	var configMapList []*v1.ConfigMap
	err = p.kubectl.newInstance().WithContext(p.kubectl.Statement.Context).
		Resource(&v1.ConfigMap{}).
		Namespace(p.kubectl.Statement.Namespace).
		RemoveManagedFields().
		Where("metadata.name in " + utils.StringListToSQLIn(configMapNames)).
		List(&configMapList).Error
	if err != nil {
		return nil, err
	}

	// item.Spec.Containers.volumeMounts
	// item.Spec.Volumes
	// 通过遍历secretNames，可以找到pod.Spec.Volumes中的volumeName。
	// 通过volumeName，可以找到pod.Spec.Containers.volumeMounts中的volumeMounts，提取mode
	// 提取volumeMounts中的mountPath、subPath

	for i := range configMapList {
		configMap := configMapList[i]
		var configMapMounts []*PodMount

		configMapName := configMap.Name
		for _, volume := range item.Spec.Volumes {
			if volume.ConfigMap != nil && volume.ConfigMap.Name == configMapName {

				for _, container := range item.Spec.Containers {
					for _, volumeMount := range container.VolumeMounts {
						if volumeMount.Name == volume.Name {
							cm := PodMount{
								Name:      volume.ConfigMap.Name,
								MountPath: volumeMount.MountPath,
								SubPath:   volumeMount.SubPath,
								ReadOnly:  volumeMount.ReadOnly,
								Mode:      volume.ConfigMap.DefaultMode,
							}
							configMapMounts = append(configMapMounts, &cm)
						}
					}
				}

			}
		}

		if len(configMapMounts) > 0 {
			if configMap.Annotations == nil {
				configMap.Annotations = make(map[string]string)
			}
			configMap.Annotations["configMapMounts"] = utils.ToJSON(configMapMounts)
		}
	}

	return configMapList, nil
}

// LinkedSecret 获取Pod相关的Secret
func (p *pod) LinkedSecret() ([]*v1.Secret, error) {
	var item *v1.Pod
	err := p.kubectl.Get(&item).Error
	if err != nil {
		return nil, err
	}
	// 找打secret 名称列表
	var secretNames []string
	for _, volume := range item.Spec.Volumes {
		if volume.Secret != nil {
			secretNames = append(secretNames, volume.Secret.SecretName)
		}
	}
	if len(secretNames) == 0 {
		return nil, nil
	}
	// 找出同ns下secret的列表，过滤secretNames
	var secretList []*v1.Secret
	err = p.kubectl.newInstance().WithContext(p.kubectl.Statement.Context).
		Resource(&v1.Secret{}).
		Namespace(p.kubectl.Statement.Namespace).
		RemoveManagedFields().
		Where("metadata.name in " + utils.StringListToSQLIn(secretNames)).
		List(&secretList).Error
	if err != nil {
		return nil, err
	}

	// item.Spec.Containers.volumeMounts
	// item.Spec.Volumes
	// 通过遍历secretNames，可以找到pod.Spec.Volumes中的volumeName。
	// 通过volumeName，可以找到pod.Spec.Containers.volumeMounts中的volumeMounts，提取mode
	// 提取volumeMounts中的mountPath、subPath

	for i := range secretList {
		secret := secretList[i]
		var secretMounts []*PodMount

		secretName := secret.Name
		for _, volume := range item.Spec.Volumes {
			if volume.Secret != nil && volume.Secret.SecretName == secretName {

				for _, container := range item.Spec.Containers {
					for _, volumeMount := range container.VolumeMounts {
						if volumeMount.Name == volume.Name {
							sm := PodMount{
								Name:      volume.Secret.SecretName,
								MountPath: volumeMount.MountPath,
								SubPath:   volumeMount.SubPath,
								ReadOnly:  volumeMount.ReadOnly,
								Mode:      volume.Secret.DefaultMode,
							}
							secretMounts = append(secretMounts, &sm)
						}
					}
				}

			}
		}

		if len(secretMounts) > 0 {
			if secret.Annotations == nil {
				secret.Annotations = make(map[string]string)
			}
			secret.Annotations["secretMounts"] = utils.ToJSON(secretMounts)
		}
	}

	return secretList, nil
}

// Env 每行三个值，容器名称、ENV名称、ENV值
type Env struct {
	ContainerName string `json:"containerName,omitempty"`
	EnvName       string `json:"envName,omitempty"`
	EnvValue      string `json:"envValue,omitempty"`
}

func (p *pod) LinkedEnv() ([]*Env, error) {
	// 先获取容器列表，然后获取容器的环境变量，然后组装到Env结构体中

	// 先获取pod，从pod中读取容器列表
	var item *v1.Pod
	err := p.kubectl.Get(&item).Error
	if err != nil {
		return nil, err
	}

	var envs []*Env

	// 获取容器名称列表
	for _, container := range item.Spec.Containers {

		// 进到容器中执行ENV命令，获取输出字符串
		var result []byte
		err = p.kubectl.newInstance().Resource(&v1.Pod{}).
			WithContext(p.kubectl.Statement.Context).
			Namespace(p.kubectl.Statement.Namespace).
			Name(p.kubectl.Statement.Name).Ctl().Pod().
			ContainerName(container.Name).
			Command("env").
			Execute(&result).Error
		if err != nil {
			klog.V(6).Infof("get %s/%s/%s env error %v", p.kubectl.Statement.Namespace, p.kubectl.Statement.Name, container.Name, err.Error())
			return nil, err
		}

		// 解析result，获取ENV名称和ENV值
		envArrays := strings.Split(string(result), "\n")
		for _, envline := range envArrays {
			envArray := strings.Split(envline, "=")
			if len(envArray) != 2 {
				continue
			}
			envs = append(envs, &Env{ContainerName: container.Name, EnvName: envArray[0], EnvValue: envArray[1]})
		}
	}

	return envs, nil
}

// LinkedEnvFromPod 提取pod 定义中的env 定义
func (p *pod) LinkedEnvFromPod() ([]*Env, error) {
	// 先获取pod，从pod中读取容器列表
	var pod v1.Pod
	err := p.kubectl.Get(&pod).Error
	if err != nil {
		return nil, err
	}
	var envs []*Env
	for _, container := range pod.Spec.Containers {

		for _, env := range container.Env {

			envHolder := &Env{ContainerName: container.Name, EnvName: env.Name, EnvValue: env.Value}
			if envHolder.EnvValue != "" {
				envs = append(envs, envHolder)
				continue
			}

			// ref 有多种情况，需要判断
			// FieldRef\ResourceFieldRef\ConfigMapKeyRef\SecretKeyRef
			// 分别获取这四种情况的值，应该是四种中的某一种
			// 获取env.ValueFrom.FieldRef.FieldPath的值
			if env.ValueFrom != nil && env.ValueFrom.FieldRef != nil && env.ValueFrom.FieldRef.FieldPath != "" {
				envHolder.EnvValue = fmt.Sprintf("[Field] %s", env.ValueFrom.FieldRef.FieldPath)
			}

			// 		 - name: CPU_REQUEST
			//     valueFrom:
			//       resourceFieldRef:
			//         containerName: multi-env-container
			//         resource: requests.cpu
			//   - name: MEMORY_LIMIT
			//     valueFrom:
			//       resourceFieldRef:
			//         containerName: multi-env-container
			//         resource: limits.memory
			if env.ValueFrom != nil && env.ValueFrom.ResourceFieldRef != nil && env.ValueFrom.ResourceFieldRef.Resource != "" {
				envHolder.EnvValue = fmt.Sprintf("[Container] %s/%s", env.ValueFrom.ResourceFieldRef.ContainerName, env.ValueFrom.ResourceFieldRef.Resource)
			}

			// configMapKeyRef:
			// name: my-env-configmap
			// key: env.list
			if env.ValueFrom != nil && env.ValueFrom.ConfigMapKeyRef != nil && env.ValueFrom.ConfigMapKeyRef.Key != "" {
				envHolder.EnvValue = fmt.Sprintf("[ConfigMap] %s/%s", env.ValueFrom.ConfigMapKeyRef.Name, env.ValueFrom.ConfigMapKeyRef.Key)
			}

			// secretKeyRef:
			// name: db-credentials
			// key: DB_PASSWORD
			if env.ValueFrom != nil && env.ValueFrom.SecretKeyRef != nil && env.ValueFrom.SecretKeyRef.Key != "" {
				envHolder.EnvValue = fmt.Sprintf("[Secret] %s/%s", env.ValueFrom.SecretKeyRef.Name, env.ValueFrom.SecretKeyRef.Key)
			}
			envs = append(envs, envHolder)

		}

		for _, envFrom := range container.EnvFrom {

			if envFrom.ConfigMapRef != nil && envFrom.ConfigMapRef.Name != "" {
				envs = append(envs,
					&Env{
						ContainerName: container.Name,
						EnvName:       envFrom.ConfigMapRef.Name,
						EnvValue:      fmt.Sprintf("[ConfigMap] %s", envFrom.ConfigMapRef.Name),
					},
				)
			}
			if envFrom.SecretRef != nil && envFrom.SecretRef.Name != "" {
				envs = append(envs,
					&Env{
						ContainerName: container.Name,
						EnvName:       envFrom.SecretRef.Name,
						EnvValue:      fmt.Sprintf("[Secret] %s", envFrom.SecretRef.Name),
					},
				)
			}
		}
	}
	return envs, nil
}

type SelectedNode struct {
	Reason  string `json:"reason,omitempty"`    // 选中类型，NodeSelector/NodeAffinity/Tolerations/NodeName
	Name    string `json:"node_name,omitempty"` // 节点名称
	Current bool   `json:"current,omitempty"`   // 是否是当前节点
}

// LinkedNode 可调度主机
// 暂不支持针对资源限定的cpu 内存主机筛选
func (p *pod) LinkedNode() ([]*SelectedNode, error) {

	var selectedNodeList []*SelectedNode

	var item *v1.Pod
	err := p.kubectl.Get(&item).Error
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, fmt.Errorf("")
	}
	var nodeList []*v1.Node
	err = p.kubectl.newInstance().Resource(&v1.Node{}).
		List(&nodeList).Error
	if err != nil {
		return nil, err
	}

	// 1. NodeSelector
	// 这个配置表示 Pod 只能调度到带有标签 disktype=ssd 的节点上。
	// NodeSelector中的标签，Node上必须全部满足
	if item.Spec.NodeSelector != nil {
		nodeList = slice.Filter(nodeList, func(index int, n *v1.Node) bool {
			labels := n.Labels

			if utils.CompareMapContains(item.Spec.NodeSelector, labels) {
				selectedNodeList = append(selectedNodeList, &SelectedNode{
					Reason: "NodeSelector",
					Name:   n.Name,
				})
				return true
			}
			return false
		})
	}

	// 2. nodeAffinity
	// requiredDuringSchedulingIgnoredDuringExecution
	if item.Spec.Affinity != nil && item.Spec.Affinity.NodeAffinity != nil && item.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution != nil {
		terms := item.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms
		for _, term := range terms {
			if term.MatchExpressions != nil && len(term.MatchExpressions) > 0 {

				for _, exp := range term.MatchExpressions {
					nodeList = slice.Filter(nodeList, func(index int, n *v1.Node) bool {
						labels := n.Labels

						if utils.MatchNodeSelectorRequirement(labels, exp) {
							for _, selectedNode := range selectedNodeList {
								if selectedNode.Name == n.Name {
									return false
								}
							}
							selectedNodeList = append(selectedNodeList, &SelectedNode{
								Reason: "NodeAffinity",
								Name:   n.Name,
							})
							return true
						}
						return false
					})
				}

			}
		}

	}

	// 污点 容忍度
	// 容忍度只要有一个满足即可。
	// 如果节点有污点，需要判断
	// 如果节点没有污点，不需要判断
	if item.Spec.Tolerations != nil && len(item.Spec.Tolerations) > 0 {

		nodeList = slice.Filter(nodeList, func(index int, n *v1.Node) bool {
			// 如果节点没有污点，不需要判断
			if n.Spec.Taints == nil || len(n.Spec.Taints) == 0 {
				return true
			}

			for _, t := range n.Spec.Taints {
				if isTaintTolerated(t, item.Spec.Tolerations) {
					for _, selectedNode := range selectedNodeList {
						if selectedNode.Name == n.Name {
							return false
						}
					}
					selectedNodeList = append(selectedNodeList, &SelectedNode{
						Reason: "Tolerations",
						Name:   n.Name,
					})
					return true
				}
			}
			return false
		})
	}
	// 最后 配置了nodeName，就只有这一个
	if item.Spec.NodeName != "" {
		nodeList = slice.Filter(nodeList, func(index int, n *v1.Node) bool {
			if n.Name == item.Spec.NodeName {

				// 看看之前有没有，如果有，就不再添加
				// 因为只要匹配上了，都要填到pod.spec.NodeName上
				// 无法区分是因为nodeSelector，还是nodeAffinity导致的调度成功
				for _, selectedNode := range selectedNodeList {
					if selectedNode.Name == item.Spec.NodeName {
						return false
					}
				}
				// 第一次
				selectedNodeList = append(selectedNodeList, &SelectedNode{
					Reason: "NodeName",
					Name:   n.Name,
				})
				return true
			}
			return false
		})

	}

	// 设置是否当前节点
	for _, selectedNode := range selectedNodeList {
		if selectedNode.Name == item.Spec.NodeName {
			selectedNode.Current = true
		}
	}

	return selectedNodeList, nil
}

// matchTaintAndToleration checks if a single taint matches a single toleration.
func matchTaintAndToleration(taint v1.Taint, toleration v1.Toleration) bool {
	// Check Effect (must match or be empty in toleration)
	if toleration.Effect != "" && toleration.Effect != taint.Effect {
		return false
	}

	// Check Operator and Key
	switch toleration.Operator {
	case "Equal":
		// Key and Value must match
		if toleration.Key != taint.Key {
			return false
		}
		if toleration.Value != taint.Value {
			return false
		}
	case "Exists":
		// Only Key needs to match
		if toleration.Key != "" && toleration.Key != taint.Key {
			return false
		}
	default:
		// Invalid operator
		return false
	}

	return true
}

// isTaintTolerated checks if a taint on a node is tolerated by any toleration of a pod.
func isTaintTolerated(taint v1.Taint, tolerations []v1.Toleration) bool {
	for _, toleration := range tolerations {
		if matchTaintAndToleration(taint, toleration) {
			return true
		}
	}
	return false
}
