package kom

import (
	"context"

	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
)

type Kubectl struct {
	ID        string     // cluster id
	Statement *Statement // statement
	Error     error      // 存放ERROR信息

	clone int
}

// 初始化 kubectl
func initKubectl(config *rest.Config, id string) *Kubectl {
	klog.V(2).Infof("k8s init 服务器地址：%s\n", config.Host)

	k := &Kubectl{ID: id, clone: 1}

	k.Statement = &Statement{
		Context: context.Background(),
		Kubectl: k,
	}

	return k
}

func (k *Kubectl) getInstance() *Kubectl {
	if k.clone > 0 {
		tx := &Kubectl{ID: k.ID, Error: k.Error}
		// clone with new statement
		tx.Statement = &Statement{
			Kubectl:     k.Statement.Kubectl,
			Context:     k.Statement.Context,
			ListOptions: k.Statement.ListOptions,
			Namespace:   k.Statement.Namespace,
			Namespaced:  k.Statement.Namespaced,
			GVR:         k.Statement.GVR,
			GVK:         k.Statement.GVK,
			Name:        k.Statement.Name,
		}
		return tx
	}

	return k
}
func (k *Kubectl) Callback() *callbacks {
	cluster := Clusters().GetClusterById(k.ID)
	return cluster.callbacks
}
func (k *Kubectl) RestConfig() *rest.Config {
	cluster := Clusters().GetClusterById(k.ID)
	return cluster.Config
}
func (k *Kubectl) Client() *kubernetes.Clientset {
	cluster := Clusters().GetClusterById(k.ID)
	return cluster.Client
}
func (k *Kubectl) DynamicClient() *dynamic.DynamicClient {
	cluster := Clusters().GetClusterById(k.ID)
	return cluster.DynamicClient
}
func (k *Kubectl) parentCluster() *clusterInst {
	cluster := Clusters().GetClusterById(k.ID)
	return cluster
}
func (k *Kubectl) Applier() *applier {
	return &applier{
		kubectl: k,
	}
}
func (k *Kubectl) Poder() *poder {
	// 执行前做必要检查
	if k.Statement.GVR.Empty() {
		k.Resource(&v1.Pod{})
	}
	return &poder{
		kubectl: k,
	}
}
func (k *Kubectl) Status() *status {
	return &status{
		kubectl: k,
	}
}
func (k *Kubectl) Tools() *tools {
	return &tools{
		kubectl: k,
	}
}

func (k *Kubectl) Ctl() *ctl {
	return &ctl{
		kubectl: k,
	}
}
