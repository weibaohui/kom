package utils

import (
	"hash/fnv"
	"crypto/md5"
	"encoding/hex"
)

const (
	offset32 = 2166136261
	prime32  = 16777619
)

func FNV1_32(data []byte) uint32 {
	hash := uint32(offset32)
	for _, b := range data {
		hash *= prime32
		hash ^= uint32(b)
	}
	return hash
}
func FNV1(data []byte) uint32 {
	h := fnv.New32a()
	h.Write(data)
	return h.Sum32()
}
// MD5Hash 计算输入字符串的MD5值并返回16进制字符串
func MD5Hash(s string) string {
	h := md5.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}
