# Phase 1：PostgreSQL 单体 MVP

## 问题与目标

本阶段在 Phase 0 骨架上完成注册、登录、商品、活动/库存、秒杀、订单查询和取消的最小闭环。目标是先证明串行路径与事务回滚正确，不解决热点并发争用。

## 实现

模块统一遵循 model/dto/errors → repository → service → handler。Handler 只处理 HTTP 绑定与响应；Service 负责校验和事务边界；Repository 接收 `context.Context` 并执行明确 SQL。

认证使用 bcrypt 密码哈希与 HS256 JWT。JWT 密钥只从环境变量读取并要求至少 32 字符；商品和活动写接口要求 admin，订单只能按 token 中的 user_id 查询和取消。

秒杀事务顺序为：读取 Activity/Inventory → 校验状态、UTC 时间和库存 → 以读取结果计算并回写库存 → 创建 Order → 创建 SeckillRecord → commit。任何步骤失败均 rollback。取消事务锁定目标订单，只允许 pending -> cancelled，同时回补库存并更新秒杀记录。

## 故意保留的并发缺陷

`inventory.Repository.SetNaive` 使用“先 SELECT、再把计算后的绝对 available/sold 写回”的 read-modify-write。两个并发事务可能读到同一个旧值并发生 lost update，订单数可能超过库存状态所表达的 sold。Phase 1 不加入 `FOR UPDATE`、条件 UPDATE、CAS、Redis 或本地锁；Phase 2 将用相同负载复现并横向比较三类 PostgreSQL 方案。

## 测试与验证

环境：Windows、Go 1.27.0（项目内 `.tools/go`）、PostgreSQL 17.6 Alpine（Compose，宿主机端口 55432）。

- 单元测试覆盖 JWT 签发/过期、配置、商品价格、活动时间范围、非法数量、活动未开始/结束/非 active、库存不足。
- PostgreSQL 集成测试覆盖成功提交、取消回补、重复用户唯一约束及库存回滚、库存更新后故障注入、订单 CHECK 约束失败回滚。
- HTTP E2E 覆盖注册 → 登录 → 管理权限 → 商品 → 活动/库存 → 秒杀 → 查询 → 取消，并验证 400/401/403/404/409/500 映射。
- `make format`、`go vet ./...`、`go test ./...`、`go build` 均通过；真实集成/E2E 使用 `TEST_DATABASE_URL` 显式启用，未设置时安全跳过数据库测试。
- Phase 1 不进行高并发压测或发布性能数字；基准、QPS、P50/P95/P99 属于 Phase 2。

## 正确性不变量

串行和故障测试中均满足 `available >= 0`、`sold + available = total`、一人一活动最多一条秒杀记录，以及事务失败时库存/订单无部分提交。并发下的已知缺陷被明确保留，因此尚不能宣称并发不超卖。

## Trade-off 与回滚

- bcrypt 安全但登录 CPU 成本高；Phase 1 接受该成本，不缓存密码验证。
- JWT 无服务端会话，易水平扩展，但密钥轮换与主动注销尚未实现。
- offset 分页简单，但深分页受限，因此设置硬上限。
- 取消回补改善库存利用率，但取消用户不可再次参与，规则简单且能抑制刷票。
- migration 回滚使用 `make migrate-down`，会删除 Phase 1 业务表和数据，仅适用于确认无保留需求的开发环境；代码回滚应与 schema 版本同步。

故障详情见 `docs/postmortems/phase1-transaction-rollback.md` 与 `docs/postmortems/phase1-postgresql-outage.md`。
