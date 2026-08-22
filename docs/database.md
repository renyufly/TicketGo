# Phase 1 数据库设计

所有时间列使用 `TIMESTAMPTZ`，数据库容器与连接会话使用 UTC；API 以 RFC3339 传输时间。金额统一为最小货币单位 `price_cents`，禁止浮点数。

## 表与约束

| 表 | 关键语义与约束 |
| --- | --- |
| `users` | 邮箱大小写不敏感唯一；密码只保存 bcrypt 哈希；角色为 customer/admin，状态为 active/disabled。 |
| `items` | 商品原价为正整数，状态为 active/inactive。 |
| `activities` | 必须引用商品；结束时间晚于开始时间；状态为 draft/active/ended/cancelled。 |
| `inventories` | 每个活动恰好一行；`total > 0`、`available/sold >= 0`、`sold + available = total`；保留 version 供 Phase 2 乐观锁实验。 |
| `orders` | 数据库主键与唯一业务号 `order_no` 分离；数量和金额必须为正；状态为 pending/paid/cancelled。 |
| `seckill_records` | `(user_id, activity_id)` 唯一，数据库最终保证一人一活动只有一条秒杀记录；`order_id` 唯一。 |

## 索引理由

- `users(LOWER(email))`：登录与注册冲突判断。
- `activities(item_id)`：外键关联与商品下活动查询；`(status, starts_at)`：活动状态/时间筛选。
- `inventories(activity_id)`：唯一约束同时支持秒杀按活动定位库存。
- `orders(user_id, created_at DESC, id DESC)`：用户订单稳定倒序分页；`orders(activity_id)`：活动对账。
- Phase 1 使用有上限的 offset 分页（limit 1–100、offset 0–10000）；大数据量游标分页和执行计划对比留到 Phase 2。

## 状态机与取消规则

- Activity：`draft -> active -> ended`，`draft/active -> cancelled`；终态不可恢复。Phase 1 只在创建时写 draft/active，尚不提供状态更新 API。
- Order：`pending -> paid` 或 `pending -> cancelled`；Phase 1 尚无支付 API，只实现 pending 取消。
- 取消 pending 订单会在同一事务内回补 `available`、扣减 `sold` 并将秒杀记录标为 cancelled。
- 为防止取消重试刷库存，同一用户取消后也不能再次参加同一活动；唯一秒杀记录会保留。这是本阶段明确的业务取舍。

Schema 只能通过 `migrations/000002_phase1_monolith.*.sql` 变更，应用启动不会自动建表。
