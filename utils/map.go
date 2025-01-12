package utils

// CompareMapContains compares two maps and returns true if the small map is contained in the big map
// small Map中的所有键值都存在于big map中，则返回true
func CompareMapContains[T string | int](small, big map[string]T) bool {
	for key, valueA := range small {
		valueB, exists := big[key]
		if !exists || valueA != valueB {
			return false
		}
	}
	return true
}
