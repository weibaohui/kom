package tools

// ResourceMetadata 封装资源的元数据信息
type ResourceMetadata struct {
	Cluster   string
	Namespace string
	Name      string
	Group     string
	Version   string
	Kind      string
}

type ResourceInfo struct {
	Group      string
	Version    string
	Kind       string
	Namespaced bool
}
