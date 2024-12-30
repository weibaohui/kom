package utils

import (
	"fmt"
	"net"
)

func CidrTotalIPs(cidr string) (int, error) {
	_, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return 0, fmt.Errorf("invalid CIDR: %v", err)
	}

	// 获取 CIDR 的掩码大小
	ones, bits := ipNet.Mask.Size()

	// 计算总 IP 数量
	totalIPs := 1 << (bits - ones)
	return totalIPs, nil
}
