package utils

import (
	"hash/fnv"
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
