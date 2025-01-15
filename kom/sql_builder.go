package kom

import (
	"fmt"
	"log"
	"strings"

	"github.com/weibaohui/kom/utils"
	"github.com/xwb1989/sqlparser"
	"k8s.io/klog/v2"
)

// Sql TODO Insert Update Delete
//
//	 已支持Select
//		解析sql为函数调用，实现支持原生sql语句
//
// select * from pod where pod.name='?', 'abc'
func (k *Kubectl) Sql(sql string, values ...interface{}) *Kubectl {
	tx := k.getInstance()
	tx.AllNamespace()

	sql = formatSql(sql, values)

	// 添加反引号，将metadata.name 转为`metadata.name`,
	// k8s中很多类似json的字段，需要用反引号进行包裹，避免被作为db.table形式使用
	// sql = NewSqlParse(sql).AddBackticks()

	stmt, err := sqlparser.Parse(sql)
	if err != nil {
		klog.Errorf("Error parsing SQL:%s,%v", sql, err)
		tx.Error = err
		return tx
	}

	var conditions []Condition // 存储解析后的条件

	// 断言为 *sqlparser.Select 类型
	selectStmt, ok := stmt.(*sqlparser.Select)
	if !ok {
		log.Fatalf("Not a SELECT statement")
	}
	// 获取 Select 语句中的 From 作为Resource
	from := sqlparser.String(selectStmt.From)
	gvk := k.Tools().FindGVKByTableNameInApiResources(from)
	if gvk == nil {
		tx.Error = fmt.Errorf("resource %s not found both in api-resource and crd", from)
		klog.V(6).Infof("resource %s not found both in api-resource and crd", from)
		names := k.Tools().ListAvailableTableNames()
		klog.V(6).Infof("Available resource: %s", names)
		return tx
	}

	// 设置GVK
	tx.GVK(gvk.Group, gvk.Version, gvk.Kind)

	// 获取 LIMIT 子句信息
	limit := selectStmt.Limit
	if limit != nil {
		// 获取 LIMIT 的 Rowcount 和 Offset
		rowCount := sqlparser.String(limit.Rowcount)
		offset := sqlparser.String(limit.Offset)

		tx.Limit(utils.ToInt(rowCount))
		tx.Offset(utils.ToInt(offset))
	}
	// 解析Where语句，活的执行条件
	conditions = parseWhereExpr(conditions, 0, "AND", selectStmt.Where.Expr)

	// 探测 conditions中的条件值类型
	for i, cond := range conditions {
		conditions[i].ValueType, conditions[i].Value = utils.DetectType(cond.Value)
	}
	tx.Statement.Filter.Conditions = conditions

	// 设置排序字段
	orderBy := selectStmt.OrderBy
	if orderBy != nil {
		tx.Statement.Filter.Order = sqlparser.String(orderBy)
	}

	tx.Statement.Filter.Parsed = true
	return tx
}

func (k *Kubectl) From(tableName string) *Kubectl {
	tx := k.getInstance()
	gvk := k.Tools().FindGVKByTableNameInApiResources(tableName)
	if gvk == nil {
		tx.Error = fmt.Errorf("resource %s not found both in api-resource and crd", tableName)
		klog.V(6).Infof("resource %s not found both in api-resource and crd", tableName)
		names := k.Tools().ListAvailableTableNames()
		klog.V(6).Infof("Available resource: %s", names)
		return tx
	}
	tx.Statement.Filter.From = tableName
	// 设置GVK
	tx.GVK(gvk.Group, gvk.Version, gvk.Kind)
	return tx
}
func (k *Kubectl) Where(condition string, values ...interface{}) *Kubectl {
	tx := k.getInstance()
	originalSql := tx.Statement.Filter.Sql
	sql := formatSql(condition, values)

	if originalSql != "" {
		sql = originalSql + " and ( " + sql + " ) "
	} else {
		sql = fmt.Sprintf(" select * from fake where ( %s )", sql)
	}

	// 添加反引号，将metadata.name 转为`metadata.name`,
	// k8s中很多类似json的字段，需要用反引号进行包裹，避免被作为db.table形式使用
	// sql = NewSqlParse(sql).AddBackticks()

	tx.Statement.Filter.Sql = sql

	stmt, err := sqlparser.Parse(sql)
	if err != nil {
		klog.Errorf("Error parsing SQL:%s,%v", sql, err)
		tx.Error = err
		return tx
	}

	var conditions []Condition // 存储解析后的条件

	// 断言为 *sqlparser.Select 类型
	selectStmt, ok := stmt.(*sqlparser.Select)
	if !ok {
		klog.Errorf("not select parsing SQL:%s,%v", sql, err)
		tx.Error = err
		return tx
	}

	// 解析Where语句，获得执行条件
	conditions = parseWhereExpr(conditions, 0, "AND", selectStmt.Where.Expr)

	// 探测 conditions中的条件值类型
	for i, cond := range conditions {
		conditions[i].ValueType, conditions[i].Value = utils.DetectType(cond.Value)
	}

	tx.Statement.Filter.Conditions = conditions

	tx.Statement.Filter.Parsed = true

	return tx
}

// formatSql
//
//	select * from pod where pod.name='?', 'abc'
func formatSql(condition string, values []interface{}) string {
	// 将 values 替换到 condition 中的占位符 ?
	for _, value := range values {
		// 对值进行安全格式化，例如字符串加单引号
		switch v := value.(type) {
		case string:
			condition = strings.Replace(condition, "?", fmt.Sprintf("'%s'", v), 1)
		default:
			condition = strings.Replace(condition, "?", fmt.Sprintf("%v", v), 1)
		}
	}
	return condition
}

// Order
// Order(" id desc")
// Order(" date asc")
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
