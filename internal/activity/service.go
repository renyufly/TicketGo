package activity

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
	"ticketgo/internal/domain"
)

type Service struct{ repo Repository }

func NewService(r Repository) *Service { return &Service{repo: r} }
func (s *Service) Create(ctx context.Context, in CreateInput) (Activity, error) {
	in.Name = strings.TrimSpace(in.Name)
	if in.Name == "" || in.ItemID <= 0 || in.PriceCents <= 0 || in.PriceCents > 1000000000000 || in.Total <= 0 || !in.EndsAt.After(in.StartsAt) {
		return Activity{}, domain.New(domain.ErrInvalid, "valid item, name, price, stock and time range are required", nil)
	}
	if in.Status == "" {
		in.Status = "draft"
	}
	a, err := s.repo.Create(ctx, in)
	if err != nil && isForeignKey(err) {
		return Activity{}, domain.New(domain.ErrInvalid, "item does not exist", err)
	}
	return a, err
}
func (s *Service) ByID(ctx context.Context, id int64) (Activity, error) {
	if id <= 0 {
		return Activity{}, domain.New(domain.ErrInvalid, "invalid activity id", nil)
	}
	a, err := s.repo.ByID(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return Activity{}, domain.New(domain.ErrNotFound, "activity not found", err)
	}
	return a, err
}
func (s *Service) List(ctx context.Context, l, o int) ([]Activity, error) {
	return s.repo.List(ctx, l, o)
}
func isForeignKey(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23503"
}
