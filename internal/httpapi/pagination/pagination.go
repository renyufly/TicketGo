// 分页参数解析工具
// 从 URL Query 中读取 limit 和 offset，并检查是否合法
/*
统一处理：HTTP请求：GET /orders?limit=20&offset=40
这样 order、item、activity 等列表接口都可以复用同一套分页逻辑
*/

package pagination

import (
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
)

// Page 保存分页参数
// Limit：一次最多查多少条
// Offset：跳过前多少条
type Page struct {
	Limit  int
	Offset int
}

// 读取参数
func Parse(c *gin.Context) (Page, error) {
	// 有 limit → 使用用户传入的值
	// 没有 → 默认 "20"
	limit, err := positiveInt(c.DefaultQuery("limit", "20"))
	
	if err != nil || limit > 100 {
		return Page{}, fmt.Errorf("limit must be between 1 and 100")
	}

	// offset 默认从第 0 条开始
	offset, err := nonNegativeInt(c.DefaultQuery("offset", "0"))
	if err != nil || offset > 10000 {
		return Page{}, fmt.Errorf("offset must be between 0 and 10000")
	}

	return Page{Limit: limit, Offset: offset}, nil
}

// 把字符串转成正整数
func positiveInt(v string) (int, error) {
	n, err := strconv.Atoi(v)
	if err != nil || n <= 0 {
		return 0, fmt.Errorf("positive integer required")
	}
	return n, nil
}

// 允许非负数：0 + 正数
func nonNegativeInt(v string) (int, error) {
	n, err := strconv.Atoi(v)
	if err != nil || n < 0 {
		return 0, fmt.Errorf("non-negative integer required")
	}
	return n, nil
}
