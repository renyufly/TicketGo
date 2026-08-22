# Phase 1 Postmortem：事务中途失败

## Incident / Impact

分别在库存 UPDATE 后、订单 INSERT 前注入错误，以及把订单 quantity 改为 0 触发数据库 CHECK 约束。两次请求均失败，无订单对用户可见。

## Root Cause / Detection

第一类是模拟应用异常，第二类是数据库拒绝非法订单。集成测试在请求后查询 inventory 与 orders，直接检查不变量。

## Resolution / Result

Service 将库存、订单、秒杀记录包在同一 PostgreSQL transaction 中并 defer rollback。两次实验的库存都保持 `total=2, available=2, sold=0`，订单数为 0。重复用户触发 `(user_id, activity_id)` 唯一约束时，也回滚此前库存修改。

## Prevention / Trade-off

保留数据库 CHECK、FK、UNIQUE 作为最终防线，并持续运行真实 PostgreSQL 集成测试。事务提高原子性，但没有自动解决并发 lost update；该问题明确进入 Phase 2。
