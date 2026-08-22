package activity

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

type Repository interface {
	Create(context.Context, CreateInput) (Activity, error)
	ByID(context.Context, int64) (Activity, error)
	List(context.Context, int, int) ([]Activity, error)
}
type SQLRepository struct{ db *sql.DB }

func NewRepository(db *sql.DB) *SQLRepository { return &SQLRepository{db: db} }

const selectColumns = `a.id,a.item_id,a.name,a.price_cents,a.starts_at,a.ends_at,a.status,i.total,i.available,i.sold,i.version,a.created_at,a.updated_at`

func scan(row interface{ Scan(...any) error }) (Activity, error) {
	var a Activity
	err := row.Scan(&a.ID, &a.ItemID, &a.Name, &a.PriceCents, &a.StartsAt, &a.EndsAt, &a.Status, &a.Total, &a.Available, &a.Sold, &a.Version, &a.CreatedAt, &a.UpdatedAt)
	return a, err
}
func (r *SQLRepository) Create(ctx context.Context, in CreateInput) (Activity, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return Activity{}, err
	}
	defer tx.Rollback()
	var a Activity
	err = tx.QueryRowContext(ctx, `INSERT INTO activities(item_id,name,price_cents,starts_at,ends_at,status) VALUES($1,$2,$3,$4,$5,$6) RETURNING id,item_id,name,price_cents,starts_at,ends_at,status,created_at,updated_at`, in.ItemID, in.Name, in.PriceCents, in.StartsAt, in.EndsAt, in.Status).Scan(&a.ID, &a.ItemID, &a.Name, &a.PriceCents, &a.StartsAt, &a.EndsAt, &a.Status, &a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		return Activity{}, fmt.Errorf("insert activity: %w", err)
	}
	err = tx.QueryRowContext(ctx, `INSERT INTO inventories(activity_id,total,available) VALUES($1,$2,$2) RETURNING total,available,sold,version`, a.ID, in.Total).Scan(&a.Total, &a.Available, &a.Sold, &a.Version)
	if err != nil {
		return Activity{}, fmt.Errorf("insert inventory: %w", err)
	}
	if err = tx.Commit(); err != nil {
		return Activity{}, err
	}
	return a, nil
}
func (r *SQLRepository) ByID(ctx context.Context, id int64) (Activity, error) {
	a, err := scan(r.db.QueryRowContext(ctx, `SELECT `+selectColumns+` FROM activities a JOIN inventories i ON i.activity_id=a.id WHERE a.id=$1`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return Activity{}, sql.ErrNoRows
	}
	return a, err
}
func (r *SQLRepository) List(ctx context.Context, limit, offset int) ([]Activity, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT `+selectColumns+` FROM activities a JOIN inventories i ON i.activity_id=a.id ORDER BY a.created_at DESC,a.id DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]Activity, 0)
	for rows.Next() {
		a, err := scan(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}
