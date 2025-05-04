package callbacks

import (
	"bufio"
	"fmt"
	"io"

	"github.com/weibaohui/kom/kom"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/klog/v2"
)

// StreamExecuteCommand 在指定的 Kubernetes Pod 容器内以流式方式执行命令，并支持实时处理标准输入、输出和错误输出。
// 如果未提供 StreamOptions，则自动创建管道并通过回调函数逐行处理输出和错误信息。
// 若命令未设置或执行过程中发生错误，将返回相应的错误信息。
func StreamExecuteCommand(k *kom.Kubectl) error {

	stmt := k.Statement
	ns := stmt.Namespace
	name := stmt.Name
	command := stmt.Command
	args := stmt.Args
	containerName := stmt.ContainerName
	ctx := stmt.Context

	if stmt.StreamOptions != nil && stmt.StreamOptions.Stdin != nil {
		stmt.Stdin = stmt.StreamOptions.Stdin
	}
	// 只有一个容器时，可以不设置
	// if stmt.ContainerName == "" {
	// 	return fmt.Errorf("请调用ContainerName()方法设置Pod容器名称")
	// }
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

	//	如果设置 tty=true，但没有 stdin=true，可能导致命令执行失败或挂起
	//	某些命令如 bash、top 在 tty=false 下会拒绝运行或自动退出
	req.Param("tty", fmt.Sprintf("%v", stmt.Stdin != nil)).
		Param("stdin", fmt.Sprintf("%v", stmt.Stdin != nil)).
		Param("stdout", "true").
		Param("stderr", "true")

	executor, err := createExecutor(req.URL(), k.RestConfig())
	if err != nil {
		return fmt.Errorf("error creating executor: %v", err)
	}

	// 判断是否传递了remotecommand.StreamOptions
	if stmt.StreamOptions == nil {

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

		stmt.StreamOptions = options
	}

	// 开始流式执行
	err = executor.StreamWithContext(ctx, *stmt.StreamOptions)
	if err != nil {
		klog.V(8).Infof("Error Stream executing command: %v", err)
		return fmt.Errorf("error Stream executing command: %v", err)
	}

	return nil
}
