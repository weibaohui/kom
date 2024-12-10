package kom

import (
	"fmt"
	"strings"
)

// Sql TODO 解析sql为函数调用，实现支持原生sql语句
// select * from pod where pod.name='?', 'abc'
func (k *Kubectl) Sql(sql string, values ...interface{}) *Kubectl {
	tx := k.getInstance()
	return tx
}

func (k *Kubectl) From(group string, version string, kind string) *Kubectl {
	tx := k.getInstance()
	tx.GVK(group, version, kind)
	return tx
}
func (k *Kubectl) Where(condition string, values ...interface{}) *Kubectl {
	tx := k.getInstance()

	// 解析条件并替换占位符 "?"
	labelConditions := []string{}
	fieldConditions := []string{}

	// 将条件按 AND 分割
	parts := strings.Split(condition, "AND")
	if len(parts) != len(values) {
		fmt.Println("Error: condition and values count mismatch")
		return tx
	}

	// 替换 "?" 并分类
	for i, part := range parts {
		part = strings.TrimSpace(part)
		// 替换 "?" 占位符
		conditionWithValue := strings.Replace(part, "?", fmt.Sprintf("%v", values[i]), 1)

		// 判断是 LabelSelector 还是 FieldSelector
		if key, value := parseConditionKeyValue(conditionWithValue); isFieldSelectorKey(key) {
			fieldConditions = append(fieldConditions, fmt.Sprintf("%s=%s", key, value))
		} else {
			labelConditions = append(labelConditions, fmt.Sprintf("%s=%s", key, value))
		}
	}

	// 组合最终的 selector 字符串
	if len(labelConditions) > 0 {
		tx.WithLabelSelector(strings.Join(labelConditions, ","))
	}
	if len(fieldConditions) > 0 {
		tx.WithFieldSelector(strings.Join(fieldConditions, ","))
	}

	return tx
}

// parseConditionKeyValue 将 "key=value" 解析成 key 和 value
func parseConditionKeyValue(condition string) (string, string) {
	parts := strings.SplitN(condition, "=", 2)
	if len(parts) != 2 {
		return "", ""
	}
	return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
}

// isFieldSelectorKey 判断 key 是否是 FieldSelector 支持的字段
// todo 相对有限，收集齐全，拓展
func isFieldSelectorKey(key string) bool {
	fieldKeys := map[string]bool{
		"metadata.name":      true,
		"metadata.namespace": true,
		"status.phase":       true,
		"spec.nodeName":      true,
	}
	return fieldKeys[key]
}

func (k *Kubectl) Order(order string) *Kubectl {
	tx := k.getInstance()
	tx.Statement.Filter.Order = order
	return tx
}
func (k *Kubectl) Limit(limit int) *Kubectl {
	tx := k.getInstance()
	tx.Statement.Filter.Limit = limit
	return tx
}
func (k *Kubectl) Offset(offset int) *Kubectl {
	tx := k.getInstance()
	tx.Statement.Filter.Offset = offset
	return tx
}

// Skip AliasFor Offset
func (k *Kubectl) Skip(skip int) *Kubectl {
	return k.Offset(skip)
}
