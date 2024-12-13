package kom

import (
	"testing"
)

// TestAddBackticks 测试 AddBackticks 方法
func TestAddBackticks(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// 基本字段测试
		{
			name:     "Basic field",
			input:    "SELECT name FROM pod WHERE namespace='kube-system'",
			expected: "SELECT `name` FROM pod WHERE `namespace`='kube-system'",
		},
		// 多级路径字段测试
		{
			name:     "Nested fields",
			input:    "SELECT metadata.namespace.xy.mn FROM pod WHERE metadata.creationTimestamp>'2024-01-01'",
			expected: "SELECT `metadata.namespace.xy.mn` FROM pod WHERE `metadata.creationTimestamp`>'2024-01-01'",
		},
		// 运算符测试
		{
			name:     "Operators",
			input:    "SELECT * FROM pod WHERE status.phase!='Running' AND spec.replicas BETWEEN 2 AND 5",
			expected: "SELECT * FROM pod WHERE `status.phase`!='Running' AND `spec.replicas` BETWEEN 2 AND 5",
		},
		// 跳过关键字
		{
			name:     "Skip SQL keywords",
			input:    "SELECT * FROM pod WHERE status.phase IS NULL OR spec.replicas NOT IN (1, 2, 3)",
			expected: "SELECT * FROM pod WHERE `status.phase` IS NULL OR `spec.replicas` NOT IN (1, 2, 3)",
		},
		// 字符串和数字值测试
		{
			name:     "String and numeric values",
			input:    "SELECT * FROM pod WHERE spec.replicas=3 AND metadata.namespace='default'",
			expected: "SELECT * FROM pod WHERE `spec.replicas`=3 AND `metadata.namespace`='default'",
		},
		// ORDER BY 测试
		{
			name:     "Order By clause",
			input:    "SELECT * FROM pod ORDER BY metadata.creationTimestamp DESC",
			expected: "SELECT * FROM pod ORDER BY `metadata.creationTimestamp` DESC",
		},
		// 复杂条件组合测试
		{
			name:     "Complex conditions",
			input:    "SELECT * FROM pod WHERE metadata.namespace='kube-system' AND status.phase='Running' OR spec.replicas>=2",
			expected: "SELECT * FROM pod WHERE `metadata.namespace`='kube-system' AND `status.phase`='Running' OR `spec.replicas`>=2",
		},
		// IS NOT NULL 测试
		{
			name:     "IS NOT NULL",
			input:    "SELECT * FROM pod WHERE status.phase IS NOT NULL",
			expected: "SELECT * FROM pod WHERE `status.phase` IS NOT NULL",
		},
		// 特殊运算符 LIKE 测试
		{
			name:     "LIKE operator",
			input:    "SELECT * FROM pod WHERE metadata.name LIKE 'nginx%'",
			expected: "SELECT * FROM pod WHERE `metadata.name` LIKE 'nginx%'",
		},
		// 特殊符号测试
		{
			name:     "Special symbols",
			input:    "SELECT * FROM pod WHERE metadata.namespace = 'default' AND metadata.labels.env='prod'",
			expected: "SELECT * FROM pod WHERE `metadata.namespace` = 'default' AND `metadata.labels.env`='prod'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewSqlParse(tt.input).AddBackticks()
			if result != tt.expected {
				t.Errorf("AddBackticks() failed: \ninput:    %s\nexpected: %s\ngot:      %s", tt.input, tt.expected, result)
			}
		})
	}
}
