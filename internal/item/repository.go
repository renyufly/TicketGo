package item

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

type Repository interface {
	Create(context.Context, CreateInput) (Item, error)
	ByID(context.Context, int64) (Item, error)
	List(context.Context, int, int) ([]Item, error)
}
type SQLRepository struct{ db *sql.DB }

func NewRepository(db *sql.DB) *SQLRepository { return &SQLRepository{db: db} }

const columns = `id,name,description,price_cents,status,created_at,updated_at`

func scan(row interface{ Scan(...any) error }) (Item, error) {
	var x Item
	err := row.Scan(&x.ID, &x.Name, &x.Description, &x.PriceCents, &x.Status, &x.CreatedAt, &x.UpdatedAt)
	return x, err
}
func (r *SQLRepository) Create(ctx context.Context, in CreateInput) (Item, error) {
	x, err := scan(r.db.QueryRowContext(ctx, `INSERT INTO items(name,description,price_cents,status) VALUES($1,$2,$3,$4) RETURNING `+columns, in.Name, in.Description, in.PriceCents, in.Status))
	if err != nil {
		return Item{}, fmt.Errorf("create item: %w", err)
	}
	return x, nil
}
func (r *SQLRepository) ByID(ctx context.Context, id int64) (Item, error) {
	x, err := scan(r.db.QueryRowContext(ctx, `SELECT `+columns+` FROM items WHERE id=$1`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return Item{}, sql.ErrNoRows
	}
	return x, err
}
func (r *SQLRepository) List(ctx context.Context, limit, offset int) ([]Item, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT `+columns+` FROM items ORDER BY created_at DESC,id DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]Item, 0)
	for rows.Next() {
		x, err := scan(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, rows.Err()
}
