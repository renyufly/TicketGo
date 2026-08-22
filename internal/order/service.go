package order

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5/pgconn"
	"ticketgo/internal/domain"
	"ticketgo/internal/inventory"
	"time"
)

type Beginner interface {
	BeginTx(context.Context, *sql.TxOptions) (*sql.Tx, error)
}
type Service struct {
	db                   Beginner
	repo                 *Repository
	inventory            *inventory.Repository
	now                  func() time.Time
	afterInventoryUpdate func() error
	beforeOrderInsert    func(*Order)
}

func NewService(db Beginner, r *Repository, i *inventory.Repository) *Service {
	return &Service{db: db, repo: r, inventory: i, now: time.Now}
}
func (s *Service) Seckill(ctx context.Context, userID, activityID, quantity int64) (Order, error) {
	if userID <= 0 || activityID <= 0 || quantity <= 0 || quantity > 10 {
		return Order{}, domain.New(domain.ErrInvalid, "valid activity and quantity between 1 and 10 are required", nil)
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Order{}, err
	}
	defer tx.Rollback()
	exists, err := s.repo.RecordExists(ctx, tx, userID, activityID)
	if err != nil {
		return Order{}, err
	}
	if err := validateNotDuplicate(exists); err != nil {
		return Order{}, err
	}
	state, err := s.inventory.State(ctx, tx, activityID)
	if errors.Is(err, sql.ErrNoRows) {
		return Order{}, domain.New(domain.ErrNotFound, "activity not found", err)
	}
	if err != nil {
		return Order{}, err
	}
	now := s.now().UTC()
	if err := validateSeckill(state, quantity, now); err != nil {
		return Order{}, err
	}
	if err = s.inventory.SetNaive(ctx, tx, activityID, state.Available-quantity, state.Sold+quantity); err != nil {
		return Order{}, err
	}
	if s.afterInventoryUpdate != nil {
		if err = s.afterInventoryUpdate(); err != nil {
			return Order{}, fmt.Errorf("injected after inventory update: %w", err)
		}
	}
	orderNo, err := newOrderNo(now)
	if err != nil {
		return Order{}, err
	}
	pending := Order{OrderNo: orderNo, UserID: userID, ActivityID: activityID, Quantity: quantity, UnitPriceCents: state.PriceCents, TotalPriceCents: state.PriceCents * quantity}
	if s.beforeOrderInsert != nil {
		s.beforeOrderInsert(&pending)
	}
	o, err := s.repo.Create(ctx, tx, pending)
	if err != nil {
		return Order{}, err
	}
	if err = s.repo.CreateRecord(ctx, tx, o); err != nil {
		if isUnique(err) {
			return Order{}, domain.New(domain.ErrConflict, "user has already joined this activity", err)
		}
		return Order{}, err
	}
	if err = tx.Commit(); err != nil {
		return Order{}, err
	}
	return o, nil
}

func validateSeckill(state inventory.SeckillState, quantity int64, now time.Time) error {
	if quantity <= 0 || quantity > 10 {
		return domain.New(domain.ErrInvalid, "quantity must be between 1 and 10", nil)
	}
	if state.ActivityStatus != "active" || now.Before(state.StartsAt) || !now.Before(state.EndsAt) {
		return domain.New(domain.ErrActivityClosed, "activity is not currently available", nil)
	}
	if state.Available < quantity {
		return domain.New(domain.ErrOutOfStock, "insufficient inventory", nil)
	}
	return nil
}

func validateNotDuplicate(exists bool) error {
	if exists {
		return domain.New(domain.ErrConflict, "user has already joined this activity", nil)
	}
	return nil
}
func (s *Service) ByID(ctx context.Context, userID, id int64) (Order, error) {
	if id <= 0 {
		return Order{}, domain.New(domain.ErrInvalid, "invalid order id", nil)
	}
	o, err := s.repo.ByID(ctx, userID, id)
	if errors.Is(err, sql.ErrNoRows) {
		return Order{}, domain.New(domain.ErrNotFound, "order not found", err)
	}
	return o, err
}
func (s *Service) List(ctx context.Context, userID int64, l, o int) ([]Order, error) {
	return s.repo.List(ctx, userID, l, o)
}
func (s *Service) Cancel(ctx context.Context, userID, id int64) (Order, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Order{}, err
	}
	defer tx.Rollback()
	o, err := s.repo.ByIDForUpdate(ctx, tx, userID, id)
	if errors.Is(err, sql.ErrNoRows) {
		return Order{}, domain.New(domain.ErrNotFound, "order not found", err)
	}
	if err != nil {
		return Order{}, err
	}
	if o.Status != "pending" {
		return Order{}, domain.New(domain.ErrConflict, "only pending orders can be cancelled", nil)
	}
	if err = s.inventory.Restore(ctx, tx, o.ActivityID, o.Quantity); err != nil {
		return Order{}, err
	}
	o, err = s.repo.Cancel(ctx, tx, o.ID)
	if err != nil {
		return Order{}, err
	}
	if err = s.repo.CancelRecord(ctx, tx, o.ID); err != nil {
		return Order{}, err
	}
	if err = tx.Commit(); err != nil {
		return Order{}, err
	}
	return o, nil
}
func newOrderNo(now time.Time) (string, error) {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return fmt.Sprintf("TG%s%s", now.UTC().Format("20060102150405"), hex.EncodeToString(b)), nil
}
func isUnique(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
