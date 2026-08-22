package activity

import "time"

type Activity struct {
	ID         int64     `json:"id"`
	ItemID     int64     `json:"item_id"`
	Name       string    `json:"name"`
	PriceCents int64     `json:"price_cents"`
	StartsAt   time.Time `json:"starts_at"`
	EndsAt     time.Time `json:"ends_at"`
	Status     string    `json:"status"`
	Total      int64     `json:"total"`
	Available  int64     `json:"available"`
	Sold       int64     `json:"sold"`
	Version    int64     `json:"version"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
type CreateInput struct {
	ItemID     int64     `json:"item_id" binding:"required,min=1"`
	Name       string    `json:"name" binding:"required,max=200"`
	PriceCents int64     `json:"price_cents" binding:"required,min=1,max=1000000000000"`
	StartsAt   time.Time `json:"starts_at" binding:"required"`
	EndsAt     time.Time `json:"ends_at" binding:"required"`
	Status     string    `json:"status" binding:"omitempty,oneof=draft active"`
	Total      int64     `json:"total" binding:"required,min=1"`
}
