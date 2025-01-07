package callbacks

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
	"github.com/weibaohui/kom/utils"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/klog/v2"
)

// executeFilter 使用 lancet 执行过滤
func executeFilter(result []unstructured.Unstructured, conditions []kom.Condition) []unstructured.Unstructured {

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
		group := slice.Filter(conditions, func(index int, item kom.Condition) bool {
			return item.Depth == key
		})
		// 按组进行过滤，一般一组为相同的and or 条件
		result = evaluateCondition(result, group)
	}

	return result
}

func groupByDepth(conditions []kom.Condition) map[int][]kom.Condition {
	groups := make(map[int][]kom.Condition)
	for _, p := range conditions {
		depth := p.Depth
		groups[depth] = append(groups[depth], p)
	}
	return groups
}

func evaluateCondition(result []unstructured.Unstructured, group []kom.Condition) []unstructured.Unstructured {

	if group[0].AndOr == "OR" {
		return matchAny(result, group)
	} else {
		return matchAll(result, group)
	}
}

// matchAll 判断所有条件都满足 (AND 逻辑)
func matchAll(result []unstructured.Unstructured, conditions []kom.Condition) []unstructured.Unstructured {
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
func matchAny(result []unstructured.Unstructured, conditions []kom.Condition) []unstructured.Unstructured {
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
func matchCondition(resource unstructured.Unstructured, condition kom.Condition) bool {
	klog.V(8).Infof("matchCondition  %s %s %s", condition.Field, condition.Operator, condition.Value)

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
	klog.V(8).Infof("compareValue (=) %s,%v(%v)", fieldValue, value, reflect.TypeOf(value))

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
		fieldValTime, err := utils.ParseTime(fieldValue)
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
		fieldValTime, err := utils.ParseTime(fieldValue)
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
		fieldValTime, err := utils.ParseTime(fieldValue)
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
		fieldValTime, err := utils.ParseTime(fieldValue)
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
			v = utils.TrimQuotes(v)
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
			fieldValueTime, err1 := utils.ParseTime(fieldValue)
			toTime, err2 := utils.ParseTime(v)
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
	if fieldValTime, err := utils.ParseTime(fieldValue); err == nil {
		fromTime, err1 := utils.ParseTime(from)
		toTime, err2 := utils.ParseTime(to)
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
