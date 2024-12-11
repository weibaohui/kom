package utils

import (
	"fmt"
	"strconv"
)

// 定义字符串的类型
const (
	TypeNumber  = "number"
	TypeTime    = "time"
	TypeString  = "string"
	TypeBoolean = "boolean"
)

// DetectType 探测字符串的类型（数字、时间、字符串）
func DetectType(value interface{}) (string, interface{}) {

	if boolean, err := strconv.ParseBool(fmt.Sprintf("%v", value)); err == nil {
		return TypeBoolean, boolean
	}

	// 1. 尝试解析为整数或浮点数
	if num, err := strconv.ParseFloat(fmt.Sprintf("%v", value), 64); err == nil {
		return TypeNumber, num
	}

	if t, err := ParseTime(fmt.Sprintf("%v", value)); err == nil {
		return TypeTime, t
	}

	// 3. 默认返回字符串类型
	return TypeString, value
}
