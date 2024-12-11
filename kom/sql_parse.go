package kom

import (
	"fmt"
	"reflect"
	"strconv"
	"time"

	"github.com/weibaohui/kom/utils"
	"github.com/xwb1989/sqlparser"
	"k8s.io/klog/v2"
)

// 定义字符串的类型
const (
	TypeNumber  = "number"
	TypeTime    = "time"
	TypeString  = "string"
	TypeBoolean = "boolean"
)

// 解析 WHERE 表达式
func parseWhereExpr(conditions []Condition, depth int, andor string, expr sqlparser.Expr) []Condition {
	klog.V(6).Infof("expr type %v,string %s", reflect.TypeOf(expr), sqlparser.String(expr))
	d := depth + 1 // 深度递增
	switch node := expr.(type) {
	case *sqlparser.ComparisonExpr:
		// 处理比较表达式 (比如 age > 80)
		cond := Condition{
			Depth:    depth,
			AndOr:    andor,
			Field:    utils.TrimQuotes(sqlparser.String(node.Left)),
			Operator: node.Operator,
			Value:    utils.TrimQuotes(sqlparser.String(node.Right)),
		}
		conditions = append(conditions, cond)
	case *sqlparser.ParenExpr:
		// 处理括号表达式
		// 括号内的表达式是一个独立的子表达式，增加深度
		conditions = parseWhereExpr(conditions, d+1, "()", node.Expr)

	case *sqlparser.AndExpr:
		// 递归解析 AND 表达式
		// 这里传递 "AND" 给左右两边
		conditions = parseWhereExpr(conditions, d, "AND", node.Left)
		conditions = parseWhereExpr(conditions, d, "AND", node.Right)
	case *sqlparser.RangeCond:
		// 递归解析 between 1 and 3 表达式
		cond := Condition{
			Depth:    depth,
			AndOr:    andor,
			Field:    utils.TrimQuotes(sqlparser.String(node.Left)),                                                                        // 左侧的字段
			Operator: node.Operator,                                                                                                        // 操作符（BETWEEN）
			Value:    fmt.Sprintf("%s and %s", utils.TrimQuotes(sqlparser.String(node.From)), utils.TrimQuotes(sqlparser.String(node.To))), // 范围值
		}
		conditions = append(conditions, cond)
	case *sqlparser.OrExpr:
		// 递归解析 OR 表达式
		// 这里传递 "OR" 给左右两边
		conditions = parseWhereExpr(conditions, d, "OR", node.Left)
		conditions = parseWhereExpr(conditions, d, "OR", node.Right)

	default:
		// 其他表达式
		fmt.Printf("Unhandled expression at depth %d: %s\n", depth, sqlparser.String(expr))
	}
	return conditions
}

// detectType 探测字符串的类型（数字、时间、字符串）
func detectType(value interface{}) (string, interface{}) {

	if boolean, err := strconv.ParseBool(fmt.Sprintf("%v", value)); err == nil {
		return TypeBoolean, boolean
	}

	// 1. 尝试解析为整数或浮点数
	if num, err := strconv.ParseFloat(fmt.Sprintf("%v", value), 64); err == nil {
		return TypeNumber, num
	}

	// 2. 尝试解析为时间
	timeLayouts := []string{
		"2006-01-02",                // 日期格式
		"2006-01-02T15:04:05Z07:00", // RFC3339 格式
		"2006-01-02 15:04:05",       // 无时区的时间格式
	}

	for _, layout := range timeLayouts {
		if t, err := time.Parse(layout, fmt.Sprintf("%v", value)); err == nil {
			return TypeTime, t
		}
	}

	// 3. 默认返回字符串类型
	return TypeString, value
}
