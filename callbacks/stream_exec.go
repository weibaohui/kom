package callbacks

import (
	"bufio"
	"fmt"
	"io"

	"github.com/weibaohui/kom/kom"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/klog/v2"
)

func StreamExecuteCommand(k *kom.Kubectl) error {

	stmt := k.Statement
	ns := stmt.Namespace
	name := stmt.Name
	command := stmt.Command
	args := stmt.Args
	containerName := stmt.ContainerName
	ctx := stmt.Context

	if stmt.ContainerName == "" {
		return fmt.Errorf("请调用ContainerName()方法设置Pod容器名称")
	}
	if stmt.Command == "" {
		return fmt.Errorf("请调用Command()方法设置命令")
	}

	var err error
	cmd := []string{command}
	cmd = append(cmd, args...)
	klog.V(8).Infof("Stream Execute %s %v in [%s/%s:%s]\n", command, args, ns, name, containerName)

	req := k.Client().CoreV1().RESTClient().
		Post().
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
		Param("stdin", fmt.Sprintf("%v", stmt.Stdin != nil)).
		Param("stdout", "true").
		Param("stderr", "true")

	executor, err := createExecutor(req.URL(), k.RestConfig())
	if err != nil {
		return fmt.Errorf("error creating executor: %v", err)
	}

	// 使用 io.Pipe 实现 Stdout 和 Stderr 的实时流式处理
	stdoutPr, stdoutPw := io.Pipe()
	stderrPr, stderrPw := io.Pipe()
	defer stdoutPr.Close()
	defer stderrPr.Close()
	defer stdoutPw.Close()
	defer stderrPw.Close()

	// Goroutine 处理 Stdout
	go func() {
		scanner := bufio.NewScanner(stdoutPr)
		for scanner.Scan() {
			line := scanner.Text()
			if stmt.StdoutCallback != nil {
				if err := stmt.StdoutCallback([]byte(line)); err != nil {
					klog.Errorf("StreamExecuteCommand Error in  Stdout callback: %v", err)
				}
			}
		}
		if err := scanner.Err(); err != nil {
			klog.Errorf("StreamExecuteCommand Error reading from Stdout pipe: %v", err)
		}
	}()

	// Goroutine 处理 Stderr
	go func() {
		scanner := bufio.NewScanner(stderrPr)
		for scanner.Scan() {
			line := scanner.Text()
			if stmt.StderrCallback != nil {
				if err := stmt.StderrCallback([]byte(line)); err != nil {
					klog.Errorf("StreamExecuteCommand Error in Stderr callback: %v", err)
				}
			}
		}
		if err := scanner.Err(); err != nil {
			klog.Errorf("StreamExecuteCommand Error reading from Stderr pipe: %v", err)
		}
	}()

	options := &remotecommand.StreamOptions{
		Stdout: stdoutPw, // 将输出写入 Stdout 的 PipeWriter
		Stderr: stderrPw, // 将错误写入 Stderr 的 PipeWriter
		Tty:    false,
	}
	if stmt.Stdin != nil {
		options.Stdin = stmt.Stdin
	}

	// 开始流式执行
	err = executor.StreamWithContext(ctx, *options)
	if err != nil {
		klog.V(8).Infof("Error Stream executing command: %v", err)
		return fmt.Errorf("error Stream executing command: %v", err)
	}

	return nil
}
