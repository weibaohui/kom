package kom

import (
	"context"

	"github.com/dgraph-io/ristretto/v2"
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

// 获取一个全新的实例，只保留ctx
func (k *Kubectl) newInstance() *Kubectl {

	tx := &Kubectl{ID: k.ID, Error: k.Error}
	// clone with new statement
	tx.Statement = &Statement{
		Kubectl: k.Statement.Kubectl,
		Context: k.Statement.Context,
	}
	return tx

}

func (k *Kubectl) getInstance() *Kubectl {
	if k.clone > 0 {
		tx := &Kubectl{ID: k.ID, Error: k.Error}
		// clone with new statement
		tx.Statement = &Statement{
			Kubectl:      k.Statement.Kubectl,
			Context:      k.Statement.Context,
			ListOptions:  k.Statement.ListOptions,
			AllNamespace: k.Statement.AllNamespace,
			Namespace:    k.Statement.Namespace,
			Namespaced:   k.Statement.Namespaced,
			GVR:          k.Statement.GVR,
			GVK:          k.Statement.GVK,
			Name:         k.Statement.Name,
			CacheTTL:     k.Statement.CacheTTL,
			Filter:       k.Statement.Filter,
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
func (k *Kubectl) Cache() *ristretto.Cache[string, any] {
	cache := Clusters().GetClusterById(k.ID).Cache
	return cache
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

// Deprecated: use Ctl().Pod() instead.
func (k *Kubectl) Poder() *pod {
	return &pod{
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
