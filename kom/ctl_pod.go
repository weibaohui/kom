package kom

import (
	"io"

	v1 "k8s.io/api/core/v1"
)

type pod struct {
	kubectl *Kubectl
	Error   error
}

func (p *pod) ContainerName(c string) *pod {
	tx := p.kubectl.getInstance()
	tx.Statement.ContainerName = c
	p.kubectl = tx
	return p
}

func (p *pod) Command(command string, args ...string) *pod {
	tx := p.kubectl.getInstance()
	tx.Statement.Command = command
	tx.Statement.Args = args
	p.kubectl = tx
	return p
}
func (p *pod) Execute(dest interface{}) *pod {
	tx := p.kubectl.getInstance()
	tx.Statement.Dest = dest
	tx.Error = tx.Callback().Exec().Execute(tx)
	p.Error = tx.Error
	return p
}

func (p *pod) StreamExecute(stdout, stderr func(data []byte) error) *Kubectl {
	tx := p.kubectl.getInstance()
	tx.Statement.StdoutCallback = stdout
	tx.Statement.StderrCallback = stderr
	tx.Error = tx.Callback().StreamExec().Execute(tx)
	p.Error = tx.Error
	return tx
}
func (p *pod) Stdin(reader io.Reader) *pod {
	tx := p.kubectl.getInstance()
	tx.Statement.Stdin = reader
	return p
}
func (p *pod) GetLogs(requestPtr interface{}, opt *v1.PodLogOptions) *pod {
	tx := p.kubectl.getInstance()
	// 如果只有一个容器，是可以不设置containerName的
	// if tx.Statement.ContainerName == "" {
	// 	p.Error = fmt.Errorf("请先设置ContainerName")
	// 	return p
	// }
	tx.Statement.PodLogOptions = opt
	tx.Statement.PodLogOptions.Container = tx.Statement.ContainerName
	tx.Statement.Dest = requestPtr
	tx.Error = tx.Callback().Logs().Execute(tx)
	p.Error = tx.Error
	return p
}
