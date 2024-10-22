package kom

import (
	"context"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
)

var (
	kom *Kom
)
var apiResources []metav1.APIResource   // 当前k8s已注册资源
var crdList []unstructured.Unstructured // 当前k8s已注册资源

type Kom struct {
	Client        *kubernetes.Clientset
	Statement     *Statement
	Error         error
	config        *rest.Config
	dynamicClient dynamic.Interface
	callbacks     *callbacks
	clone         int
}

func Init() *Kom {
	return kom
}

// InitConnection 在主入口处进行初始化
func InitConnection(path string) *Kom {
	klog.V(2).Infof("k8s Client init")
	kom = &Kom{clone: 1}

	config, err := getKubeConfig(path)
	if err != nil {
		panic(err.Error())
	}
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	dynClient, err := dynamic.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	kom.Client = client
	kom.config = config
	kom.dynamicClient = dynClient

	// 注册回调参数
	kom.Statement = &Statement{
		Context:       context.Background(),
		Client:        client,
		DynamicClient: dynClient,
		config:        config,
	}

	kom.callbacks = initializeCallbacks(kom)

	// 提取ApiResources
	_, lists, _ := kom.Client.Discovery().ServerGroupsAndResources()
	for _, list := range lists {
		resources := list.APIResources
		version := list.GroupVersionKind().Version
		group := list.GroupVersionKind().Group
		groupVersion := list.GroupVersion
		gvs := strings.Split(groupVersion, "/")
		if len(gvs) == 2 {
			group = gvs[0]
			version = gvs[1]
		} else {
			// 只有version的情况"v1"
			version = groupVersion
		}

		for _, resource := range resources {
			resource.Group = group
			resource.Version = version
			apiResources = append(apiResources, resource)
		}
	}

	// 提取crdList
	crdList, _ = kom.ListResources(context.TODO(), "CustomResourceDefinition", "")
	return kom
}

func getKubeConfig(path string) (*rest.Config, error) {
	config, err := rest.InClusterConfig()

	if err != nil {
		klog.V(2).Infof("尝试读取集群内访问配置：%v\n", err)
		klog.V(2).Infof("尝试读取本地配置%s", path)
		// 不是在集群中,读取参数配置
		config, err = clientcmd.BuildConfigFromFlags("", path)
		if err != nil {
			klog.Errorf(err.Error())
		}

	}
	if config != nil {
		klog.V(2).Infof("服务器地址：%s\n", config.Host)
	}
	return config, err
}

func (kom *Kom) getInstance() *Kom {
	if kom.clone > 0 {
		tx := &Kom{Error: kom.Error}

		if kom.clone == 1 {
			// clone with new statement
			tx.Statement = &Statement{
				Context:       kom.Statement.Context,
				Client:        kom.Statement.Client,
				DynamicClient: kom.Statement.DynamicClient,
				config:        kom.Statement.config,
				ListOptions:   kom.Statement.ListOptions,
				Namespace:     kom.Statement.Namespace,
				Namespaced:    kom.Statement.Namespaced,
				GVR:           kom.Statement.GVR,
				GVK:           kom.Statement.GVK,
				Name:          kom.Statement.Name,
			}
			tx.callbacks = kom.callbacks
			tx.Client = kom.Client
		} else {
			// with clone statement
			tx.Statement = kom.Statement.clone()
			tx.callbacks = kom.callbacks
		}

		return tx
	}

	return kom
}
func (kom *Kom) Callback() *callbacks {
	return kom.callbacks
}
func (kom *Kom) RestConfig() *rest.Config {
	return kom.config
}
