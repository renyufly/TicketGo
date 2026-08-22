# TicketGo 高并发秒杀系统执行计划

> 本计划由 `Idea.md` 落地而来。主线数据库统一为 **PostgreSQL**；MySQL/InnoDB 只作为面试对比材料，不进入主业务运行链路。
>
> 执行原则：每个阶段都必须完成“业务场景 → 最小实现 → 问题复现 → 方案实现 → 压测/故障验证 → trade-off 报告”。未通过阶段门禁，不提前引入下一阶段组件。

## 1. 项目目标与完成定义

最终交付一个从 PostgreSQL 单体逐步演进到 Redis、Kafka、gRPC、微服务和 Kubernetes 的抢票/秒杀系统，并保留完整的实验数据、故障记录和技术决策依据。

项目完成需同时满足：

- 能完整运行用户、商品、活动、库存、秒杀和订单主链路；
- 在并发下不超卖，同一用户同一活动最多成功一次；
- 能解释每次架构演进解决的问题及引入的代价；
- 具备单元、集成、端到端、负载和故障测试；
- 指标、日志、Trace 可以定位主要性能和可靠性问题；
- 所有简历性能数据均来自可复现的测试报告；
- 新环境可按 README 从零启动，不依赖开发者机器的隐式配置。

## 2. 全程统一约定

### 2.1 技术基线

- Go：使用项目启动时的稳定版本，并在 `go.mod`、CI、Dockerfile 中锁定；
- HTTP：Gin；数据库：PostgreSQL；迁移工具：golang-migrate 或 Goose，Phase 0 选定后不混用；
- 前期 CRUD 可使用 GORM，库存扣减、锁、Explain 和性能敏感查询必须使用明确可审查的 SQL；
- 日志使用 zap，配置先用环境变量，服务增多后再引入 Viper/etcd；
- 测试使用 Go testing、真实容器集成环境和 k6；
- 后续按阶段引入 Redis、Kafka、Nginx、gRPC、etcd、Prometheus、Grafana、OpenTelemetry、Jaeger/Tempo 和 Kubernetes。

### 2.2 API 与数据约定

- 所有外部 API 使用 `/api/v1`；健康检查使用 `/health/live` 和 `/health/ready`；
- 写请求支持请求 ID；异步链路增加 `event_id` 和 `trace_id`；
- 时间在数据库统一使用 UTC，API 使用 RFC3339；金额禁止 float，使用最小货币单位整数或定点数；
- 订单号为业务标识，数据库主键与业务订单号分离；
- 错误响应至少包含 `code`、`message`、`request_id`，不得把数据库内部错误直接返回客户端；
- 日志禁止记录密码、Token、完整个人敏感信息；
- 所有表结构变更只能通过 migration 完成，不依赖自动建表作为正式方案。

### 2.3 阶段报告模板

每个阶段的报告至少包含：

1. 问题与目标；
2. 环境（CPU、内存、Go/PostgreSQL/Redis/Kafka 版本）；
3. 数据规模、k6 脚本和测试参数；
4. 优化前后 QPS、P50、P95、P99、错误率、资源使用；
5. 正确性不变量验证；
6. 故障注入、观察和恢复结果；
7. 方案优点、缺点、适用边界和回滚方式；
8. 结论及下一阶段为什么必要。

### 2.4 核心正确性不变量

每次改动秒杀链路后都必须自动验证：

- `available >= 0`；
- `sold + available = total`（若存在冻结库存，则明确扩展等式）；
- 成功订单数量/数量合计不超过活动总库存；
- `(user_id, activity_id)` 最多存在一条有效秒杀记录/订单；
- 消费同一 `event_id` 多次不会产生重复订单；
- 请求失败或事务回滚时，库存与订单不会形成不可解释的永久不一致；
- 异步模型下的暂时不一致可被监控、重试或对账任务收敛。

## 3. 建议仓库演进

Phase 0–4 保持模块化单体，Phase 5 再迁移为 Monorepo 多服务。

```text
TicketGo/
├── cmd/server/
├── internal/
│   ├── user/
│   ├── item/
│   ├── activity/
│   ├── inventory/
│   └── order/
├── pkg/
├── configs/
├── migrations/
├── scripts/
├── tests/
│   ├── integration/
│   ├── e2e/
│   ├── load/
│   └── chaos/
├── deployments/
├── docs/
├── Dockerfile
├── docker-compose.yml
├── Makefile
└── README.md
```

Phase 5 目标结构：

```text
TicketGo/
├── services/{gateway,user,activity,inventory,order}/
├── api/proto/
├── pkg/
├── deployments/
├── scripts/
├── tests/
└── docs/
```

---

## Phase 0：工程准备与可启动骨架

### 阶段目标

从空仓库建立一个可重复启动、可迁移、可观测、可优雅退出的 Go 单体骨架。本阶段不实现复杂业务。

### 执行步骤

1. 初始化工程：
   - 创建 Go module、`cmd/server/main.go` 和 `internal`/`pkg` 基础目录；
   - 确定配置加载、依赖装配和模块边界；
   - 增加 `.gitignore`、`.env.example`、Makefile 和基础 README；
   - 约定 lint、format、test、build、run、migrate 命令。
2. 搭建基础设施：
   - Docker Compose 启动 PostgreSQL；
   - 配置持久卷、健康检查和仅用于本地开发的账号；
   - 服务启动时建立连接池并执行 ping，但不自动修改 schema；
   - 设置连接池最大连接数、空闲连接数、连接寿命和查询超时。
3. 建立数据库版本管理：
   - 选定 migration 工具；
   - 创建 `schema_migrations` 管理流程；
   - 提供 `make migrate-up`、`make migrate-down`、`make migrate-create`；
   - 验证空库升级、完整回滚、再次升级。
4. 建立 HTTP 基础能力：
   - Gin Router、JSON 统一响应、错误映射；
   - request ID、访问日志、panic recovery 和请求超时中间件；
   - `/health/live` 只表示进程存活，`/health/ready` 检查必要依赖；
   - 设置服务端 read/write/idle/header timeout。
5. 建立日志和退出流程：
   - zap 输出结构化日志，至少包含 level、time、request_id、method、path、status、latency；
   - 捕获 SIGINT/SIGTERM，停止接收请求，等待在途请求，关闭数据库和 logger；
   - 配置错误时快速失败并给出可操作的日志。
6. 建立最低质量门槛：
   - 添加基础单元测试；
   - 执行 `go test ./...`、`go vet ./...`、`go build ./...`；
   - CI 暂可只运行格式检查、测试和构建，完整流水线在 Phase 6/7 固化。

### 交付物

- 可编译的服务骨架；
- `docker-compose.yml`、Dockerfile、migration、Makefile、`.env.example`；
- `docs/phase0-bootstrap.md`，说明启动链路和一次 HTTP 请求的完整路径。

### 阶段门禁

- 新环境可依次执行 `docker compose up -d`、迁移、启动服务和 curl 健康检查；
- PostgreSQL 未就绪时 readiness 失败且错误明确；
- SIGTERM 后服务在设定宽限期内干净退出；
- 不含 Redis、Kafka、gRPC、etcd 或微服务代码。

---

## Phase 1：PostgreSQL 单体 MVP

### 阶段目标

完成“建用户 → 建商品 → 建活动/库存 → 秒杀 → 扣库存并创建订单 → 查询/取消订单”的最小闭环，先保证事务原子性，不解决高并发争用。

### 执行步骤

1. 固化领域和状态机：
   - User、Item、Activity、Inventory、Order、SeckillRecord；
   - 明确 Activity 和 Order 的合法状态转换；
   - 定义价格、数量、活动时间、用户状态等业务校验；
   - 决定取消订单是否回补库存，并把规则写入领域文档。
2. 设计 schema 和约束：
   - 为所有表创建 migration、主键、外键、非空约束和时间字段；
   - `activities.item_id`、`inventories.activity_id` 建立必要索引/唯一约束；
   - `orders.order_no` 唯一；`seckill_records(user_id, activity_id)` 唯一；
   - Inventory 包含 `total`、`available`、`sold`、`version` 和检查约束；
   - 写 `docs/database.md`，记录字段语义和索引理由。
3. 按模块实现固定调用链：
   - model/dto/errors → repository → service → handler；
   - repository 方法接收 `context.Context`；
   - service 承担业务规则与事务边界，handler 不写核心业务；
   - 数据库错误转换成稳定的领域错误。
4. 实现首批 API：
   - `POST /api/v1/users`、`POST /api/v1/login`、`GET /api/v1/users/me`；
   - `POST/GET /api/v1/items`；
   - `POST/GET /api/v1/activities`；
   - `POST /api/v1/activities/:id/seckill`；
   - `GET /api/v1/orders`、`GET /api/v1/orders/:id`、`POST /api/v1/orders/:id/cancel`；
   - 列表 API 增加明确的分页、排序上限。
5. 认证采用够用的安全实现：
   - 密码只保存强哈希；
   - 登录令牌密钥从环境变量读取；
   - 管理类 API 与普通用户 API 区分权限；
   - 不在这一阶段扩展 OAuth、复杂 RBAC。
6. 实现“故意朴素”的秒杀：
   - 读取 Inventory，检查库存，再更新库存并创建 Order/SeckillRecord；
   - 扣库存、更新 sold、创建订单和秒杀记录必须处于同一数据库事务；
   - 保留 read-modify-write 竞争窗口，为 Phase 2 复现超卖；
   - 不使用 Redis、本地锁或隐藏的并发优化。
7. 建立测试：
   - service 单测覆盖状态、库存不足、重复用户、非法数量、活动未开始/已结束；
   - PostgreSQL 集成测试覆盖事务提交/回滚、唯一约束和外键；
   - E2E 覆盖完整业务闭环；
   - API 测试覆盖 400/401/403/404/409/500 等错误映射。
8. 执行故障实验：
   - 扣库存后、订单插入前注入错误，验证全事务回滚；
   - 订单插入故意违反约束，验证库存不变化；
   - 关闭 PostgreSQL，记录超时、HTTP 响应、日志和恢复行为；
   - 每次实验写简短 postmortem。

### 交付物

- 单体 MVP 和完整 migration；
- API 示例/接口文档；
- `docs/phase1-monolith.md`；
- `docs/postmortems/phase1-*.md`。

### 阶段门禁

- 串行请求下业务闭环正确；
- 任一事务步骤失败都不会留下部分更新；
- 单元、集成、E2E 测试全部通过；
- 能用日志中的 request_id 串起单个请求；
- 明确承认并保留并发缺陷，禁止提前加 Redis/Kafka。

---

## Phase 2：单机并发、数据库锁与 PostgreSQL 实验

### 阶段目标

用相同负载先复现超卖，再依次比较悲观锁、条件 UPDATE 和乐观锁，选出主链路方案；同时掌握 PostgreSQL 的执行计划、MVCC、VACUUM 和 WAL。

### 执行步骤

1. 建立可重复基线：
   - 编写 k6 数据准备和秒杀脚本；
   - 固定场景：1000 个不同用户抢 100 份库存；
   - 清理并重建测试数据，确保每轮起点一致；
   - 记录硬件、容器资源、连接池、VUs、duration 和阈值；
   - 收集总请求、成功/失败、QPS、P50/P95/P99、库存、订单数和错误率。
2. 复现并解释超卖：
   - 对 Phase 1 朴素方案重复运行，扩大竞争窗口以稳定复现；
   - 保存订单数量、最终库存和相关 SQL/log；
   - 画出两个事务的 read-modify-write 交错时间线；
   - 将这组结果作为所有后续方案的 baseline。
3. 实现方案 A——悲观锁：
   - 在事务内使用 `SELECT ... FOR UPDATE`；
   - 检查锁范围、锁等待和事务持有时间；
   - 设置合理的 statement/lock timeout；
   - 增加死锁或锁等待实验，确认失败行为可控；
   - 运行相同 k6 场景并验证不变量。
4. 实现方案 B——条件原子 UPDATE：
   - 使用 `UPDATE ... SET available=available-$n, sold=sold+$n WHERE ... AND available >= $n`；
   - 以 affected rows 判定售罄/竞争失败；
   - 将订单创建放在同一事务；
   - 使用相同数据和负载测试正确性与吞吐。
5. 实现方案 C——乐观锁：
   - 基于 `version` 做 CAS；
   - 设定最大重试次数、退避和冲突错误；
   - 分别运行低冲突与热点高冲突测试；
   - 记录重试次数、失败率和尾延迟，防止无限自旋。
6. 决策并固化：
   - 对三种方案比较正确性、QPS、P95/P99、DB CPU、锁等待、复杂度；
   - 默认候选为条件原子 UPDATE，但必须以实测和业务约束做最终决定；
   - 用 ADR 记录选型和未选方案适用场景；
   - 保留独立实验实现或 tag，主分支只保留最终清晰实现。
7. 完成订单索引实验：
   - 生成约 100 万条分布合理的订单测试数据；
   - 对用户订单分页查询运行 `EXPLAIN (ANALYZE, BUFFERS)`；
   - 建立 `(user_id, created_at DESC)` 联合索引后重测；
   - 记录 Seq/Index/Bitmap Scan、actual time、rows、loops、shared hit/read；
   - 验证索引对写入和磁盘空间的代价。
8. 完成 PostgreSQL 专项实验：
   - MVCC：两个 session 对比 READ COMMITTED 和 REPEATABLE READ 快照；
   - Dead Tuple：批量 UPDATE/DELETE，查看 `pg_stat_user_tables`，执行 VACUUM 后比较；
   - WAL：观察 WAL LSN/checkpoint，解释提交、数据页刷盘与崩溃恢复；
   - 索引类型：在隔离实验表比较 B-Tree、GIN、GiST、BRIN、Partial、Expression；
   - MySQL 对比只写入 `docs/database/mysql-vs-postgresql.md`，不更换运行数据库。

### 交付物

- `tests/load/phase2-seckill.js` 和数据生成脚本；
- `docs/phase2-concurrency-report.md`；
- `docs/database/postgresql-internals.md`；
- `docs/database/mysql-vs-postgresql.md`；
- 库存并发控制 ADR。

### 阶段门禁

- 1000 抢 100 的多轮测试中成功数量不超过 100、库存不为负、一人一单成立；
- 三种并发控制方案使用相同环境和脚本完成横向数据对比；
- 能解释最终方案为何正确，以及在热点竞争下的瓶颈；
- 先证明 PostgreSQL 成为瓶颈，才进入 Redis 阶段。

---

## Phase 3A：Redis 活动缓存

### 阶段目标

解决热点活动详情读取造成的数据库压力，并用实验处理缓存穿透、击穿、雪崩和一致性问题。

### 执行步骤

1. 用“热门活动详情高频刷新”建立无缓存基线，记录 DB QPS、连接池饱和、CPU 和 P95/P99；
2. Docker Compose 增加 Redis，封装带 timeout 的客户端和健康指标；
3. 实现 `activity:{id}` Cache Aside：读 miss 回源并写缓存，活动更新后删除缓存；
4. TTL 基线设为 5 分钟并加入随机 jitter，序列化格式增加版本意识；
5. 对不存在的活动做短 TTL 空值缓存，压测随机无效 ID 验证穿透缓解；
6. 对单热点 key 使用 Go `singleflight` 合并同实例回源；多实例限制需在报告中说明；
7. 制造同 TTL 批量过期，比较无 jitter 和有 jitter 的 DB 峰值；
8. 讨论并测试 DB 更新成功但删缓存失败、并发读写造成旧值回填的窗口；
9. 暴露 cache hit/miss、回源次数、Redis 延迟和错误指标；
10. Redis 故障时采用有界降级：限制 DB 回源并避免缓存故障直接放大为 DB 雪崩。

### 交付物与门禁

- `docs/phase3-cache.md` 和缓存故障 postmortem；
- 相同读负载下提供优化前后数据，命中率和 DB QPS 变化可观测；
- 穿透、击穿、雪崩均有可复现脚本和结果；
- 缓存不作为业务正确性的唯一来源。

---

## Phase 3B：Redis Lua 库存预扣与幂等入口

### 阶段目标

把海量竞争请求尽早挡在 PostgreSQL 之前，通过 Lua 原子完成活动校验、库存预扣和重复购买校验，同时保留数据库最终防线。

### 执行步骤

1. 设计 Redis key：库存、用户购买集合、活动状态，并使用 hash tag 预留 Cluster 同槽要求；
2. 实现活动发布/开始前预热：从 PostgreSQL 装载活动详情、库存和资格数据；
3. 编写 Lua 脚本原子执行：检查活动状态/时间 → 检查用户标记 → 检查库存 → 扣减 → 写购买标记；
4. 为 Lua 返回稳定结果码：成功、售罄、重复、活动无效、参数错误；
5. 同一用户并发请求 100 次，验证 Redis 层只接受一次；
6. 多用户高并发测试库存不为负，数据库唯一约束继续保留；
7. 设计后续失败补偿所需的 reservation/event 标识，禁止只扣 Redis 而没有可追踪记录；
8. 验证脚本超时、Redis 重启、预热中断和活动重复预热的行为；
9. 记录 Redis 成为新热点后的单 key、容量、持久化和 Cluster 限制。

### 交付物与门禁

- 版本化 Lua 脚本、预热/校验工具和并发测试；
- `docs/phase3-redis-seckill.md`；
- 高并发下 Redis 预扣不超卖、一人一次；
- 明确“Redis 接受”不等于“订单已创建”，API 状态返回“排队中/已受理”。

---

## Phase 3C：限流、Nginx 与多实例

### 阶段目标

建立由入口到用户维度的多层流量保护，并将应用扩为多实例，为分布式问题创造真实环境。

### 执行步骤

1. 依次实现并测试固定窗口、滑动窗口、Token Bucket，比较突刺容忍、公平性和内存开销；
2. 先用 Gin middleware 做单实例 IP/用户限流，响应返回 429 和可选 `Retry-After`；
3. 扩为 Redis Lua 分布式限流，确保多实例共享配额；
4. 引入 Nginx 反向代理，配置 upstream、keepalive、超时、访问日志和请求体上限；
5. 启动 3 个 Go 实例并验证负载分配、健康摘除和请求 ID 传播；
6. 压测单 IP 10000 req/s、正常多用户流量和突发流量；
7. 确定 Nginx 粗粒度、Gateway/API 级、用户级限流的职责，避免重复规则不可控；
8. 检查任何进程内锁、计数器、singleflight 在多实例下的语义变化并记录。

### 交付物与门禁

- Nginx 配置、限流实现和负载脚本；
- `docs/phase3-rate-limit-nginx.md`；
- 超限请求尽早失败，合法流量错误率和延迟在设定阈值内；
- 任一应用实例停止后，入口能摘除故障实例并继续服务。

---

## Phase 3D：Kafka 异步订单与削峰

### 阶段目标

将 Redis 预扣后的订单持久化从同步 HTTP 链路移到 Kafka，建立 at-least-once 下的幂等消费和积压处理能力。

### 执行步骤

1. 建立同步链路性能基线，证明订单/日志/统计等操作拉长响应或压垮 DB；
2. 设计 `seckill.order.create` topic、partition key、保留期、消息 schema 和版本字段；
3. 消息至少包含 `event_id`、user/activity、quantity、request/trace ID、occurred_at、schema_version；
4. Producer 仅在 Lua 接受后发布，并明确 ack、超时、重试和发送失败处理；
5. Consumer Group 消费，在 PostgreSQL 事务内创建订单和必要业务记录；
6. `event_id` 建唯一键，`(user_id, activity_id)` 继续作为业务唯一防线；
7. 订单提交成功后再提交 offset；重复投递返回成功但不重复写入；
8. 注入“DB 已提交、offset 未提交、进程崩溃”，验证重启后的幂等；
9. 停止消费者持续生产，观察 lag、恢复速度和积压告警；
10. 制造 producer 10000 msg/s、consumer 1000 msg/s，比较增加 partition/consumer、批写等策略；
11. 定义不可重试消息的隔离/DLQ 策略以及人工重放工具；
12. 暂时记录 Redis 已扣但 Kafka 发布失败的一致性缺口，在 Phase 4 用可靠事件/补偿解决。

### 交付物与门禁

- Kafka schema/producer/consumer、故障脚本；
- `docs/phase3-kafka.md`；
- 重复消息不会生成重复订单；
- consumer 停机后消息可恢复处理，lag 有指标和阈值；
- HTTP 快速返回“排队中”，另有订单状态查询形成用户闭环。

---

## Phase 4：分布式正确性、可靠消息与 gRPC

### 阶段目标

在多实例条件下处理本地状态失效、全局 ID、一致性和网络调用失败，并在正式拆微服务前抽出 Inventory RPC 边界。

### 执行步骤

1. 多实例状态审计：
   - 列出 mutex、内存缓存、计数器、定时任务、ID 生成器；
   - 分别验证它们在 3 实例下是否失效；
   - 优先使用唯一约束、原子操作或无共享设计，确有必要才使用分布式锁。
2. 分布式锁实验：
   - 实现 `SET key token NX PX`，释放时用 Lua 比对 owner token；
   - 制造业务超过租期、客户端暂停、误删和 Redis 切换；
   - 研究 watchdog、Redlock、fencing token，但只把适合当前场景的最小方案用于代码；
   - 输出“何时不应使用分布式锁”的 ADR。
3. 全局唯一 ID：
   - 比较 UUID、数据库序列、Redis INCR、Snowflake-like；
   - 实现并发安全的 Snowflake-like 生成器；
   - 处理机器 ID 分配、同毫秒序列耗尽和时钟回拨；
   - 用并发测试验证唯一性、排序趋势和重启行为。
4. 可靠事件与最终一致性：
   - 先精确定义 Redis 预扣、Kafka 发送、PostgreSQL 订单间的失败矩阵；
   - 实现可重试事件状态/Outbox 或等价可靠投递设计；
   - 注意传统 Outbox 只能原子覆盖同一 PostgreSQL 事务，无法天然把 Redis Lua 与 DB 事务合并；
   - 为“Redis 已扣、消息未可靠落地”增加 reservation 日志、重试/回补和超时扫描；
   - 实现 reconciliation worker，对比 Redis 库存、成功订单和事件状态，差异告警后再按安全规则修复；
   - 所有补偿操作本身必须幂等且有审计记录。
5. 抽出 Inventory gRPC：
   - 先定义 proto 和兼容性规则，再生成 client/server；
   - 提供 GetInventory、DeductInventory，并为写 RPC 加幂等键；
   - 传播 context deadline、request_id/trace metadata 和结构化 status；
   - 配置连接复用、最大消息、keepalive 和健康检查。
6. RPC 故障实验：
   - Inventory 延迟 3 秒、调用方 deadline 500ms，验证资源及时释放；
   - kill Inventory，记录错误传播和恢复；
   - 对可重试错误做有限次数指数退避+jitter；
   - 人为制造响应丢失后重试，验证扣库存不执行两次；
   - 记录 retry storm 风险。
7. CAP/一致性文档：
   - 使用 Redis 分区、服务发现不可达、订单与库存暂时不一致三个项目场景解释选择；
   - 对 2PC、3PC、TCC、Saga、Outbox 做比较，但不为展示技术而全部实现。

### 交付物

- proto、Inventory RPC 服务和客户端；
- Snowflake-like 生成器及测试；
- reconciliation/补偿任务；
- `docs/phase4-distributed.md`、一致性 ADR 和故障 postmortem。

### 阶段门禁

- 多实例下仍满足库存和一人一单不变量；
- Redis 扣减、消息发送、消费和 DB 写入的每个失败点都有检测与收敛路径；
- RPC 超时/重试不会造成重复扣减；
- 对账任务可发现人为注入的不一致，并以可审计方式处理。

---

## Phase 5：微服务拆分与服务治理

### 阶段目标

基于已经验证的边界拆为 gateway、user、activity、inventory、order 服务，引入服务发现、配置治理、弹性和分布式可观测性。

### 执行步骤

1. 拆分前建立基线：整理模块依赖、数据库表归属、同步/异步调用和现有 SLO；
2. 制定拆分顺序：先迁移已抽出的 Inventory，再 Order、Activity、User，最后收敛 Gateway；每一步都保持可运行；
3. 采用 Monorepo，禁止服务直接导入其他服务的 internal 包；公共包只放真正稳定的横切能力；
4. 明确数据所有权：服务不能绕过 API/RPC 随意写其他服务表；如暂时共享数据库，记录为迁移状态和退出条件；
5. Gateway 对外 REST、内部 gRPC，负责路由、认证、request ID、限流和日志，不承载核心业务；
6. 使用 etcd 注册实例地址和租约，客户端 watch 服务列表并实现负载选择；
7. 模拟实例异常退出和 etcd 短暂不可用，验证租约过期、缓存地址和恢复；
8. 逐步把配置从环境变量归一到版本化配置，敏感信息仍由 Secret 管理；动态配置需校验和审计；
9. 实现客户端超时预算、有限重试、指数退避+jitter；只对明确可重试且幂等的调用启用；
10. 实现 Circuit Breaker 的 CLOSED/OPEN/HALF-OPEN，并用 Inventory 80% 请求延迟 5 秒验证；
11. 为非核心读路径定义降级，例如活动详情在库存异常时返回“库存暂不可用”；核心写路径禁止伪成功；
12. 接入 OpenTelemetry，贯通 Gateway → Order → Inventory → Redis/Kafka；
13. 每个服务暴露 Prometheus 指标：请求量/延迟/错误、goroutine/GC、DB pool、Redis latency、Kafka lag；
14. Grafana 建立 Traffic、Errors、Latency、Saturation 看板和基础告警；
15. 回归 E2E、负载和故障测试，比较拆分前后的延迟、吞吐和故障面。

### 交付物

- 五个服务、proto、Gateway、etcd 集成；
- Prometheus/Grafana/OpenTelemetry/Jaeger 或 Tempo 本地部署；
- `docs/architecture.md`、`docs/phase5-microservices.md`、服务边界 ADR；
- 拆分前后性能与复杂度对比。

### 阶段门禁

- 任一业务请求可从 trace 定位最慢 span；
- 服务实例上下线可自动发现，无需重启调用方；
- 故障服务不会无限占用 goroutine/连接；
- 核心链路 E2E 和不变量测试全部通过；
- 能明确回答“为什么拆、拆分收益、增加的成本、什么情况下应保持单体”。

---

## Phase 6：容量估算、最终架构与 CI/CD

### 阶段目标

把实际测量与架构推导结合，形成“10 万 QPS 抢票系统”的可解释设计，并建立自动化质量流水线。

### 执行步骤

1. 固定设计假设：注册用户 1000 万、同时在线 100 万、峰值 50 万请求/s、库存 5 万；区分入口请求、有效秒杀和最终订单写入；
2. 计算活动详情 read QPS、秒杀入口 QPS、Redis 操作量、Kafka 吞吐、DB 写 QPS、网络带宽和存储增长；
3. 所有估算写出公式、单位、峰值系数、副本和安全余量，禁止只给结论；
4. 用本地/可用环境的单实例实测值推导实例数，并明确该外推的限制；
5. 绘制最终数据流：CDN/Nginx → Gateway → Redis Lua → Kafka → Order Consumer → PostgreSQL → Outbox/对账；
6. 为每层写扩容方式、单点、数据持久性、一致性选择和降级策略；
7. 定义 SLI/SLO 示例：可用性、受理延迟、订单最终生成时间、错误率、lag、对账差异；
8. 建立 CI：format/lint、单元测试、集成测试、构建、依赖/镜像安全检查；
9. 构建多阶段 Docker 镜像，使用非 root 用户、固定版本和最小运行镜像；
10. 在受控环境执行负载测试，不把大规模压测作为每次 PR 的默认步骤；
11. 汇总 `docs/performance.md`，只记录真实数据，估算数据明确标注为估算。

### 交付物与门禁

- `docs/system-design.md`、容量估算表和最终架构图；
- CI 工作流、可复现镜像构建、`docs/performance.md`；
- 能从业务假设逐步推导容量和架构，而非背诵组件列表；
- CI 对正常提交稳定通过，并能阻止测试失败的变更。

---

## Phase 7：Kubernetes 与云原生部署

### 阶段目标

将各服务部署到 Kubernetes，验证健康探针、滚动升级、弹性伸缩和优雅退出。该阶段不负责重新解决业务正确性。

### 执行步骤

1. 为每个服务准备 Deployment、Service、ConfigMap/Secret 引用、资源 requests/limits；
2. 配置 startup/readiness/liveness probe，三者语义分离，避免依赖短暂异常导致反复重启；
3. 配置 Ingress/Gateway 入口以及内部服务访问；
4. 使用受管理的外部状态组件或明确的教学型部署；禁止把生产可靠性假设建立在单副本 Redis/Kafka/PostgreSQL Pod 上；
5. 配置 rolling update、PodDisruptionBudget 和合理 terminationGracePeriod；
6. 删除 Pod 注入 SIGTERM，验证停止接新请求、等待在途请求、停止 consumer、提交/保留正确 offset、关闭连接；
7. 配置 HPA，优先评估 CPU 之外的请求率/队列 lag 指标；
8. 从 1000 QPS 逐步提高到环境可承受上限，观察 2→5→10 副本或实际扩容结果；
9. 验证冷启动、扩容滞后、缩容期间连接与消费分区再均衡；
10. 将部署步骤、回滚方式和常见故障写入 runbook。

### 交付物与门禁

- `deployments/k8s` 清单或 Helm/Kustomize 配置；
- `docs/phase7-kubernetes.md` 和运维 runbook；
- 滚动升级期间达到设定可用性目标且无重复/丢失业务结果；
- HPA 实验有指标截图/数据，不虚构扩容收益；
- 可一键回滚到上一可用镜像版本。

---

## Phase 8：故障演练、恢复与项目收口

### 阶段目标

系统化破坏关键依赖，验证检测、降级、恢复、数据收敛和人员操作流程，并把项目整理成可展示、可复现的成果。

### 执行步骤

1. 先为每个实验写 steady state、爆炸半径、停止条件和恢复步骤；禁止在未知环境直接执行破坏性实验；
2. Kill Redis：观察缓存失效、入口保护、DB 回源和恢复后的缓存重建；
3. Kill Kafka：观察 producer 行为、可靠事件积压、API 语义和恢复后投递；
4. Kill PostgreSQL：观察 consumer backoff、Kafka lag、连接池、恢复后消费；
5. Inventory 增加 3 秒延迟：观察 deadline、retry、breaker、goroutine 和降级；
6. 注入 5% packet loss/连接抖动：观察错误率、尾延迟和重试放大；
7. 注入重复消息、乱序消息和 poison message，验证幂等、状态机和隔离队列；
8. 制造 Redis/订单/事件数量不一致，验证对账发现、告警和补偿；
9. 每次演练记录 Detection、Impact、Root Cause、Resolution、Prevention、Trade-off；
10. 根据事故结果修复监控、阈值、runbook 或代码，然后重复实验验证；
11. 完整运行 unit → integration → E2E → load → chaos 测试矩阵；
12. 更新 README：架构、快速启动、测试命令、演进故事、真实性能结果、已知限制；
13. 整理面试材料：每项技术对应业务问题、代码位置、实测结果和取舍；
14. 最终安全检查：示例配置无真实密钥、日志无敏感信息、依赖和镜像无已知高危问题。

### 交付物与门禁

- `tests/chaos`、所有 postmortem、故障矩阵和恢复 runbook；
- 完整 README、架构/数据库/性能文档；
- 五类测试均可重复运行；
- 所有已知数据差异均可检测且有恢复路径；
- 简历描述中的每个数字都能指向脚本、环境和报告。

---

## 4. 推荐提交与里程碑

保持每次提交可构建、可测试，建议按以下里程碑推进：

1. `feat: initialize flashsale monolith`：Phase 0 骨架；
2. `feat: implement postgres monolith mvp`：Phase 1 主链路；
3. `test: reproduce inventory overselling`：Phase 2 基线；
4. `feat: enforce atomic inventory deduction`：数据库并发方案；
5. `perf: add activity cache and cache protections`：Phase 3A；
6. `feat: reserve seckill inventory with redis lua`：Phase 3B；
7. `feat: add distributed rate limiting and nginx`：Phase 3C；
8. `feat: create orders asynchronously with kafka`：Phase 3D；
9. `feat: add reliable events and reconciliation`：Phase 4 一致性；
10. `feat: extract inventory grpc service`：Phase 4 RPC；
11. `refactor: split monolith into services`：Phase 5；
12. `feat: add observability and resilience`：Phase 5 治理；
13. `ci: add build test and image pipeline`：Phase 6；
14. `deploy: add kubernetes manifests`：Phase 7；
15. `test: add chaos scenarios and runbooks`：Phase 8。

## 5. 每周执行节奏

不绑定绝对工期，以“通过门禁”为完成标准。每个工作周期按以下节奏：

1. 选取当前 Phase 的一个问题，写清验收条件和基线；
2. 只实现能解决该问题的最小改动；
3. 先跑正确性测试，再跑性能/故障测试；
4. 保存原始结果，更新报告和 ADR；
5. 执行全量回归，提交一个可独立解释的变更；
6. 周末复盘：本周解决了什么、数据是否支持结论、新问题是什么、是否满足阶段门禁。

## 6. 开始执行时的前 10 个任务

1. 确认 Go module 名称和锁定 Go 版本；
2. 创建 Phase 0 目录、Makefile、`.env.example`；
3. 创建 PostgreSQL Docker Compose 服务和健康检查；
4. 选定并接入 migration 工具；
5. 实现配置、zap logger、数据库连接池；
6. 实现 Gin router、request ID、recovery、timeout、统一响应；
7. 实现 live/readiness health check；
8. 实现 SIGTERM graceful shutdown；
9. 加入 build/test/vet 的本地命令和最小 CI；
10. 从空环境完整演练启动流程，记录到 `docs/phase0-bootstrap.md`，通过 Phase 0 门禁后再开始 User/Item/Activity 表。

## 7. 明确延期项

以下内容只有在前述阶段出现真实需求时才加入：Bloom Filter、Redis Cluster、Redlock、CDC、完整配置中心、Payment 服务、Service Mesh、CQRS、Event Sourcing、多地域容灾。它们可以写研究笔记，但不能阻塞主线，也不能仅为丰富技术栈进入生产代码。
