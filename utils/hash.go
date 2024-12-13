package utils

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
