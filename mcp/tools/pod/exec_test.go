package pod

import (
	"os"
	"strings"
	"testing"
)

// TestExecHandlerUsesByteSlice 回归保护：run_command_in_k8s_pod handler 必须用 *[]byte 作为
// kom.Cluster(...).Ctl().Pod().Execute() 的 dest。
//
// 历史上 ExecHandler 错误地使用 `var execResult string`，导致 Execute 的反射校验必失败，
// 错误信息："请确保dest 是一个指向字节切片的指针。定义var s []byte 使用&s"。
// kom/callbacks/exec.go 的反射契约要求 dest 是 *[]byte。
//
// 这里用源码字符串匹配锁定关键修复，避免回归。
func TestExecHandlerUsesByteSlice(t *testing.T) {
	src, err := os.ReadFile("exec.go")
	if err != nil {
		t.Fatalf("read exec.go: %v", err)
	}
	srcStr := string(src)
	if !strings.Contains(srcStr, "var execResult []byte") {
		t.Fatalf("exec.go must declare 'var execResult []byte' (see kom/callbacks/exec.go ExecuteCommand)")
	}
	if !strings.Contains(srcStr, "Execute(&execResult)") {
		t.Fatalf("exec.go must call Execute(&execResult)")
	}
}