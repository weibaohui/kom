package kom

import (
	"fmt"
	"reflect"

	"github.com/weibaohui/kom/utils"
	"github.com/xwb1989/sqlparser"
	"k8s.io/klog/v2"
)

// 解析 WHERE 表达式
func parseWhereExpr(conditions []*Condition, depth int, andor string, expr sqlparser.Expr) []*Condition {
	klog.V(6).Infof("expr type [%v],string %s, type [%s]", reflect.TypeOf(expr), sqlparser.String(expr), andor)
	d := depth + 1 // 深度递增
	switch node := expr.(type) {
	case *sqlparser.ComparisonExpr:
		// 处理比较表达式 (比如 age > 80)
		cond := &Condition{
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
		conditions = parseWhereExpr(conditions, d+1, "AND", node.Expr)

	case *sqlparser.AndExpr:
		// 递归解析 AND 表达式
		// 这里传递 "AND" 给左右两边
		conditions = parseWhereExpr(conditions, d, "AND", node.Left)
		conditions = parseWhereExpr(conditions, d, "AND", node.Right)
	case *sqlparser.RangeCond:
		// 递归解析 between 1 and 3 表达式
		cond := &Condition{
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
