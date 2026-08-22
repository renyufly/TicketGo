// 定义了订单的数据结构和秒杀请求的输入结构

package order

import "time"

// Order：订单模型
// 注意：价格使用 int64 的分而不是 float64 的元，为了避免浮点数精度问题
// JSON Tag 表示转换成json后的名字
// 注意：CancelledAt是 *time.Time，因为订单可能没有被取消，
//
//	omitempty 表示如果它是 nil，JSON 中可以直接省略这个字段
type Order struct {
	ID              int64      `json:"id"`
	OrderNo         string     `json:"order_no"`
	UserID          int64      `json:"user_id"`
	ActivityID      int64      `json:"activity_id"`
	Quantity        int64      `json:"quantity"`
	UnitPriceCents  int64      `json:"unit_price_cents"`
	TotalPriceCents int64      `json:"total_price_cents"`
	Status          string     `json:"status"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	CancelledAt     *time.Time `json:"cancelled_at,omitempty"`
}

// 表示秒杀接口接收 { "quantity": 2 }
type SeckillInput struct {
	Quantity int64 `json:"quantity" binding:"required,min=1,max=10"`
}
