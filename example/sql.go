package example

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/duke-git/lancet/v2/slice"
	"github.com/weibaohui/kom/kom"
	"github.com/xwb1989/sqlparser"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// 定义过滤条件结构体
type Condition struct {
	Depth    int
	AndOr    string
	Field    string
	Operator string
	Value    string
}

var conditions []Condition // 存储解析后的条件

func sqlTest() {
	sql := "select * from fake where id!=1 and `x.y.z.name`!='xxx' and (`yyy.uuu.ii.sage`>80 and sex=0) and (x='ttt' or yyy='ttt') order by id desc"
	sql = "select * from fake where `metadata.namespace`='kube-system' order by id desc"
	fmt.Println("SQL:", sql)

	stmt, err := sqlparser.Parse(sql)
	if err != nil {
		fmt.Println("Error parsing SQL:", err)
		return
	}

	// 解析 SQL 中的 WHERE 子句
	switch stmt := stmt.(type) {
	case *sqlparser.Select:
		parseWhereExpr(0, "AND", stmt.Where.Expr)
	}

	// 打印存储的条件列表
	for _, cond := range conditions {
		fmt.Printf("Depth: %d, AndOr: %s, Field: %s, Operator: %s, Value: %s\n",
			cond.Depth, cond.AndOr, cond.Field, cond.Operator, cond.Value)
	}
	var list []unstructured.Unstructured
	err = kom.DefaultCluster().Resource(&v1.Pod{}).AllNamespace().List(&list).Error
	if err != nil {
		fmt.Printf("list error %v \n", err.Error())
	}
	result := executeFilter(list, conditions)
	for _, r := range result {
		fmt.Printf("fount %s/%s\n", r.GetNamespace(), r.GetName())
	}

}

func groupByDepth(conditions []Condition) map[int][]Condition {
	groups := make(map[int][]Condition)
	for _, p := range conditions {
		depth := p.Depth
		groups[depth] = append(groups[depth], p)
	}
	return groups
}

// executeFilter 使用 lancet 执行过滤
func executeFilter(result []unstructured.Unstructured, conditions []Condition) []unstructured.Unstructured {

	// 最终的结果，按照条件列表，逐一执行过滤

	// 2. 按 depth 分组
	groupedConditions := groupByDepth(conditions)

	// 3. 从小到大逐层过滤
	// 3.1. 提取所有的 key
	keys := make([]int, 0, len(groupedConditions))
	for k := range groupedConditions {
		keys = append(keys, k)
	}

	// 3.2. 对 keys 进行排序,从小到大
	sort.Ints(keys)
	slice.SortBy(keys, func(a, b int) bool {
		return a < b
	})
	// 3. 按排序后的 keys 顺序逐个访问 map 并处理
	for _, key := range keys {
		// 这个key 就是depth 的distinct值
		group := slice.Filter(conditions, func(index int, item Condition) bool {
			return item.Depth == key
		})
		// 按组进行过滤，一般一组为相同的and or 条件
		result = evaluateCondition(result, group)
	}

	return result
}
func trim(str string) string {
	str = strings.TrimPrefix(str, "`")
	str = strings.TrimSuffix(str, "`")
	str = strings.TrimPrefix(str, "'")
	str = strings.TrimSuffix(str, "'")
	return str
}

// 解析 WHERE 表达式
func parseWhereExpr(depth int, andor string, expr sqlparser.Expr) {
	d := depth + 1 // 深度递增
	switch node := expr.(type) {
	case *sqlparser.ComparisonExpr:
		// 处理比较表达式 (比如 age > 80)
		cond := Condition{
			Depth:    depth,
			AndOr:    andor,
			Field:    trim(sqlparser.String(node.Left)),
			Operator: node.Operator,
			Value:    trim(sqlparser.String(node.Right)),
		}
		conditions = append(conditions, cond)
	case *sqlparser.ParenExpr:
		// 处理括号表达式
		// 括号内的表达式是一个独立的子表达式，增加深度
		parseWhereExpr(d+1, "()", node.Expr)

	case *sqlparser.AndExpr:
		// 递归解析 AND 表达式
		// 这里传递 "AND" 给左右两边
		parseWhereExpr(d, "AND", node.Left)
		parseWhereExpr(d, "AND", node.Right)

	case *sqlparser.OrExpr:
		// 递归解析 OR 表达式
		// 这里传递 "OR" 给左右两边
		parseWhereExpr(d, "OR", node.Left)
		parseWhereExpr(d, "OR", node.Right)

	default:
		// 其他表达式
		fmt.Printf("Unhandled expression at depth %d: %s\n", depth, sqlparser.String(expr))
	}
}

// /
// // val, found, err := getNestedFieldAsString(item.Object, "spec.strategy.rollingUpdate.maxSurge")
//		// val, found, err := getNestedFieldAsString(item.Object, "spec.template.metadata.labels.app")
//		val, found, err := getNestedFieldAsString(item.Object, "spec.template.spec.restartPolicy")
//		if err != nil {
//			fmt.Printf("getNestedFieldAsString spec.dnsPolicy error :%v\n", err)
//		}
//		if found {
//			fmt.Printf("spec.dnsPolicy found :%v\n", val)
//		}
//		klog.Errorf("val=%s, found= %v, err=%v \n", val, found, err)
//
//

func evaluateCondition(result []unstructured.Unstructured, group []Condition) []unstructured.Unstructured {

	if group[0].AndOr == "OR" {
		return matchAny(result, group)
	} else {
		return matchAll(result, group)

	}
}

// matchAll 判断所有条件都满足 (AND 逻辑)
func matchAll(result []unstructured.Unstructured, conditions []Condition) []unstructured.Unstructured {
	return slice.Filter(result, func(index int, item unstructured.Unstructured) bool {
		// 遍历所有条件，只有全部条件成立才返回 true
		for _, c := range conditions {
			if !matchCondition(item, c) {
				return false // 只要有一个条件不成立，直接返回 false
			}
		}
		return true // 所有条件都成立，返回 true
	})
}

// matchAny 判断任一条件满足 (OR 逻辑)
func matchAny(result []unstructured.Unstructured, conditions []Condition) []unstructured.Unstructured {
	return slice.Filter(result, func(index int, item unstructured.Unstructured) bool {
		// 遍历所有条件，任意一个条件成立就返回 true
		for _, c := range conditions {
			if matchCondition(item, c) {
				return true
			}
		}
		return false
	})
}

// matchCondition 判断单个条件是否匹配
func matchCondition(resource unstructured.Unstructured, condition Condition) bool {
	// 获取字段值
	fieldValue, found, err := getNestedFieldAsString(resource.Object, condition.Field)
	if err != nil || !found {
		return false
	}

	// 判断字段类型并进行相应的比较
	switch condition.Operator {
	case "=":
		return compareValue(fieldValue, condition.Value)
	case "!=":
		return compareValue(fieldValue, condition.Value) == false
	case "LIKE":
		return compareLike(fieldValue, condition.Value)
	case ">":
		return compareGreater(fieldValue, condition.Value)
	case "<":
		return compareLess(fieldValue, condition.Value)
	case ">=":
		return compareGreaterOrEqual(fieldValue, condition.Value)
	case "<=":
		return compareLessOrEqual(fieldValue, condition.Value)
	case "BETWEEN":
		return compareBetween(fieldValue, condition.Value)
	default:
		return false
	}
}

// compareValue 比较值是否相等
func compareValue(fieldValue string, value interface{}) bool {
	switch v := value.(type) {
	case string:
		return fieldValue == v
	case float64, int, int64:
		fieldValFloat, err := strconv.ParseFloat(fieldValue, 64)
		if err != nil {
			return false
		}
		return fieldValFloat == v
	case time.Time:
		// 需要将 fieldValue 转换为 time 类型进行比较
		fieldTime, err := time.Parse(time.RFC3339, fieldValue)
		if err != nil {
			return false
		}
		return fieldTime.Equal(v)
	default:
		return false
	}
}

// compareLike 判断字符串是否匹配
func compareLike(fieldValue string, pattern string) bool {
	// 使用正则表达式来判断 LIKE
	regexPattern := strings.ReplaceAll(pattern, "%", ".*")
	matched, _ := regexp.MatchString("^"+regexPattern+"$", fieldValue)
	return matched
}

// compareGreater 比较数值是否大于
func compareGreater(fieldValue string, value interface{}) bool {
	switch v := value.(type) {
	case float64:
		fieldValFloat, err := strconv.ParseFloat(fieldValue, 64)
		if err != nil {
			return false
		}
		return fieldValFloat > v
	case int, int64:
		fieldValFloat, err := strconv.ParseFloat(fieldValue, 64)
		if err != nil {
			return false
		}
		return fieldValFloat > float64(v.(int))
	default:
		return false
	}
}

// compareLess 比较数值是否小于
func compareLess(fieldValue string, value interface{}) bool {
	switch v := value.(type) {
	case float64:
		fieldValFloat, err := strconv.ParseFloat(fieldValue, 64)
		if err != nil {
			return false
		}
		return fieldValFloat < v
	case int, int64:
		fieldValFloat, err := strconv.ParseFloat(fieldValue, 64)
		if err != nil {
			return false
		}
		return fieldValFloat < float64(v.(int))
	default:
		return false
	}
}

// compareGreaterOrEqual 比较数值是否大于或等于
func compareGreaterOrEqual(fieldValue string, value interface{}) bool {
	switch v := value.(type) {
	case float64:
		fieldValFloat, err := strconv.ParseFloat(fieldValue, 64)
		if err != nil {
			return false
		}
		return fieldValFloat >= v
	case int, int64:
		fieldValFloat, err := strconv.ParseFloat(fieldValue, 64)
		if err != nil {
			return false
		}
		return fieldValFloat >= float64(v.(int))
	default:
		return false
	}
}

// compareLessOrEqual 比较数值是否小于或等于
func compareLessOrEqual(fieldValue string, value interface{}) bool {
	switch v := value.(type) {
	case float64:
		fieldValFloat, err := strconv.ParseFloat(fieldValue, 64)
		if err != nil {
			return false
		}
		return fieldValFloat <= v
	case int, int64:
		fieldValFloat, err := strconv.ParseFloat(fieldValue, 64)
		if err != nil {
			return false
		}
		return fieldValFloat <= float64(v.(int))
	default:
		return false
	}
}

// compareBetween 判断值是否在范围内
func compareBetween(fieldValue string, value interface{}) bool {
	if rangeVal, ok := value.([]float64); ok && len(rangeVal) == 2 {
		start, end := rangeVal[0], rangeVal[1]
		fieldValFloat, err := strconv.ParseFloat(fieldValue, 64)
		if err != nil {
			return false
		}
		return fieldValFloat >= start && fieldValFloat <= end
	}
	return false
}

// compareTime 比较时间类型
func compareTime(fieldValue string, value interface{}) bool {
	switch v := value.(type) {
	case time.Time:
		// 需要将 fieldValue 转换为时间进行比较
		fieldTime, err := time.Parse(time.RFC3339, fieldValue)
		if err != nil {
			return false
		}
		return fieldTime.Equal(v)
	default:
		return false
	}
}

// getNestedFieldAsString 获取嵌套字段值，返回字符串
func getNestedFieldAsString(obj map[string]interface{}, path string) (string, bool, error) {
	fields := strings.Split(path, ".")
	value, found, err := unstructured.NestedFieldCopy(obj, fields...)
	if err != nil || !found {
		return "", false, err
	}

	switch v := value.(type) {
	case string:
		return v, true, nil
	case bool, int, int64, float64:
		return fmt.Sprintf("%v", v), true, nil
	default:
		return "", false, nil
	}
}
