package user

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

type Repository interface {
	Create(context.Context, string, string, string) (User, error)
	ByEmail(context.Context, string) (User, error)
	ByID(context.Context, int64) (User, error)
}
type SQLRepository struct{ db *sql.DB }

func NewRepository(db *sql.DB) *SQLRepository { return &SQLRepository{db: db} }

const userColumns = `id, email, password_hash, role, status, created_at, updated_at`

func scanUser(row interface{ Scan(...any) error }) (User, error) {
	var u User
	err := row.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Role, &u.Status, &u.CreatedAt, &u.UpdatedAt)
	return u, err
}
func (r *SQLRepository) Create(ctx context.Context, email, hash, role string) (User, error) {
	u, err := scanUser(r.db.QueryRowContext(ctx, `INSERT INTO users(email,password_hash,role) VALUES($1,$2,$3) RETURNING `+userColumns, email, hash, role))
	if err != nil {
		return User{}, fmt.Errorf("create user: %w", err)
	}
	return u, nil
}
func (r *SQLRepository) ByEmail(ctx context.Context, email string) (User, error) {
	u, err := scanUser(r.db.QueryRowContext(ctx, `SELECT `+userColumns+` FROM users WHERE LOWER(email)=LOWER($1)`, email))
	if errors.Is(err, sql.ErrNoRows) {
		return User{}, sql.ErrNoRows
	}
	return u, err
}
func (r *SQLRepository) ByID(ctx context.Context, id int64) (User, error) {
	u, err := scanUser(r.db.QueryRowContext(ctx, `SELECT `+userColumns+` FROM users WHERE id=$1`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return User{}, sql.ErrNoRows
	}
	return u, err
}
