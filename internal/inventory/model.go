package inventory

import "time"

type SeckillState struct {
	ActivityID     int64
	PriceCents     int64
	StartsAt       time.Time
	EndsAt         time.Time
	ActivityStatus string
	Total          int64
	Available      int64
	Sold           int64
	Version        int64
}
