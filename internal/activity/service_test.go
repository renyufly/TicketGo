package activity

import (
	"context"
	"errors"
	"testing"
	"ticketgo/internal/domain"
	"time"
)

type fakeRepo struct{}

func (fakeRepo) Create(context.Context, CreateInput) (Activity, error) { return Activity{ID: 1}, nil }
func (fakeRepo) ByID(context.Context, int64) (Activity, error) {
	return Activity{}, errors.New("unused")
}
func (fakeRepo) List(context.Context, int, int) ([]Activity, error) { return nil, nil }
func TestCreateRejectsInvalidTimeRange(t *testing.T) {
	now := time.Now()
	_, err := NewService(fakeRepo{}).Create(context.Background(), CreateInput{ItemID: 1, Name: "sale", PriceCents: 1, Total: 1, StartsAt: now, EndsAt: now})
	if !errors.Is(err, domain.ErrInvalid) {
		t.Fatalf("error=%v", err)
	}
}
