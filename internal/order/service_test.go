package order

import (
	"errors"
	"testing"
	"time"

	"ticketgo/internal/domain"
	"ticketgo/internal/inventory"
)

func TestValidateSeckill(t *testing.T) {
	now := time.Date(2026, 8, 22, 12, 0, 0, 0, time.UTC)
	base := inventory.SeckillState{ActivityStatus: "active", StartsAt: now.Add(-time.Hour), EndsAt: now.Add(time.Hour), Available: 5}
	tests := []struct {
		name     string
		state    inventory.SeckillState
		quantity int64
		want     error
	}{
		{"valid", base, 1, nil},
		{"invalid quantity", base, 0, domain.ErrInvalid},
		{"not started", withTimes(base, now.Add(time.Minute), now.Add(time.Hour)), 1, domain.ErrActivityClosed},
		{"ended", withTimes(base, now.Add(-time.Hour), now), 1, domain.ErrActivityClosed},
		{"draft", withStatus(base, "draft"), 1, domain.ErrActivityClosed},
		{"out of stock", base, 6, domain.ErrOutOfStock},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateSeckill(tt.state, tt.quantity, now)
			if tt.want == nil && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.want != nil && !errors.Is(err, tt.want) {
				t.Fatalf("error=%v want %v", err, tt.want)
			}
		})
	}
}

func TestValidateNotDuplicate(t *testing.T) {
	if err := validateNotDuplicate(false); err != nil {
		t.Fatalf("new user rejected: %v", err)
	}
	if err := validateNotDuplicate(true); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("duplicate error=%v, want conflict", err)
	}
}
func withTimes(s inventory.SeckillState, start, end time.Time) inventory.SeckillState {
	s.StartsAt = start
	s.EndsAt = end
	return s
}
func withStatus(s inventory.SeckillState, status string) inventory.SeckillState {
	s.ActivityStatus = status
	return s
}
