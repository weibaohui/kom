package callbacks

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/duke-git/lancet/v2/slice"
	"github.com/duke-git/lancet/v2/stream"
	"github.com/weibaohui/kom/kom"
	"github.com/weibaohui/kom/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
)

func List(k *kom.Kubectl) error {

	// 前序步骤有任何Error及时终止
	if k.Error != nil {
		return k.Error
	}

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

	// 对List.Items进行过滤

	// 对结果进行过滤，执行where 条件
	result := executeFilter(list.Items, stmt.Filter.Conditions)

	if stmt.Filter.Order != "" {
		// 对结果执行OrderBy
		klog.V(6).Infof("order by = %s", stmt.Filter.Order)
		executeOrderBy(result, stmt.Filter.Order)
	} else {
		// 默认按创建时间倒序
		utils.SortByCreationTime(result)
	}

	// 先清空之前的值
	destValue.Elem().Set(reflect.MakeSlice(destValue.Elem().Type(), 0, 0))
	streamTmp := stream.FromSlice(result)
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

func executeOrderBy(result []unstructured.Unstructured, order string) {
	// order by `metadata.name` asc, `metadata.host` asc
	// todo 目前只实现了单一字段的排序，还没有搞定多个字段的排序
	order = strings.TrimPrefix(strings.TrimSpace(order), "order by")
	order = strings.TrimSpace(order)
	orders := strings.Split(order, ",")
	for _, ord := range orders {
		var field string
		var desc bool
		// 判断排序方向
		if strings.Contains(ord, "desc") {
			desc = true
			field = strings.ReplaceAll(ord, "desc", "")
		} else {
			field = strings.ReplaceAll(ord, "asc", "")
		}
		field = strings.TrimSpace(field)
		field = strings.TrimSpace(utils.TrimQuotes(field))
		klog.V(6).Infof("Sorting by field: %s, Desc: %v", field, desc)

		slice.SortBy(result, func(a, b unstructured.Unstructured) bool {
			// 获取字段值
			aFieldValue, found, err := getNestedFieldAsString(a.Object, field)
			if err != nil || !found {
				return false
			}
			bFieldValue, found, err := getNestedFieldAsString(b.Object, field)
			if err != nil || !found {
				return false
			}
			t, va := utils.DetectType(aFieldValue)
			_, vb := utils.DetectType(bFieldValue)

			switch t {
			case utils.TypeString:
				if desc {
					return va.(string) > vb.(string)
				}
				return va.(string) < vb.(string)
			case utils.TypeNumber:
				if desc {
					return va.(float64) > vb.(float64)
				}
				return va.(float64) < vb.(float64)
			case utils.TypeTime:
				tva, err := utils.ParseTime(fmt.Sprintf("%s", va))
				if err != nil {
					return false
				}
				tvb, err := utils.ParseTime(fmt.Sprintf("%s", vb))
				if err != nil {
					return false
				}
				if desc {
					return tva.After(tvb)
				}
				return tva.Before(tvb)
			default:
				return false
			}
		})

	}

}
