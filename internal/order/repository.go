package order

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

type Querier interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}
type Repository struct{ db *sql.DB }

func NewRepository(db *sql.DB) *Repository { return &Repository{db: db} }

const columns = `id,order_no,user_id,activity_id,quantity,unit_price_cents,total_price_cents,status,created_at,updated_at,cancelled_at`

func scan(row interface{ Scan(...any) error }) (Order, error) {
	var o Order
	err := row.Scan(&o.ID, &o.OrderNo, &o.UserID, &o.ActivityID, &o.Quantity, &o.UnitPriceCents, &o.TotalPriceCents, &o.Status, &o.CreatedAt, &o.UpdatedAt, &o.CancelledAt)
	return o, err
}
func (r *Repository) Create(ctx context.Context, q Querier, o Order) (Order, error) {
	created, err := scan(q.QueryRowContext(ctx, `INSERT INTO orders(order_no,user_id,activity_id,quantity,unit_price_cents,total_price_cents) VALUES($1,$2,$3,$4,$5,$6) RETURNING `+columns, o.OrderNo, o.UserID, o.ActivityID, o.Quantity, o.UnitPriceCents, o.TotalPriceCents))
	if err != nil {
		return Order{}, fmt.Errorf("insert order: %w", err)
	}
	return created, nil
}
func (r *Repository) CreateRecord(ctx context.Context, q Querier, o Order) error {
	err := q.QueryRowContext(ctx, `INSERT INTO seckill_records(user_id,activity_id,order_id) VALUES($1,$2,$3) RETURNING id`, o.UserID, o.ActivityID, o.ID).Scan(new(int64))
	if err != nil {
		return fmt.Errorf("insert seckill record: %w", err)
	}
	return nil
}
func (r *Repository) RecordExists(ctx context.Context, q Querier, userID, activityID int64) (bool, error) {
	var exists bool
	err := q.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM seckill_records WHERE user_id=$1 AND activity_id=$2)`, userID, activityID).Scan(&exists)
	return exists, err
}
func (r *Repository) ByID(ctx context.Context, userID, id int64) (Order, error) {
	o, err := scan(r.db.QueryRowContext(ctx, `SELECT `+columns+` FROM orders WHERE id=$1 AND user_id=$2`, id, userID))
	if errors.Is(err, sql.ErrNoRows) {
		return Order{}, sql.ErrNoRows
	}
	return o, err
}
func (r *Repository) ByIDForUpdate(ctx context.Context, q Querier, userID, id int64) (Order, error) {
	o, err := scan(q.QueryRowContext(ctx, `SELECT `+columns+` FROM orders WHERE id=$1 AND user_id=$2 FOR UPDATE`, id, userID))
	if errors.Is(err, sql.ErrNoRows) {
		return Order{}, sql.ErrNoRows
	}
	return o, err
}
func (r *Repository) List(ctx context.Context, userID int64, limit, offset int) ([]Order, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT `+columns+` FROM orders WHERE user_id=$1 ORDER BY created_at DESC,id DESC LIMIT $2 OFFSET $3`, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]Order, 0)
	for rows.Next() {
		o, err := scan(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, o)
	}
	return out, rows.Err()
}
func (r *Repository) Cancel(ctx context.Context, q Querier, id int64) (Order, error) {
	return scan(q.QueryRowContext(ctx, `UPDATE orders SET status='cancelled',cancelled_at=CURRENT_TIMESTAMP,updated_at=CURRENT_TIMESTAMP WHERE id=$1 RETURNING `+columns, id))
}
func (r *Repository) CancelRecord(ctx context.Context, q Querier, orderID int64) error {
	err := q.QueryRowContext(ctx, `UPDATE seckill_records SET status='cancelled',updated_at=CURRENT_TIMESTAMP WHERE order_id=$1 RETURNING id`, orderID).Scan(new(int64))
	return err
}
