package example

import (
	"fmt"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/duke-git/lancet/v2/slice"
	"github.com/weibaohui/kom/kom"
	"github.com/xwb1989/sqlparser"
	v1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/klog/v2"
)

// 定义过滤条件结构体
type Condition struct {
	Depth     int
	AndOr     string
	Field     string
	Operator  string
	Value     interface{} // 通过detectType 赋值为精确类型值，detectType之前都是string
	ValueType string      // number, string, bool, time
}

var conditions []Condition // 存储解析后的条件

func sqlTest() {
	sql := "select * from fake where id!=1 and `x.y.z.name`!='xxx' and (`yyy.uuu.ii.sage`>80 and sex=0) and (x='ttt' or yyy='ttt') order by id desc"
	sql = "select * from fake where (`metadata.namespace`='kube-system' or `metadata.namespace`='default' ) and  `metadata.name`!='hello-28890812-wj2dh' order by id desc limit 2"
	sql = "select * from fake where (`metadata.namespace`='kube-system' or `metadata.namespace`='default' ) and  `spec.replicas`>=1 order by id desc limit 2"
	sql = "select * from fake where (`metadata.namespace`='kube-system' or `metadata.namespace`='default' ) and  `spec.replicas`>=2 order by id desc limit 2"
	sql = "select * from fake where (`metadata.namespace`='kube-system' or `metadata.namespace`='default' ) And  `metadata.name` Like '%dns%' order by id desc limit 2"
	sql = "select * from fake where (`metadata.namespace`='kube-system' or `metadata.namespace`='default' ) and  `metadata.name` LIKE '%dns' order by id desc limit 2"
	sql = "select * from fake where (`metadata.namespace`='kube-system' or `metadata.namespace`='default' ) and  `metadata.name` LIKE 'dns%' order by id desc limit 2"
	sql = "select * from fake where (`metadata.namespace`='kube-system' or `metadata.namespace`='default' ) and  `spec.replicas`<>2 order by id desc limit 2"
	sql = "select * from fake where (`metadata.namespace`='kube-system' or `metadata.namespace`='default' ) and  `spec.replicas`!=2 order by id desc limit 2"
	sql = "select * from fake where (`metadata.namespace`='kube-system' or `metadata.namespace`='default' ) and  `spec.replicas`>1 order by id desc limit 2"
	sql = "select * from fake where (`metadata.namespace`='kube-system' or `metadata.namespace`='default' ) and  `spec.replicas`<1 order by id desc limit 2"
	sql = "select * from fake where (`metadata.namespace`='kube-system' or `metadata.namespace`='default' ) and  `spec.replicas`<3 order by id desc limit 2"
	sql = "select * from fake where (`metadata.namespace`='kube-system' or `metadata.namespace`='default' ) and  `spec.replicas`=3 order by id desc limit 2"
	sql = "select * from fake where (`metadata.namespace`='kube-system' or `metadata.namespace`='default' ) and  `spec.replicas`=2 order by id desc limit 2"
	sql = "select * from fake where (`metadata.namespace`='kube-system' or `metadata.namespace`='default' ) and  `spec.replicas`in (2,1) order by id desc limit 2"
	sql = "select * from fake where (`metadata.namespace`='kube-system' or `metadata.namespace`='default' ) and  `spec.replicas` not in (2,3) order by id desc limit 2"
	sql = "select * from fake where (`metadata.namespace`='kube-system' or `metadata.namespace`='default' ) and  `spec.replicas` between 3 and 4 order by id desc limit 2"
	sql = "select * from fake where (`metadata.namespace`='kube-system' or `metadata.namespace`='default' ) and  `spec.replicas` not between 3 and 4 order by id desc limit 2"
	sql = "select * from fake where (`metadata.namespace`='kube-system' or `metadata.namespace`='default' ) and  `spec.replicas` not between 1 and 2 order by id desc limit 2"
	sql = "select * from fake where (`metadata.namespace`='kube-system' or `metadata.namespace`='default' ) and  `spec.replicas` not between 2 and 3 order by id desc limit 2"
	sql = "select * from fake where (`metadata.namespace`='kube-system' or `metadata.namespace`='default' ) and  `metadata.creationTimestamp`  between '2024-11-09 00:00:00' and '2024-11-09 23:59:39' order by id desc limit 2"
	sql = "select * from fake where (`metadata.namespace`='kube-system' or `metadata.namespace`='default' ) and  `metadata.creationTimestamp` not between '2024-11-08' and '2024-11-10' order by id desc limit 2"
	sql = "select * from fake where (`metadata.namespace`='kube-system' or `metadata.namespace`='default' ) and  `metadata.creationTimestamp` > '2024-11-08' order by id desc limit 2"
	sql = "select * from fake where (`metadata.namespace`='kube-system' or `metadata.namespace`='default' ) and  `spec.replicas`in (2,1) order by id desc limit 2"
	sql = "select * from fake where (`metadata.namespace`='kube-system' or `metadata.namespace`='default' ) and  `metadata.creationTimestamp` in ('2024-11-08','2024-11-09','2024-11-10') order by id desc limit 2"

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

	// 探测 conditions中的条件值类型
	for i, cond := range conditions {
		conditions[i].ValueType, conditions[i].Value = detectType(cond.Value)
	}

	// 打印存储的条件列表
	for _, cond := range conditions {
		fmt.Printf("Depth: %d, AndOr: %s, Field: %s, Operator: %s, Value: %s\n",
			cond.Depth, cond.AndOr, cond.Field, cond.Operator, cond.Value)
	}
	var list []unstructured.Unstructured
	err = kom.DefaultCluster().Resource(&v1.Deployment{}).AllNamespace().List(&list).Error
	if err != nil {
		fmt.Printf("list error %v \n", err.Error())
	}
	result := executeFilter(list, conditions)
	fmt.Printf("count:%d\n", len(result))
	for _, r := range result {
		fmt.Printf("found %s/%s\n", r.GetNamespace(), r.GetName())
	}

}

// 定义字符串的类型
const (
	TypeNumber  = "number"
	TypeTime    = "time"
	TypeString  = "string"
	TypeBoolean = "boolean"
)

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
func trimQuotes(str string) string {
	str = strings.TrimPrefix(str, "`")
	str = strings.TrimSuffix(str, "`")
	str = strings.TrimPrefix(str, "'")
	str = strings.TrimSuffix(str, "'")
	return str
}

// 解析 WHERE 表达式
func parseWhereExpr(depth int, andor string, expr sqlparser.Expr) {
	klog.V(6).Infof("expr type %v,string %s", reflect.TypeOf(expr), sqlparser.String(expr))
	d := depth + 1 // 深度递增
	switch node := expr.(type) {
	case *sqlparser.ComparisonExpr:
		// 处理比较表达式 (比如 age > 80)
		cond := Condition{
			Depth:    depth,
			AndOr:    andor,
			Field:    trimQuotes(sqlparser.String(node.Left)),
			Operator: node.Operator,
			Value:    trimQuotes(sqlparser.String(node.Right)),
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
	case *sqlparser.RangeCond:
		// 递归解析 between 1 and 3 表达式
		cond := Condition{
			Depth:    depth,
			AndOr:    andor,
			Field:    trimQuotes(sqlparser.String(node.Left)),                                                                  // 左侧的字段
			Operator: node.Operator,                                                                                            // 操作符（BETWEEN）
			Value:    fmt.Sprintf("%s and %s", trimQuotes(sqlparser.String(node.From)), trimQuotes(sqlparser.String(node.To))), // 范围值
		}
		conditions = append(conditions, cond)
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
	klog.V(6).Infof("matchCondition  %s %s %s", condition.Field, condition.Operator, condition.Value)

	// 获取字段值
	fieldValue, found, err := getNestedFieldAsString(resource.Object, condition.Field)
	if err != nil || !found {
		klog.V(6).Infof("not found %s,%v", condition.Field, err)
		return false
	}

	// 判断字段类型并进行相应的比较
	switch condition.Operator {
	case "=":
		return compareValue(fieldValue, condition.Value)
	case "!=":
		return compareValue(fieldValue, condition.Value) == false
	case "like":
		return compareLike(fieldValue, condition.Value)
	case ">":
		return compareGreater(fieldValue, condition.Value)
	case "<":
		return compareLess(fieldValue, condition.Value)
	case ">=":
		return compareGreaterOrEqual(fieldValue, condition.Value)
	case "<=":
		return compareLessOrEqual(fieldValue, condition.Value)
	case "in":
		return compareIn(fieldValue, condition.Value)
	case "not in":
		return compareIn(fieldValue, condition.Value) == false
	case "between":
		return compareBetween(fieldValue, condition.Value)
	case "not between":
		return compareBetween(fieldValue, condition.Value) == false
	default:
		return false
	}
}

// compareValue 比较值是否相等
func compareValue(fieldValue string, value interface{}) bool {
	klog.V(6).Infof("compareValue (=) %s,%v(%v)", fieldValue, value, reflect.TypeOf(value))

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
func compareLike(fieldValue string, value interface{}) bool {
	klog.V(6).Infof("compareLike (like) %s,%v(%v)", fieldValue, value, reflect.TypeOf(value))

	targetValue := fmt.Sprintf("%v", value)

	// 提取值
	val := strings.TrimPrefix(targetValue, "%")
	val = strings.TrimSuffix(val, "%")

	// 判断是否包含%
	if strings.HasSuffix(targetValue, "%") && strings.HasPrefix(targetValue, "%") {
		// 以%开头，以%结尾，表示包含即可
		return strings.Contains(fieldValue, val)

	} else if strings.HasSuffix(targetValue, "%") {
		// abc%， 只以%结尾，表示开头必须是abc
		return strings.HasPrefix(fieldValue, val)
	} else if strings.HasPrefix(targetValue, "%") {
		// %abc 只以%开头，表示结尾必须是abc
		return strings.HasSuffix(fieldValue, val)
	} else {
		// 不包含%，表示必须相等
		return fieldValue == val
	}
}

// compareGreater 比较数值是否大于
func compareGreater(fieldValue string, value interface{}) bool {
	klog.V(6).Infof("compareGreater(>) %s,%v(%v)", fieldValue, value, reflect.TypeOf(value))
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
	case string:
		fieldValFloat, err := strconv.ParseFloat(fieldValue, 64)
		if err != nil {
			return false
		}
		valueFloat, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return false
		}
		return fieldValFloat > valueFloat
	case time.Time:
		fieldValTime, err := parseTime(fieldValue)
		if err != nil {
			return false
		}

		return fieldValTime.After(v)
	default:
		klog.V(6).Infof("%s,%v(%v)", fieldValue, value, reflect.TypeOf(value))
		return false
	}
}

// compareLess 比较数值是否小于
func compareLess(fieldValue string, value interface{}) bool {
	klog.V(6).Infof("compareLess(<) %s,%v(%v)", fieldValue, value, reflect.TypeOf(value))

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
	case string:
		fieldValFloat, err := strconv.ParseFloat(fieldValue, 64)
		if err != nil {
			return false
		}
		valueFloat, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return false
		}
		return fieldValFloat < valueFloat
	case time.Time:
		fieldValTime, err := parseTime(fieldValue)
		if err != nil {
			return false
		}
		return fieldValTime.Before(v)
	default:
		klog.V(6).Infof("%s,%v(%v)", fieldValue, value, reflect.TypeOf(value))
		return false
	}
}

// compareGreaterOrEqual 比较数值是否大于或等于
func compareGreaterOrEqual(fieldValue string, value interface{}) bool {
	klog.V(6).Infof("compareGreaterOrEqual(>=) %s,%v(%v)", fieldValue, value, reflect.TypeOf(value))
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
	case string:
		fieldValFloat, err := strconv.ParseFloat(fieldValue, 64)
		if err != nil {
			return false
		}
		valueFloat, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return false
		}
		return fieldValFloat >= valueFloat
	case time.Time:
		fieldValTime, err := parseTime(fieldValue)
		if err != nil {
			return false
		}
		return fieldValTime.After(v) || fieldValTime.Equal(v)
	default:
		klog.V(6).Infof("%s,%v(%v)", fieldValue, value, reflect.TypeOf(value))
		return false
	}

}

// compareLessOrEqual 比较数值是否小于或等于
func compareLessOrEqual(fieldValue string, value interface{}) bool {
	klog.V(6).Infof("compareLessOrEqual(<=) %s,%v(%v)", fieldValue, value, reflect.TypeOf(value))

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
	case string:
		fieldValFloat, err := strconv.ParseFloat(fieldValue, 64)
		if err != nil {
			return false
		}
		valueFloat, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return false
		}
		return fieldValFloat <= valueFloat
	case time.Time:
		fieldValTime, err := parseTime(fieldValue)
		if err != nil {
			return false
		}
		return fieldValTime.Before(v) || fieldValTime.Equal(v)
	default:
		klog.V(6).Infof("%s,%v(%v)", fieldValue, value, reflect.TypeOf(value))
		return false
	}
}

// compareIn 判断值是否在列表中
func compareIn(fieldValue string, value interface{}) bool {

	klog.V(6).Infof("compareIn(in []) %s,%v(%v)", fieldValue, value, reflect.TypeOf(value))

	// value 类型字符串 = (1,2,3,4) ，这个格式决定了只能是string类型
	// 如何判断fieldValue 是否在1,2,3,4范围内?
	if str, ok := value.(string); ok {
		// 去掉首尾的括号
		str = strings.TrimPrefix(str, "(")
		str = strings.TrimSuffix(str, ")")
		// 以逗号分割
		values := strings.Split(str, ",")
		for _, v := range values {
			v = strings.Trim(v, " ")
			v = trimQuotes(v)
			// 时间、字符串、数字
			// 只有相等，才能返回，因为in操作符，是or的关系。一个不行，需要判断下一个。

			// 先按数字比较
			fieldValueNum, err1 := strconv.ParseFloat(fieldValue, 64)
			toNum, err2 := strconv.ParseFloat(v, 64)
			if err1 == nil && err2 == nil {
				if fieldValueNum == toNum {
					return true
				}
			}

			// 时间不能简单判断，而要判断是否日期、小时、分钟，是否in。
			// 是否包含时间部分，如果包含，就是精确匹配。如果不不含，就是判断日期
			fieldValueTime, err1 := parseTime(fieldValue)
			toTime, err2 := parseTime(v)
			if err1 == nil && err2 == nil {

				// 判断目标时间字符串是否包含时间部分（即时分秒）
				if hasTimeComponent(v) {
					// 逐级比较时间分量（小时、分钟、秒）
					if fieldValueTime.Hour() == toTime.Hour() &&
						fieldValueTime.Minute() == toTime.Minute() &&
						fieldValueTime.Second() == toTime.Second() {
						return true
					}
				}
				// 比较日期部分（年、月、日）
				if isSameDate(fieldValueTime, toTime) {
					return true
				}
			}

			if fieldValue == v {
				return true
			}

		}
	}
	return false
}

// 判断是否包含时间部分
func hasTimeComponent(value string) bool {
	return strings.Contains(value, ":") // 如果字符串中包含冒号，说明有时间部分
}

// 判断两个时间的日期部分是否相同
func isSameDate(t1, t2 time.Time) bool {
	y1, m1, d1 := t1.Date()
	y2, m2, d2 := t2.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}

// compareBetween 判断值是否在范围内
func compareBetween(fieldValue string, value interface{}) bool {
	klog.V(6).Infof("compareBetween (between x and y) %s,%v(%v)", fieldValue, value, reflect.TypeOf(value))

	// value格式 举例: 1 and 5 这个格式决定了只能是string类型
	// 正则表达式匹配 "from AND to"
	var from, to string
	re := regexp.MustCompile(`(?i)(.+?)\s+AND\s+(.+)`)
	matches := re.FindStringSubmatch(fmt.Sprintf("%v", value))

	// 如果匹配成功，提取出 from 和 to
	if len(matches) == 3 {
		from = matches[1]
		to = matches[2]
	}

	// 判断 from to 是否为时间类型、数字类、还是字符串
	// 数字类型，要做 fieldValue 要转换为对应的类型，并进行>=from <=to的判断
	// 1. 尝试作为数字比较
	if fieldValNum, err := strconv.ParseFloat(fieldValue, 64); err == nil {
		fromNum, err1 := strconv.ParseFloat(from, 64)
		toNum, err2 := strconv.ParseFloat(to, 64)
		if err1 == nil && err2 == nil {
			klog.V(6).Infof("compareBetween(between x and y) as number %s,%v(%v)", fieldValue, value, reflect.TypeOf(value))
			return fieldValNum >= fromNum && fieldValNum <= toNum
		} else {
			klog.V(6).Infof("compareBetween(between x and y) as number  error %v %v", err1, err2)
		}

	}

	// 2. 尝试作为时间比较
	if fieldValTime, err := parseTime(fieldValue); err == nil {
		fromTime, err1 := parseTime(from)
		toTime, err2 := parseTime(to)
		if err1 == nil && err2 == nil {
			klog.V(6).Infof("compareBetween(between x and y) as date %s,%v(%v)", fieldValue, value, reflect.TypeOf(value))
			return (fieldValTime.Equal(fromTime) || fieldValTime.After(fromTime)) &&
				(fieldValTime.Equal(toTime) || fieldValTime.Before(toTime))
		} else {
			klog.V(6).Infof("compareBetween(between x and y) as date  error %v %v", err1, err2)
		}
	}

	// 3. 作为字符串比较
	return fieldValue >= from && fieldValue <= to
}
func parseTime(value string) (time.Time, error) {
	// 尝试不同的时间格式
	layouts := []string{
		time.RFC3339,          // "2006-01-02T15:04:05Z07:00"
		"2006-01-02",          // "2006-01-02"  (日期)
		"2006-01-02 15:04:05", // "2006-01-02 15:04:05" (无时区)
	}

	var t time.Time
	var err error
	for _, layout := range layouts {
		t, err = time.Parse(layout, value)
		if err == nil {
			return t, nil
		}
	}
	return t, err
}

// getNestedFieldAsString 获取嵌套字段值，返回字符串
func getNestedFieldAsString(obj map[string]interface{}, path string) (string, bool, error) {
	// todo 目前只能从一层一层的属性下取值，需要优化
	// 需要增加对数组下的取值判断，如pod container是数组，如何从数组中某一个项的属性中取值？
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
