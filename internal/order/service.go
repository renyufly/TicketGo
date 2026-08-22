// 订单业务层 Service，核心负责两件事：秒杀下单和取消订单。
// 它把库存、订单、事务、业务规则串起来
// 订单业务规则，并通过数据库事务保证“库存变化”和“订单变化”保持一致
// 涉及：事务、并发、库存一致性、唯一约束、防重复下单、行锁和错误转换

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

// db：开启数据库事务
// repo：操作订单表
// inventory：操作库存
// now：获取当前时间
// afterInventoryUpdate / beforeOrderInsert：主要用于测试时人为插入异常
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

// 核心-秒杀下单
/* 检查参数 -> 开启事务 -> 检查用户是否已经参加过
   -> 查询活动和库存 -> 检查活动是否开始/结束、库存是否足够
   -> 扣库存 -> 生成订单号 -> 创建订单 -> 创建“用户已参与活动”的记录
   -> 提交事务     */
func (s *Service) Seckill(ctx context.Context, userID, activityID, quantity int64) (Order, error) {
	/*
		ctx：请求上下文，用于超时、取消请求，以及传递给数据库操作
		userID：谁在买
		activityID：参加哪个秒杀活动
		quantity：买几件
		返回 Order：成功时返回创建好的订单
		返回 error：失败原因
	*/

	// 参数校验
	if userID <= 0 || activityID <= 0 || quantity <= 0 || quantity > 10 {
		return Order{}, domain.New(domain.ErrInvalid, "valid activity and quantity between 1 and 10 are required", nil)
	}

	// 开启了一个数据库事务:
	// "扣库存 + 创建订单 + 创建参与记录" 必须一起成功.
	// 事务保证的原子性.
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Order{}, err
	}
	// 使用事务tx后，只要中间任何一步失败，函数退出时：
	// 就会回滚事务。于是之前扣掉的库存也会撤销.
	/* defer 一定会在函数return后执行，但
	   事务已经 Commit() 之后，再调用 Rollback() 通常只会返回类似
	   sql.ErrTxDone，不会把已经提交的数据撤销.
	   这样不需要每个错误分支都手写：tx.Rollback()
	*/
	defer tx.Rollback()

	// 防止重复秒杀, 先检查是否已存在参与记录
	// 代码不仅在业务层检查一次，数据库唯一约束也会再检查一次，
	// 防止并发情况下两个请求同时通过前面的判断.
	exists, err := s.repo.RecordExists(ctx, tx, userID, activityID)
	if err != nil {
		return Order{}, err
	}
	if err := validateNotDuplicate(exists); err != nil {
		return Order{}, err
	}

	// 查询当前秒杀活动状态和库存
	state, err := s.inventory.State(ctx, tx, activityID)
	if errors.Is(err, sql.ErrNoRows) {
		return Order{}, domain.New(domain.ErrNotFound, "activity not found", err)
	}
	if err != nil {
		return Order{}, err
	}
	now := s.now().UTC()

	// 正式检查当前是否允许秒杀
	if err := validateSeckill(state, quantity, now); err != nil {
		return Order{}, err
	}

	// 重点：扣库存 - TODO
	// SetNaive 表明是 较“朴素”的库存更新方式.
	// 后续优化重点在这里，高并发秒杀.
	if err = s.inventory.SetNaive(ctx, tx, activityID, state.Available-quantity, state.Sold+quantity); err != nil {
		return Order{}, err
	}

	// afterInventoryUpdate主要是一个测试钩子 Hook
	// 可人为制造错误，测试事务原子性
	if s.afterInventoryUpdate != nil {
		if err = s.afterInventoryUpdate(); err != nil {
			return Order{}, fmt.Errorf("injected after inventory update: %w", err)
		}
	}

	// 创建订单，生成订单号
	orderNo, err := newOrderNo(now)
	if err != nil {
		return Order{}, err
	}
	// 计算订单价格
	pending := Order{OrderNo: orderNo, UserID: userID, ActivityID: activityID, Quantity: quantity, UnitPriceCents: state.PriceCents, TotalPriceCents: state.PriceCents * quantity}

	// beforeOrderInsert：测试Hook
	// 允许测试代码在订单真正插入数据库前修改 pending.
	if s.beforeOrderInsert != nil {
		s.beforeOrderInsert(&pending)
	}

	// 真正插入数据库，创建订单
	o, err := s.repo.Create(ctx, tx, pending)
	if err != nil {
		return Order{}, err
	}

	// 创建：用户参加活动的记录
	// 防止同一个用户重复参加活动
	if err = s.repo.CreateRecord(ctx, tx, o); err != nil {
		if isUnique(err) {
			return Order{}, domain.New(domain.ErrConflict, "user has already joined this activity", err)
		}
		return Order{}, err
	}

	// tx.Commit()之后，数据库才正式确认这些修改
	if err = tx.Commit(); err != nil {
		return Order{}, err
	}

	return o, nil
}

// 检查能不能秒杀：秒杀业务规则校验器
func validateSeckill(state inventory.SeckillState, quantity int64, now time.Time) error {
	// 1.数量是否合法
	if quantity <= 0 || quantity > 10 {
		return domain.New(domain.ErrInvalid, "quantity must be between 1 and 10", nil)
	}

	// 2.活动是不是正在进行
	if state.ActivityStatus != "active" || now.Before(state.StartsAt) || !now.Before(state.EndsAt) {
		return domain.New(domain.ErrActivityClosed, "activity is not currently available", nil)
	}
	// 3.库存够不够
	if state.Available < quantity {
		return domain.New(domain.ErrOutOfStock, "insufficient inventory", nil)
	}
	return nil
}

// 检查用户是否已经参加过这个秒杀活动
func validateNotDuplicate(exists bool) error {
	if exists {
		// 返回业务错误
		return domain.New(domain.ErrConflict, "user has already joined this activity", nil)
	}
	return nil
}

// 根据订单 ID 查询某个用户自己的订单
func (s *Service) ByID(ctx context.Context, userID, id int64) (Order, error) {
	if id <= 0 {
		return Order{}, domain.New(domain.ErrInvalid, "invalid order id", nil)
	}
	// 真正调用 Repository 查询数据库 （DAO层）
	// 同时使用 userID + orderID 查询，是一种防止越权访问的设计.
	o, err := s.repo.ByID(ctx, userID, id)
	if errors.Is(err, sql.ErrNoRows) {
		// Service 把数据库层错误sql 转换成业务层错误domain
		return Order{}, domain.New(domain.ErrNotFound, "order not found", err)
	}
	return o, err
}

// 查询某个用户的订单列表
/* 分页参数：
l = limit
o = offset
*/
func (s *Service) List(ctx context.Context, userID int64, l, o int) ([]Order, error) {
	return s.repo.List(ctx, userID, l, o)
}

// 取消订单
/*
开启事务 -> 锁定并查询订单 -> 检查订单是不是 pending
-> 恢复库存 -> 订单改为 cancelled -> 删除/取消参与记录
-> 提交事务     */
func (s *Service) Cancel(ctx context.Context, userID, id int64) (Order, error) {
	// 开启事务Transaction
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Order{}, err
	}
	defer tx.Rollback()

	// ForUpdate 通常意味着 SQL 会给订单记录加行锁
	// 避免两个请求同时取消同一个订单
	o, err := s.repo.ByIDForUpdate(ctx, tx, userID, id)

	if errors.Is(err, sql.ErrNoRows) {
		// 如果数据库没有查到订单
		return Order{}, domain.New(domain.ErrNotFound, "order not found", err)
	}
	if err != nil {
		// 如果不是“没找到”，而是其他数据库问题
		return Order{}, err
	}

	// 检查订单状态：规定只有 pending(待处理) 状态的订单可以取消
	if o.Status != "pending" {
		return Order{}, domain.New(domain.ErrConflict, "only pending orders can be cancelled", nil)
	}

	// 恢复库存：
	// 原来秒杀时扣了库存，现在取消订单，就需要恢复对应活动的库存
	// 注意：恢复库存也是整个取消事务 tx 的一部分
	if err = s.inventory.Restore(ctx, tx, o.ActivityID, o.Quantity); err != nil {
		return Order{}, err
	}

	// 真正修改订单状态（DAO层）
	o, err = s.repo.Cancel(ctx, tx, o.ID)

	if err != nil {
		// 如果取消订单 SQL 失败，就直接退出
		return Order{}, err
	}

	// 处理“参与记录”：取消或者删除这条用户的秒杀活动参与记录
	// 因为若取消后仍保留一个“有效参与记录”，用户可能永远无法再次参加
	if err = s.repo.CancelRecord(ctx, tx, o.ID); err != nil {
		return Order{}, err
	}

	// 正式提交整个取消事务操作
	if err = tx.Commit(); err != nil {
		return Order{}, err
	}
	return o, nil
}

// 生成订单号
func newOrderNo(now time.Time) (string, error) {
	// 生成 8 字节随机数
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	// 结构：TG + 时间 + 随机数
	return fmt.Sprintf("TG%s%s", now.UTC().Format("20060102150405"), hex.EncodeToString(b)), nil
}

// 数据库唯一约束 再次防止重复秒杀
func isUnique(err error) bool {
	var pgErr *pgconn.PgError
	// 23505 是 PostgreSQL 的 unique_violation 唯一约束冲突
	// 同一个用户重复秒杀时，数据库就可能返回 23505
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
