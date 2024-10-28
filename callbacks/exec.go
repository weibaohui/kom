package callbacks

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/weibaohui/kom/kom"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/klog/v2"
)

func ExecuteCommand(k *kom.Kubectl) error {

	stmt := k.Statement
	ns := stmt.Namespace
	name := stmt.Name
	command := stmt.Command
	args := stmt.Args
	containerName := stmt.ContainerName

	var err error
	cmd := []string{command}
	cmd = append(cmd, args...)
	klog.V(8).Infof("ExecuteCommand %s %v in [%s/%s:%s]\n", command, args, ns, name, containerName)

	req := k.Client().CoreV1().RESTClient().
		Get().
		Namespace(ns).
		Resource("pods").
		Name(name).
		SubResource("exec").
		Param("container", containerName).
		Param("command", cmd[0])

	for _, arg := range cmd[1:] {
		req.Param("command", arg)
	}

	req.Param("tty", "false").
		Param("stdin", "false").
		Param("stdout", "true").
		Param("stderr", "true")

	executor, err := remotecommand.NewSPDYExecutor(k.RestConfig(), "POST", req.URL())
	if err != nil {
		return fmt.Errorf("error creating executor: %v", err)
	}

	var outBuf bytes.Buffer
	var errBuf bytes.Buffer
	err = executor.Stream(remotecommand.StreamOptions{
		Stdout: &outBuf,
		Stderr: &errBuf,
	})

	if err != nil {
		s := errBuf.String()
		if strings.Contains(s, "Invalid argument") {
			return fmt.Errorf("系统参数错误 %v", s)
		}
		return fmt.Errorf("error executing command: %v %v", err, s)
	}

	// 将结果写入 tx.Statement.Dest
	if destStr, ok := k.Statement.Dest.(*string); ok {
		*destStr = outBuf.String()
		klog.V(8).Infof("ExecuteCommand result %s", *destStr)
	} else {
		return fmt.Errorf("dest is not a *string")
	}
	return nil
}
