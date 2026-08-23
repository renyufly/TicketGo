package inventory

import (
	"context"
	"database/sql"
	"errors"
)

type Querier interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}
type Repository struct{}

func NewRepository() *Repository { return &Repository{} }
func (r *Repository) State(ctx context.Context, q Querier, activityID int64) (SeckillState, error) {
	return r.state(ctx, q, activityID, false)
}

// StateForUpdate locks only the matching inventory row. Activity data remains a
// normal MVCC read because the seckill transaction never modifies activities.
func (r *Repository) StateForUpdate(ctx context.Context, q Querier, activityID int64) (SeckillState, error) {
	return r.state(ctx, q, activityID, true)
}

func (r *Repository) state(ctx context.Context, q Querier, activityID int64, forUpdate bool) (SeckillState, error) {
	var s SeckillState
	query := `SELECT a.id,a.price_cents,a.starts_at,a.ends_at,a.status,i.total,i.available,i.sold,i.version FROM activities a JOIN inventories i ON i.activity_id=a.id WHERE a.id=$1`
	if forUpdate {
		query += ` FOR UPDATE OF i`
	}
	err := q.QueryRowContext(ctx, query, activityID).Scan(&s.ActivityID, &s.PriceCents, &s.StartsAt, &s.EndsAt, &s.ActivityStatus, &s.Total, &s.Available, &s.Sold, &s.Version)
	if errors.Is(err, sql.ErrNoRows) {
		return SeckillState{}, sql.ErrNoRows
	}
	return s, err
}

// DeductAtomic combines the stock predicate and mutation in one PostgreSQL
// statement. A false result means another transaction consumed the remaining
// stock before this transaction acquired the row lock.
func (r *Repository) DeductAtomic(ctx context.Context, q Querier, activityID, quantity int64) (bool, error) {
	var id int64
	err := q.QueryRowContext(ctx, `UPDATE inventories SET available=available-$2,sold=sold+$2,version=version+1,updated_at=CURRENT_TIMESTAMP WHERE activity_id=$1 AND available >= $2 RETURNING id`, activityID, quantity).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	return err == nil, err
}

// DeductCAS performs one optimistic-lock compare-and-swap attempt.
func (r *Repository) DeductCAS(ctx context.Context, q Querier, activityID, quantity, version int64) (bool, error) {
	var id int64
	err := q.QueryRowContext(ctx, `UPDATE inventories SET available=available-$2,sold=sold+$2,version=version+1,updated_at=CURRENT_TIMESTAMP WHERE activity_id=$1 AND available >= $2 AND version=$3 RETURNING id`, activityID, quantity, version).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	return err == nil, err
}

// SetNaive deliberately writes values calculated from a prior SELECT. Phase 2 uses
// this read-modify-write window to reproduce lost updates before comparing locks.
func (r *Repository) SetNaive(ctx context.Context, q Querier, activityID, available, sold int64) error {
	err := q.QueryRowContext(ctx, `UPDATE inventories SET available=$2,sold=$3,version=version+1,updated_at=CURRENT_TIMESTAMP WHERE activity_id=$1 RETURNING id`, activityID, available, sold).Scan(new(int64))
	return err
}
func (r *Repository) Restore(ctx context.Context, q Querier, activityID, quantity int64) error {
	err := q.QueryRowContext(ctx, `UPDATE inventories SET available=available+$2,sold=sold-$2,version=version+1,updated_at=CURRENT_TIMESTAMP WHERE activity_id=$1 AND sold >= $2 RETURNING id`, activityID, quantity).Scan(new(int64))
	return err
}
