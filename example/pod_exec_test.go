package example

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/weibaohui/kom/kom"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/klog/v2"
)

func TestPodExec(t *testing.T) {
	var dest []byte
	err := kom.DefaultCluster().Namespace("default").
		Name("bash-runner-5bd4b4bdbb-8zpnp").
		Ctl().Pod().Command("ls", "-l").Execute(&dest).Error
	if err != nil {
		klog.Errorf("Error executing command: %v", err)
	}
	fmt.Printf("Standard Output:\n%s", dest)
}
func TestPodStreamExec(t *testing.T) {
	var dest []byte
	err := kom.DefaultCluster().Namespace("default").
		Name("bash-runner-5bd4b4bdbb-8zpnp").
		Ctl().Pod().Command("ls", "-l").StreamExecute(
		func(data []byte) error {
			klog.Infof("收到消息 = %s", string(data))
			return nil
		}, func(data []byte) error {
			klog.Infof("收到错误 = %s", string(data))
			return nil
		}).Error
	if err != nil {
		klog.Errorf("Error executing command: %v", err)
	}
	fmt.Printf("Standard Output:\n%s", dest)
}

type fakeTerminalSizeQueue struct {
	maxSizes      int
	terminalSizes []remotecommand.TerminalSize
}

// randomTerminalSize returns a TerminalSize with random values in the
// range (0-65535) for the fields Width and Height.
func randomTerminalSize() remotecommand.TerminalSize {
	randWidth := uint16(rand.Intn(int(math.Pow(2, 16))))
	randHeight := uint16(rand.Intn(int(math.Pow(2, 16))))
	return remotecommand.TerminalSize{
		Width:  randWidth,
		Height: randHeight,
	}
}

// Next returns a pointer to the next random TerminalSize, or nil if we have
// already returned "maxSizes" TerminalSizes already. Stores the randomly
// created TerminalSize in "terminalSizes" field for later validation.
func (f *fakeTerminalSizeQueue) Next() *remotecommand.TerminalSize {
	if len(f.terminalSizes) >= f.maxSizes {
		return nil
	}
	size := randomTerminalSize()
	f.terminalSizes = append(f.terminalSizes, size)
	return &size
}

func TestPodStreamExecWithOptions(t *testing.T) {
	var waiter sync.WaitGroup
	waiter.Add(1)

	inReader, inWriter := io.Pipe()
	sizeQueue := &fakeTerminalSizeQueue{}
	outBuffer := new(bytes.Buffer)
	errBuffer := new(bytes.Buffer)

	go func() {
		for i := 0; i < 3; i++ {
			sizeQueue.Next()
			time.Sleep(2 * time.Second)
		}
	}()
	go func() {
		for {
			if outBuffer.Len() > 0 {
				data := outBuffer.Bytes()
				outBuffer.Reset()
				klog.Infof("收到标准输出 (%d bytes): %q", len(data), string(data))
			}
			if errBuffer.Len() > 0 {
				data := errBuffer.Bytes()
				errBuffer.Reset()
				klog.Errorf("收到错误输出 (%d bytes): %q", len(data), string(data))
			}
		}

	}()
	go func() {
		// 在后台定时写入命令
		time.Sleep(3 * time.Second)
		for i := 0; i < 30; i++ {
			cmd := []byte("ls  \r")
			bytesWritten, err := inWriter.Write(cmd)
			if err != nil {
				klog.Errorf("Error writing to inWriter: %v", err)
				return
			}
			klog.V(6).Infof("Wrote %d bytes to inBuffer: %q", bytesWritten, string(cmd))
			time.Sleep(2 * time.Second)
		}
	}()
	opt := &remotecommand.StreamOptions{
		Stdin:             inReader,
		Stdout:            outBuffer,
		Stderr:            errBuffer,
		Tty:               true,
		TerminalSizeQueue: sizeQueue, // 传递 TTY 尺寸管理队列
	}

	kom.DefaultCluster().Namespace("default").
		Name("bash-runner-5bd4b4bdbb-qzgmx").
		Ctl().Pod().
		Command("/bin/sh", "-c", "TERM=xterm-256color; export TERM; [ -x /bin/bash ] && ([ -x /usr/bin/script ] && /usr/bin/script -q -c '/bin/bash' /dev/null || exec /bin/bash) || exec /bin/sh").
		StreamExecuteWithOptions(opt)

	waiter.Wait()
}
