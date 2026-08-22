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
	var s SeckillState
	err := q.QueryRowContext(ctx, `SELECT a.id,a.price_cents,a.starts_at,a.ends_at,a.status,i.total,i.available,i.sold,i.version FROM activities a JOIN inventories i ON i.activity_id=a.id WHERE a.id=$1`, activityID).Scan(&s.ActivityID, &s.PriceCents, &s.StartsAt, &s.EndsAt, &s.ActivityStatus, &s.Total, &s.Available, &s.Sold, &s.Version)
	if errors.Is(err, sql.ErrNoRows) {
		return SeckillState{}, sql.ErrNoRows
	}
	return s, err
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
