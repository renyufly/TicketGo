# ADR-0001：库存并发控制采用条件原子 UPDATE

- 状态：已接受
- 日期：2026-08-23
- 适用范围：Phase 2 同步 PostgreSQL 秒杀主链路

## 背景

Phase 1 先读取 `available/sold`，再把 Go 计算出的绝对值写回。相同库存快照可被多个事务重复使用；数据库中的 `available + sold = total` 约束仍成立，但订单数可以超过库存。固定的 1000 用户、100 库存、200 VU、每种方案 3 轮实验中，朴素方案每轮都返回 1000 个成功订单，而数据库只记录售出 40–41 份。

## 决策

默认使用一条带条件的原子 SQL：

```sql
UPDATE inventories
SET available = available - $2,
    sold = sold + $2,
    version = version + 1,
    updated_at = CURRENT_TIMESTAMP
WHERE activity_id = $1
  AND available >= $2
RETURNING id;
```

`RETURNING` 无行表示并发请求已消费剩余库存，Service 返回稳定的 `out_of_stock`；库存更新、订单和秒杀记录仍在同一事务。数据库 CHECK 与唯一约束继续作为最终防线。

## 数据依据

| 方案 | 正确性 | 平均 QPS | 平均 P95 | 平均 P99 | 三轮结果 |
| --- | --- | ---: | ---: | ---: | --- |
| 悲观锁 | 通过 | 1083.77 | 459.00 ms | 570.24 ms | 每轮 100 成功、900 售罄 |
| 条件原子 UPDATE | 通过 | 1561.46 | 401.82 ms | 457.16 ms | 每轮 100 成功、900 售罄 |
| version CAS | 通过 | 1047.48 | 714.50 ms | 775.02 ms | 每轮 100 成功、900 售罄 |

条件 UPDATE 在本机固定场景中吞吐最高、尾延迟最低，而且无需显式读锁或应用重试循环。原始汇总位于 `docs/results/phase2/concurrency-summary.json`。

## 未选方案

- `SELECT ... FOR UPDATE`：语义直观，适合锁定后还要执行多项依赖当前行状态的复杂决策；热点下事务排队明显，锁持有时间直接进入尾延迟。
- version CAS：低冲突时失败快且不长期持锁；热点库存行上发生额外 SELECT、失败 UPDATE、退避和重试。本实验热点成功请求约需 1 次额外 CAS 重试，P95/P99 明显更高。
- 朴素 read-modify-write：只保留为受控实验策略，禁止作为默认配置。

## 后果与回滚

- 优点：检查和扣减在数据库内原子完成，代码短，正确性容易审查；同一事务继续保证库存与订单原子提交。
- 代价：所有请求仍竞争同一 PostgreSQL 热点行，数据库连接和行锁等待成为吞吐上限；这正是 Phase 3 引入 Redis 前需要证明的瓶颈。
- 回滚：设置 `SECKILL_INVENTORY_STRATEGY=pessimistic` 并重启服务可回到已验证的悲观锁实现。实验环境也可选择 `optimistic`；`naive` 只用于复现缺陷。
