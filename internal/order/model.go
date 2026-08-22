package order

import "time"

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
type SeckillInput struct {
	Quantity int64 `json:"quantity" binding:"required,min=1,max=10"`
}
