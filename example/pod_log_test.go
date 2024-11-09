package example

import (
	"bufio"
	"io"
	"strings"
	"testing"

	"github.com/weibaohui/kom/kom"
	corev1 "k8s.io/api/core/v1"
)

func TestPodLogs(t *testing.T) {

	var stream io.ReadCloser
	err := kom.DefaultCluster().
		Namespace("default").
		Name("random").
		ContainerName("random").
		GetLogs(&stream, &corev1.PodLogOptions{}).Error
	if err != nil {
		t.Logf("Error getting pod logs:%v\n", err)
	}
	if stream == nil {
		return
	}
	reader := bufio.NewReader(stream)

	line, err := reader.ReadString('\n')
	if err != nil {
		if err == io.EOF {

		}
		t.Fatalf("Error reading stream: %v", err)
	}
	if !strings.Contains(line, "A") {
		t.Fatalf("日志读取测试失败,应该包含A。%s", line)
	}

}
