package item

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"ticketgo/internal/domain"
)

type Service struct{ repo Repository }

func NewService(r Repository) *Service { return &Service{repo: r} }
func (s *Service) Create(ctx context.Context, in CreateInput) (Item, error) {
	in.Name = strings.TrimSpace(in.Name)
	if in.Name == "" || in.PriceCents <= 0 || in.PriceCents > 1000000000000 {
		return Item{}, domain.New(domain.ErrInvalid, "name and positive price_cents are required", nil)
	}
	if in.Status == "" {
		in.Status = "active"
	}
	return s.repo.Create(ctx, in)
}
func (s *Service) ByID(ctx context.Context, id int64) (Item, error) {
	if id <= 0 {
		return Item{}, domain.New(domain.ErrInvalid, "invalid item id", nil)
	}
	x, err := s.repo.ByID(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return Item{}, domain.New(domain.ErrNotFound, "item not found", err)
	}
	return x, err
}
func (s *Service) List(ctx context.Context, limit, offset int) ([]Item, error) {
	return s.repo.List(ctx, limit, offset)
}
