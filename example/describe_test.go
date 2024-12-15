package example

import (
	"fmt"
	"testing"

	"github.com/weibaohui/kom/kom"
	"github.com/weibaohui/kom/kom/describe"
)

func TestDescribePod(t *testing.T) {
	// 3. 获取 Pod 描述信息
	namespace := "default"
	podName := "nginx-label-5c5444cd8b-pfrzg"

	client := kom.DefaultCluster().Client()
	podDescriber := &describe.PodDescriber{Interface: client}

	output, err := podDescriber.Describe(namespace, podName, describe.DescriberSettings{})
	if err != nil {
		panic(err)
	}

	// 4. 打印描述结果
	fmt.Println(output)
	t.Logf("\n%s\n", output)
}
