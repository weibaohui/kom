package kom

import (
	"context"

	"github.com/weibaohui/kom/kom/docer"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
)

type Kom struct {
	ID        string
	Statement *Statement
	Error     error

	clone int
}

func Init() *Kom {
	return Clusters().Default().Kom
}
func Cluster(id string) *Kom {
	return Clusters().GetById(id).Kom
}

func InitConnectionByConfig(config *rest.Config, id string) *Kom {
	klog.V(2).Infof("k8s Client InitConnectionByConfig")
	klog.V(2).Infof("服务器地址：%s\n", config.Host)

	kom := &Kom{ID: id, clone: 1}

	// 注册回调参数
	kom.Statement = &Statement{
		Context: context.Background(),
		Kom:     kom,
	}

	return kom
}

func (k *Kom) getInstance() *Kom {
	if k.clone > 0 {
		tx := &Kom{ID: k.ID, Error: k.Error}
		// clone with new statement
		tx.Statement = &Statement{
			Kom:         k.Statement.Kom,
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
func (k *Kom) Callback() *callbacks {
	cluster := Clusters().GetById(k.ID)
	return cluster.callbacks
}
func (k *Kom) RestConfig() *rest.Config {
	cluster := Clusters().GetById(k.ID)
	return cluster.Config
}
func (k *Kom) Client() *kubernetes.Clientset {
	cluster := Clusters().GetById(k.ID)
	return cluster.Client
}
func (k *Kom) DynamicClient() *dynamic.DynamicClient {
	cluster := Clusters().GetById(k.ID)
	return cluster.DynamicClient
}
func (k *Kom) APIResources() []*metav1.APIResource {
	cluster := Clusters().GetById(k.ID)
	return cluster.apiResources
}
func (k *Kom) CRDList() []*unstructured.Unstructured {
	cluster := Clusters().GetById(k.ID)
	return cluster.crdList
}
func (k *Kom) Docs() *docer.Docs {
	cluster := Clusters().GetById(k.ID)
	return cluster.Docs
}
