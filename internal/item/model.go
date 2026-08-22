package item

import "time"

type Item struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	PriceCents  int64     `json:"price_cents"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
type CreateInput struct {
	Name        string `json:"name" binding:"required,max=200"`
	Description string `json:"description" binding:"max=2000"`
	PriceCents  int64  `json:"price_cents" binding:"required,min=1,max=1000000000000"`
	Status      string `json:"status" binding:"omitempty,oneof=active inactive"`
}
