package item

import (
	"context"
	"errors"
	"testing"
	"ticketgo/internal/domain"
)

type fakeRepo struct{}

func (fakeRepo) Create(context.Context, CreateInput) (Item, error) { return Item{ID: 1}, nil }
func (fakeRepo) ByID(context.Context, int64) (Item, error)         { return Item{}, errors.New("unused") }
func (fakeRepo) List(context.Context, int, int) ([]Item, error)    { return nil, nil }
func TestCreateRejectsInvalidPrice(t *testing.T) {
	_, err := NewService(fakeRepo{}).Create(context.Background(), CreateInput{Name: "ticket", PriceCents: 0})
	if !errors.Is(err, domain.ErrInvalid) {
		t.Fatalf("error=%v", err)
	}
}
