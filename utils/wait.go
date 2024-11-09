package utils

import (
	"time"

	"k8s.io/klog/v2"
)

// WaitUntil 传入一个函数 f，如果 f 返回 true 则停止等待，否则持续检查直到超时
// Example:
// // 定义一个检查函数，模拟每 10 秒满足一次条件
//
//	checkCondition := func() bool {
//		// 例如条件为当前秒数是 10 的倍数
//		return time.Now().Second()%10 == 0
//	}
//
//	// 每 2 秒检查一次，超时设定为 30 秒
//	interval := 2 * time.Second
//	timeout := 30 * time.Second
//
//	if waitUntil(checkCondition, interval, timeout) {
//		fmt.Println("Check succeeded, main process exiting.")
//	} else {
//		fmt.Println("Check failed due to timeout.")
//	}
func WaitUntil(f func() bool, interval time.Duration, timeout time.Duration) bool {
	timeoutTimer := time.After(timeout)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-timeoutTimer:
			klog.V(4).Infof("Timeout reached, stopping monitoring.")
			return false // 超时返回 false
		case <-ticker.C:
			if f() {
				klog.V(4).Infof("Condition met, stopping monitoring.")
				return true // 如果 f 返回 true 则停止
			}
			klog.V(2).Infof("Condition not met, retrying...")
		}
	}
}
