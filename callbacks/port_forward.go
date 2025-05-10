package callbacks

import (
	"fmt"
	"net/http"
	"os"

	"github.com/weibaohui/kom/kom"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
	"k8s.io/klog/v2"
)

// PortForward 建立本地端口到指定 Kubernetes Pod 端口的端口转发会话。
// 
// 如果未设置本地端口或 Pod 端口，则返回错误。端口转发会监听多个本地地址，并在连接建立后异步通知就绪状态。
// 
// 返回端口转发过程中的任何错误。
func PortForward(k *kom.Kubectl) error {

	stmt := k.Statement
	ns := stmt.Namespace
	name := stmt.Name
	stopCh := stmt.PortForwardStopCh
	podPort := stmt.PortForwardPodPort
	localPort := stmt.PortForwardLocalPort
	containerName := stmt.ContainerName

	// 检查端口，必须设置
	if localPort == "" || podPort == "" {
		return fmt.Errorf("localPort and podPort must be set")
	}

	if stopCh == nil {
		stopCh = make(chan struct{}, 1)
	}
	req := k.Client().CoreV1().RESTClient().
		Post().
		Namespace(ns).
		Name(name).
		Resource("pods").
		SubResource("portforward").
		Param("container", containerName)
	 
	readyChan := make(chan struct{})

	// 创建 PortForward 请求
	transport, upgrader, err := spdy.RoundTripperFor(k.RestConfig())
	if err != nil {
		return fmt.Errorf("failed to create round tripper: %w", err)
	}
	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, "POST", req.URL())

	addresses := []string{"localhost", "127.0.0.1", "0.0.0.0"}
	ports := []string{fmt.Sprintf("%s:%s", localPort, podPort)}
	forwarder, err := portforward.NewOnAddresses(dialer, addresses, ports, stopCh, readyChan, os.Stdout, os.Stderr)
	if err != nil {
		return fmt.Errorf("failed to create forwarder: %w", err)
	}

	go func() {
		<-readyChan
		klog.V(6).Infof("Port forwarding ready: 0.0.0.0:%s -> %s/%s:%s\n", localPort, ns, name, podPort)
	}()

	return forwarder.ForwardPorts()

}
