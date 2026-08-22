// 订单模块的 Repository（DAO数据访问层）
// 主要负责直接操作 PostgreSQL：
// 创建订单、查询订单、取消订单，以及维护秒杀记录
/*
Handler
  ↓
Service
  ↓
Repository   ← 这段代码
  ↓
PostgreSQL
*/
// 重点：Querier + 事务、FOR UPDATE 行锁、RETURNING、防止重复秒杀记录

package order

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// 让方法既可以接收 *sql.DB 也可以 *sql.Tx.
// 这样创建订单、秒杀记录、取消订单等操作就可以放进数据库事务里执行
type Querier interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

// 保存数据库连接
type Repository struct{ db *sql.DB }

// 创建 Repository
func NewRepository(db *sql.DB) *Repository { return &Repository{db: db} }

// 把订单表经常查询的字段统一保存起来，避免每条 SQL 重复写一遍
const columns = `id,order_no,user_id,activity_id,quantity,unit_price_cents,total_price_cents,status,created_at,updated_at,cancelled_at`

// 把数据库查询出来的一行数据转换成 Order 对象
func scan(row interface{ Scan(...any) error }) (Order, error) {
	var o Order
	/* Scan()：database/sql 包提供的方法
	作用：把 SQL 查询结果中的一行数据，依次取出来放进 Go 变量.
	传参是&，因为需要变量的地址，才能修改变量的值.
	*/
	err := row.Scan(&o.ID, &o.OrderNo, &o.UserID, &o.ActivityID, &o.Quantity, &o.UnitPriceCents, &o.TotalPriceCents, &o.Status, &o.CreatedAt, &o.UpdatedAt, &o.CancelledAt)
	return o, err
}

// 创建订单
// RETURNING 是 PostgreSQL 的功能，可以插入之后立刻把完整订单返回
func (r *Repository) Create(ctx context.Context, q Querier, o Order) (Order, error) {
	created, err := scan(q.QueryRowContext(ctx, `INSERT INTO orders(order_no,user_id,activity_id,quantity,unit_price_cents,total_price_cents) VALUES($1,$2,$3,$4,$5,$6) RETURNING `+columns, o.OrderNo, o.UserID, o.ActivityID, o.Quantity, o.UnitPriceCents, o.TotalPriceCents))
	if err != nil {
		return Order{}, fmt.Errorf("insert order: %w", err)
	}
	return created, nil
}

// 创建一条秒杀记录
// 用来限制：一个用户不能重复参加同一个秒杀活动
func (r *Repository) CreateRecord(ctx context.Context, q Querier, o Order) error {
	err := q.QueryRowContext(ctx, `INSERT INTO seckill_records(user_id,activity_id,order_id) VALUES($1,$2,$3) RETURNING id`, o.UserID, o.ActivityID, o.ID).Scan(new(int64))
	if err != nil {
		return fmt.Errorf("insert seckill record: %w", err)
	}
	return nil
}
// 检查这个用户是否已经参加过这个秒杀活动
func (r *Repository) RecordExists(ctx context.Context, q Querier, userID, activityID int64) (bool, error) {
	var exists bool
	err := q.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM seckill_records WHERE user_id=$1 AND activity_id=$2)`, userID, activityID).Scan(&exists)
	return exists, err
}

// 查询某个用户自己的订单
func (r *Repository) ByID(ctx context.Context, userID, id int64) (Order, error) {
	o, err := scan(r.db.QueryRowContext(ctx, `SELECT `+columns+` FROM orders WHERE id=$1 AND user_id=$2`, id, userID))
	if errors.Is(err, sql.ErrNoRows) {
		return Order{}, sql.ErrNoRows
	}
	return o, err
}

// 查询订单的同时把这一行锁住 (FOR UPDATE)，通常用于事务。
// 这样两个请求同时取消同一个订单时，不容易发生并发冲突
func (r *Repository) ByIDForUpdate(ctx context.Context, q Querier, userID, id int64) (Order, error) {
	o, err := scan(q.QueryRowContext(ctx, `SELECT `+columns+` FROM orders WHERE id=$1 AND user_id=$2 FOR UPDATE`, id, userID))
	if errors.Is(err, sql.ErrNoRows) {
		return Order{}, sql.ErrNoRows
	}
	return o, err
}

// 查询用户的订单列表，并进行分页
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

// 把订单状态修改为：cancelled ，并记录取消时间
func (r *Repository) Cancel(ctx context.Context, q Querier, id int64) (Order, error) {
	return scan(q.QueryRowContext(ctx, `UPDATE orders SET status='cancelled',cancelled_at=CURRENT_TIMESTAMP,updated_at=CURRENT_TIMESTAMP WHERE id=$1 RETURNING `+columns, id))
}

// 订单取消后，同时把对应的秒杀记录标记为取消
func (r *Repository) CancelRecord(ctx context.Context, q Querier, orderID int64) error {
	err := q.QueryRowContext(ctx, `UPDATE seckill_records SET status='cancelled',updated_at=CURRENT_TIMESTAMP WHERE order_id=$1 RETURNING id`, orderID).Scan(new(int64))
	return err
}
