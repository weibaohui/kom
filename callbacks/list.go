package callbacks

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/duke-git/lancet/v2/stream"
	"github.com/weibaohui/kom/kom"
	"github.com/weibaohui/kom/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

func List(k *kom.Kubectl) error {

	stmt := k.Statement
	gvr := stmt.GVR
	namespaced := stmt.Namespaced
	ns := stmt.Namespace
	ctx := stmt.Context

	opts := stmt.ListOptions
	listOptions := metav1.ListOptions{}
	if len(opts) > 0 {
		listOptions = opts[0]
	}

	// 使用反射获取 dest 的值
	destValue := reflect.ValueOf(stmt.Dest)

	// 确保 dest 是一个指向切片的指针
	if destValue.Kind() != reflect.Ptr || destValue.Elem().Kind() != reflect.Slice {
		// 处理错误：dest 不是指向切片的指针
		return fmt.Errorf("请传入数组类型")
	}
	// 获取切片的元素类型
	elemType := destValue.Elem().Type().Elem()

	var list *unstructured.UnstructuredList
	var err error

	if namespaced {

		if stmt.AllNamespace {
			ns = metav1.NamespaceAll
		} else {
			if ns == "" {
				ns = "default"
			}
		}
		list, err = stmt.Kubectl.DynamicClient().Resource(gvr).Namespace(ns).List(ctx, listOptions)
	} else {
		list, err = stmt.Kubectl.DynamicClient().Resource(gvr).List(ctx, listOptions)
	}
	if err != nil {
		return err
	}

	utils.SortByCreationTime(list.Items)
	// 先清空之前的值
	destValue.Elem().Set(reflect.MakeSlice(destValue.Elem().Type(), 0, 0))
	streamTmp := stream.FromSlice(list.Items)
	// 查看是否有filter ，先使用filter 形成一个最终的list.Items
	if stmt.Filter.Offset > 0 {
		streamTmp = streamTmp.Skip(stmt.Filter.Offset)
	}
	if stmt.Filter.Limit > 0 {
		streamTmp = streamTmp.Limit(stmt.Filter.Limit)
	}

	for _, item := range streamTmp.ToSlice() {

		obj := item.DeepCopy()
		if stmt.RemoveManagedFields {
			utils.RemoveManagedFields(obj)
		}
		// 创建新的指向元素类型的指针
		newElemPtr := reflect.New(elemType)
		// unstructured 转换为原始目标类型
		err = runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Object, newElemPtr.Interface())
		// 将指针的值添加到切片中
		destValue.Elem().Set(reflect.Append(destValue.Elem(), newElemPtr.Elem()))

	}
	stmt.RowsAffected = int64(len(list.Items))

	if err != nil {
		return err
	}
	return nil
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

// Condition 表示单个查询条件
type Condition struct {
	Key      string      // 字段路径，例如 "spec.replicas"
	Operator string      // 操作符，例如 "=", ">", "<", "BETWEEN"
	Value    interface{} // 目标值（支持数值类型和时间类型）
}

// Query 表示查询节点，支持逻辑组合和嵌套
type Query struct {
	Logic      string        // 逻辑关系: "AND" 或 "OR"
	Conditions []interface{} // 支持 Condition 或 Query 类型
}

// FilterResources 根据复杂查询条件过滤 unstructured 资源列表
func FilterResources(resources []unstructured.Unstructured, query Query) []unstructured.Unstructured {
	var filtered []unstructured.Unstructured

	for _, resource := range resources {
		if matchQuery(resource, query) {
			filtered = append(filtered, resource)
		}
	}
	return filtered
}

// matchQuery 递归判断资源是否满足查询条件树
func matchQuery(resource unstructured.Unstructured, query Query) bool {
	if query.Logic == "AND" {
		return matchAll(resource, query.Conditions)
	} else if query.Logic == "OR" {
		return matchAny(resource, query.Conditions)
	}
	return false
}

// matchAll 判断所有条件都满足 (AND 逻辑)
func matchAll(resource unstructured.Unstructured, conditions []interface{}) bool {
	for _, c := range conditions {
		switch v := c.(type) {
		case Condition:
			if !matchCondition(resource, v) {
				return false
			}
		case Query:
			if !matchQuery(resource, v) {
				return false
			}
		}
	}
	return true
}

// matchAny 判断任一条件满足 (OR 逻辑)
func matchAny(resource unstructured.Unstructured, conditions []interface{}) bool {
	for _, c := range conditions {
		switch v := c.(type) {
		case Condition:
			if matchCondition(resource, v) {
				return true
			}
		case Query:
			if matchQuery(resource, v) {
				return true
			}
		}
	}
	return false
}

// matchCondition 判断单个条件是否匹配
func matchCondition(resource unstructured.Unstructured, condition Condition) bool {
	// 获取字段值
	fieldValue, found, err := getNestedFieldAsString(resource.Object, condition.Key)
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
		return compareLike(fieldValue, condition.Value.(string))
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
