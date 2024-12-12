package kom

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/weibaohui/kom/utils"
	"github.com/xwb1989/sqlparser"
	"k8s.io/klog/v2"
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

// addBackticksToColumns 遍历 AST 并为列名加反引号
func addBackticksToColumns(node sqlparser.SQLNode) {
	// 遍历 AST 的节点
	sqlparser.Walk(func(node sqlparser.SQLNode) (kontinue bool, err error) {
		// 处理列名
		if col, ok := node.(*sqlparser.ColName); ok {
			// 为列名加反引号（如果尚未存在）
			klog.V(4).Infof("addBackticksToColumns: %s.%s[%v]\n", col.Qualifier, col.Name, col.Metadata)
			col.Name = sqlparser.NewColIdent(fmt.Sprintf("`%s`", strings.Trim(col.Name.String(), "`")))
		}
		return true, nil
	}, node)
}

// AddBackticks 给 SQL 中的字段名加反引号，保留多级路径
func AddBackticks(sql string) string {
	// SQL 保留关键字
	keywords := map[string]bool{
		"and": true, "or": true, "not": true, "in": true, "like": true, "is": true,
		"null": true, "true": true, "false": true, "between": true, "exists": true,
		"case": true, "when": true, "then": true, "else": true, "end": true,
	}

	// 运算符正则（识别字段名和值之间的运算符）
	operators := `=|!=|<>|<=|>=|<|>|like|in|not\s+in|is\s+null|is\s+not\s+null|between|like`

	// 正则表达式：匹配字段名 运算符 值 的形式
	re := regexp.MustCompile(`\b([a-zA-Z_][a-zA-Z0-9_.]*)\b\s*(` + operators + `)\s*([^,()\s]+|'.*?'|".*?")`)

	// 替换逻辑
	result := re.ReplaceAllStringFunc(sql, func(match string) string {
		// 分组匹配：字段名、运算符、右侧值
		parts := re.FindStringSubmatch(match)
		if len(parts) < 4 {
			return match // 不完整的匹配，跳过
		}

		left := parts[1]     // 左侧字段名
		operator := parts[2] // 运算符
		right := parts[3]    // 右侧值

		// 检查字段名是否为关键字
		if keywords[strings.ToLower(left)] {
			return match // 如果字段名是关键字，则不加反引号
		}

		// 给字段名整体加反引号
		if !strings.HasPrefix(left, "`") && !strings.HasSuffix(left, "`") {
			left = fmt.Sprintf("`%s`", left)
		}

		// 重新组合
		return fmt.Sprintf("%s %s %s", left, operator, right)
	})

	return result
}
