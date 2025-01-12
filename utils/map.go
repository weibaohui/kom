package utils

import (
	"strconv"

	v1 "k8s.io/api/core/v1"
)

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

// MatchNodeSelectorRequirement checks if a map satisfies the NodeSelectorRequirement.
func MatchNodeSelectorRequirement(nodeLabels map[string]string, req v1.NodeSelectorRequirement) bool {
	value, exists := nodeLabels[req.Key]

	switch req.Operator {
	case v1.NodeSelectorOpIn:
		if !exists {
			return false
		}
		for _, v := range req.Values {
			if value == v {
				return true
			}
		}
		return false
	case v1.NodeSelectorOpNotIn:
		if !exists {
			return true
		}
		for _, v := range req.Values {
			if value == v {
				return false
			}
		}
		return true
	case v1.NodeSelectorOpExists:
		return exists
	case v1.NodeSelectorOpDoesNotExist:
		return !exists
	case v1.NodeSelectorOpGt, v1.NodeSelectorOpLt:
		if !exists || len(req.Values) != 1 {
			return false
		}
		threshold, err := strconv.Atoi(req.Values[0])
		if err != nil {
			return false
		}
		nodeValue, err := strconv.Atoi(value)
		if err != nil {
			return false
		}
		if req.Operator == v1.NodeSelectorOpGt {
			return nodeValue > threshold
		}
		return nodeValue < threshold
	default:
		return false
	}
}
