package hunter

import (
	"strings"
)

// processSimpleQuery 处理简单查询语句（单个键值对）
func processSimpleQuery(query string) string {
	// 分析查询语句中的键值对
	parts := strings.SplitN(query, "=", 2)
	if len(parts) == 2 {
		key := parts[0]
		value := parts[1]
		
		// 如果值不包含引号，但可能原本应该有引号
		if !strings.HasPrefix(value, "'") && !strings.HasPrefix(value, "\"") &&
		   !strings.HasSuffix(value, "'") && !strings.HasSuffix(value, "\"") {
			// 将值用双引号包围
			return key + "=\"" + value + "\""
		}
	}
	
	return query
}

// processComplexQuery 处理复杂查询语句（包含逻辑运算符）
func processComplexQuery(query string) string {
	// 先将查询语句按逻辑运算符分割
	// 先替换 && 为临时标记，以便处理嵌套的 ||
	query = strings.ReplaceAll(query, "&&", "__AND__")
	// 分割 ||
	orParts := strings.Split(query, "||") 
	
	// 处理每个 OR 部分
	for i, orPart := range orParts {
		// 恢复 AND 标记
		andParts := strings.Split(orPart, "__AND__")
		
		// 处理每个 AND 部分
		for j, andPart := range andParts {
			// 处理单个条件
			andParts[j] = processSimpleQuery(strings.TrimSpace(andPart))
		}
		
		// 重新组合 AND 部分
		orParts[i] = strings.Join(andParts, " && ")
	}
	
	// 重新组合 OR 部分
	return strings.Join(orParts, " || ")
}
