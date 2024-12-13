package kom

import (
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/antlr4-go/antlr/v4"
	"github.com/weibaohui/kom/kom/parser"
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

// EnterColumn_name 处理字段名
// 不要重命名，因为是一个方法的实现
func (s *sqlParser) EnterColumn_name(ctx *parser.Column_nameContext) {
	// 这里可以处理 SELECT 语句的其他部分，如果需要
	field := ctx.GetText()
	// 将识别出来的字段保存下来
	s.Fields = append(s.Fields, field)
}

type sqlParser struct {
	*parser.BaseSQLiteParserListener
	Sql    string
	Fields []string
}

func NewSqlParse(sql string) *sqlParser {
	return &sqlParser{
		Sql:    sql,
		Fields: []string{},
	}
}

// AddBackticks 给 SQL 中的字段名加反引号，保留多级路径
// 添加反引号，将metadata.name 转为`metadata.name`,
// k8s中很多类似json的字段，需要用反引号进行包裹，避免被作为db.table形式使用
func (s *sqlParser) AddBackticks() string {
	klog.V(6).Infof("sql before add backticks [ %s ]", s.Sql)
	// 创建输入流
	input := antlr.NewInputStream(s.Sql)

	// 创建词法分析器和解析器
	lexer := parser.NewSQLiteLexer(input)
	stream := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)
	p := parser.NewSQLiteParser(stream)

	// 解析 SQL，遍历解析树
	antlr.ParseTreeWalkerDefault.Walk(s, p.Sql_stmt())

	// 按照字段长度从长到短排序
	sort.Slice(s.Fields, func(i, j int) bool {
		return len(s.Fields[i]) > len(s.Fields[j])
	})
	// 逐个替换字段
	// 创建一个占位符的映射
	fieldPlaceholders := make(map[string]string)
	for i, field := range s.Fields {
		// 为每个字段分配一个唯一的占位符
		placeholder := fmt.Sprintf("__FIELD%d__", i+1)
		fieldPlaceholders[field] = placeholder
		// 将 SQL 中的字段替换成占位符
		s.Sql = strings.ReplaceAll(s.Sql, field, placeholder)
	}
	klog.V(6).Infof("sql temp add backticks [ %s ]", s.Sql)

	// 替换占位符为带反引号的字段
	for field, placeholder := range fieldPlaceholders {
		quotedField := "`" + field + "`"
		s.Sql = strings.ReplaceAll(s.Sql, placeholder, quotedField)
	}
	klog.V(6).Infof("sql after add backticks [ %s ]", s.Sql)
	return s.Sql
}
