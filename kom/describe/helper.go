package describe

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/rest"
)

func InitializeDescriberMap(config *rest.Config) map[schema.GroupKind]ResourceDescriber {
	m, _ := describerMap(config)
	return m
}
