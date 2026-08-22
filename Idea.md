# Go 高并发秒杀系统 ------ 项目需求与学习路线（Idea.md）

> 项目定位：只维护 **1
> 个主项目**，从最简单的单体后端开始，持续"制造问题---解决问题---压测验证---总结权衡"，最终演进为一个能够覆盖后端核心面试知识的高并发秒杀系统。
>
> 业务场景：演唱会抢票、限量商品抢购、活动预约、优惠券抢领、秒杀活动等典型高并发场景。
>
> 核心目标不是"堆技术栈"，而是让每一个技术都能回答四个问题：
>
> 1.  **为什么需要它？**
> 2.  **它解决了项目里的什么具体问题？**
> 3.  **代码放在哪里、怎么实现？**
> 4.  **它带来了什么新问题和 trade-off？**

---

# 1. 项目总目标

本项目最终要训练的不是"会写一个秒杀 Demo"，而是建立完整的后端工程能力：

```text
业务需求
  ↓
API 设计
  ↓
数据建模
  ↓
最小可用实现
  ↓
并发问题
  ↓
性能瓶颈
  ↓
缓存 / 消息队列 / 限流
  ↓
分布式一致性
  ↓
微服务治理
  ↓
可观测性
  ↓
压测与故障演练
  ↓
系统设计与面试表达
```

项目始终遵循同一个学习循环：

```text
① 先有业务需求
        ↓
② 写最简单实现
        ↓
③ 发现问题
        ↓
④ 学一个技术解决
        ↓
⑤ 写代码
        ↓
⑥ 压测 / 制造故障
        ↓
⑦ 看方案是否有效
        ↓
⑧ 总结 trade-off
```

任何新技术都不允许因为"教程里有"而加入。

只有出现了具体问题，才允许引入对应方案。

---

# 2. 推荐技术栈

## 2.1 总体选择原则

技术选择优先级：

1.  后端面试高频；
2.  真实工程中常见；
3.  Go 生态成熟；
4.  能够自然支撑项目逐阶段演进；
5.  不为了炫技引入冷门组件。

## 2.2 推荐主线

---

层级 推荐技术 使用阶段 学习重点

---

语言 Go 全程 goroutine、channel、context、error、interface、并发

Web Gin Phase 1 起 HTTP、REST、Middleware、参数校验

数据库 PostgreSQL Phase 1 起 SQL、事务、索引、锁、MVCC

ORM GORM Phase 1 起 CRUD、事务、工程开发效率

SQL 原生 SQL Phase 2 起 Explain、锁、库存扣减、性能优化

Migration golang-migrate / Phase 1 起 Schema 版本管理
Goose（二选一）

日志 zap Phase 1 起 结构化日志、request_id

配置 env → Viper Phase 1/后期 配置分环境、配置管理

缓存 Redis Phase 3 起 缓存、原子操作、Lua、分布式锁

MQ Kafka Phase 3 起 异步、削峰、幂等、可靠消息

反向代理 Nginx Phase 3 起 LB、反向代理、限流

RPC gRPC + Protobuf Phase 4 起 RPC、IDL、序列化、超时

服务发现 etcd Phase 4/5 注册发现、租约、Raft 延伸

API Gateway Nginx → 独立 Phase 5 路由、鉴权、限流、治理
Gateway

Metrics Prometheus Phase 3 起 QPS、延迟、错误率、资源指标

Dashboard Grafana Phase 3 起 可视化

Tracing OpenTelemetry Phase 5 Trace、Span、上下文传播

Trace Backend Jaeger / Tempo Phase 5 分布式链路追踪

压测 k6 Phase 2 起 QPS、P95、P99、吞吐

容器 Docker + Docker Phase 1 起 本地基础设施
Compose

编排 Kubernetes Phase 5/6 Deployment、Service、HPA

CI/CD GitHub Actions Phase 6 测试、构建、镜像、部署

---

---

# 3. 为什么选择这些技术

## 3.1 为什么数据库选择 PostgreSQL

这个项目的主数据库改为 PostgreSQL。

数据库学习重点不只是 CRUD，而是围绕 PostgreSQL 的真实实现理解：

- PostgreSQL MVCC；
- Tuple / Row Version；
- xmin / xmax；
- Snapshot；
- WAL（Write-Ahead Logging）；
- Checkpoint；
- VACUUM / Autovacuum；
- Dead Tuple；
- Transaction ID 与 Wraparound；
- B-Tree Index；
- Hash Index；
- GIN；
- GiST；
- BRIN；
- Partial Index；
- Expression Index；
- Composite Index；
- Index Only Scan；
- Heap Fetch；
- EXPLAIN；
- EXPLAIN ANALYZE；
- Seq Scan / Index Scan / Bitmap Scan；
- Planner / Cost；
- Row Lock；
- SELECT ... FOR UPDATE；
- 乐观锁；
- 悲观锁；
- PostgreSQL 默认 Read Committed；
- Repeatable Read；
- Serializable / SSI；
- Deadlock；
- Connection / Connection Pool；
- WAL 与复制基础。

同时保留一个独立的 **MySQL 面试补充模块**，用于横向比较：

- InnoDB；
- 聚簇索引与二级索引；
- 回表；
- Undo Log / Redo Log；
- MySQL MVCC；
- Read View；
- Gap Lock；
- Next-Key Lock；
- MySQL 默认 Repeatable Read；
- PostgreSQL 与 MySQL 在事务、MVCC、锁、索引和日志机制上的差异。

目标是做到：

> 项目真正使用 PostgreSQL，但面试遇到 MySQL/InnoDB
> 高频题仍然能够回答，并且能够进行 PostgreSQL vs MySQL 的原理对比。

秒杀中的：

```text
查询库存
↓
检查库存
↓
扣减库存
↓
创建订单
```

天然适合学习事务和并发控制。

---

## 3.2 为什么 Gin

Gin 足够轻量。

项目重点不是学习复杂 Web Framework，而是理解：

```text
HTTP Request
    ↓
Router
    ↓
Middleware
    ↓
Handler
    ↓
Service
    ↓
Repository
    ↓
PostgreSQL
```

后续即使换 Echo、Fiber、Chi，也不改变核心思想。

---

## 3.3 为什么 GORM + 手写 SQL

前期使用 GORM 是为了快速建立工程结构：

```text
model
repository
service
handler
```

但是涉及：

- 秒杀库存扣减；
- 行锁；
- 乐观锁；
- Explain；
- 索引；
- 高性能查询；

必须主动写 SQL。

目标不是成为"只会 ORM"的后端开发。

---

## 3.4 为什么 Redis

Redis 在本项目中不会一开始加入。

只有当出现：

- DB 查询压力；
- 热点活动；
- 热点库存；
- 高并发库存竞争；
- 分布式锁；
- 限流；

才加入。

需要真正掌握：

```text
Cache Aside
TTL
缓存穿透
缓存击穿
缓存雪崩
String / Hash / Set / ZSet
Lua
原子性
Pipeline
分布式锁
Redis Cluster
```

---

## 3.5 为什么 Kafka

Kafka 不用于"项目看起来高级"。

它解决两个非常明确的问题：

### 问题 1：同步链路过长

```text
用户秒杀
↓
创建订单
↓
扣库存
↓
记录日志
↓
发送通知
↓
更新活动统计
↓
返回用户
```

大量非核心操作没有必要阻塞用户请求。

### 问题 2：瞬时流量太高

```text
100000 Request/s
        ↓
        MQ
        ↓
Consumer 按自身能力处理
```

Kafka 用来学习：

- Producer；
- Broker；
- Topic；
- Partition；
- Consumer Group；
- Offset；
- At-most-once；
- At-least-once；
- 幂等；
- 重复消费；
- 消息丢失；
- 消息积压；
- 顺序；
- 延迟；
- Dead Letter 思想。

---

## 3.6 为什么 gRPC

微服务阶段需要真实体验：

```go
service.Method()
```

变成：

```text
Service A
   ↓ network
Service B
```

之后自然出现：

- 网络延迟；
- 超时；
- 重试；
- 服务不可用；
- 序列化；
- 服务发现；
- 连接池；
- 熔断；
- Trace Context。

所以 gRPC 不只是一个 RPC 框架，而是学习分布式调用链的入口。

---

# 4. 项目业务模型

项目统一命名：

```text
FlashSale
```

中文：

```text
高并发秒杀 / 抢票平台
```

系统支持四类业务，但底层统一抽象：

1.  演唱会门票抢购；
2.  限量商品秒杀；
3.  活动名额预约；
4.  限量优惠券抢领。

核心抽象统一为：

```text
活动 Activity
+
资源 Item
+
库存 Inventory
+
用户 User
+
订单 Order
```

---

# 5. 核心领域对象

## User

```text
id
username
email
password_hash
status
created_at
updated_at
```

## Item

代表：

- 演唱会门票；
- 商品；
- 活动名额；
- 优惠券。

```text
id
name
type
description
price
status
created_at
updated_at
```

## Activity

```text
id
item_id
name
start_time
end_time
status
total_stock
created_at
updated_at
```

## Inventory

```text
id
activity_id
total
available
sold
version
updated_at
```

其中：

```text
version
```

后续用于实现乐观锁。

## Order

```text
id
order_no
user_id
activity_id
quantity
amount
status
created_at
updated_at
```

订单状态：

```text
PENDING
PAID
CANCELLED
EXPIRED
```

## SeckillRecord

记录用户秒杀行为：

```text
id
user_id
activity_id
order_id
status
created_at
```

用于后续实现：

```text
同一用户同一活动只能抢一次
```

---

# 6. 初始 API

## 用户

```http
POST /api/v1/users
POST /api/v1/login
GET  /api/v1/users/me
```

## 商品 / 门票

```http
POST /api/v1/items
GET  /api/v1/items/:id
GET  /api/v1/items
```

## 秒杀活动

```http
POST /api/v1/activities
GET  /api/v1/activities/:id
GET  /api/v1/activities
```

## 秒杀

```http
POST /api/v1/activities/:id/seckill
```

请求：

```json
{
  "quantity": 1
}
```

## 订单

```http
GET /api/v1/orders/:id
GET /api/v1/orders
POST /api/v1/orders/:id/cancel
```

---

# 7. 项目目录总原则

一开始不要照搬大型互联网公司的复杂架构。

Phase 1 使用清晰的单体分层。

推荐：

```text
flashsale/
│
├── cmd/
│   └── server/
│       └── main.go
│
├── internal/
│   ├── user/
│   ├── item/
│   ├── activity/
│   ├── inventory/
│   └── order/
│
├── pkg/
│
├── configs/
│
├── migrations/
│
├── scripts/
│
├── tests/
│
├── deployments/
│
├── docs/
│
├── go.mod
├── go.sum
├── Makefile
├── docker-compose.yml
├── Dockerfile
└── README.md
```

每个业务模块：

```text
internal/activity/
├── handler.go
├── service.go
├── repository.go
├── model.go
├── dto.go
└── errors.go
```

形成固定认知：

```text
handler
   ↓
service
   ↓
repository
   ↓
database
```

---

# 8. Phase 0 ------ 工程准备

## 目标

先确保自己会独立创建一个 Go 后端项目，而不是开始研究高并发。

必须做到：

```text
git clone
↓
docker compose up
↓
make migrate
↓
go run ./cmd/server
↓
curl API
```

可以完整启动项目。

## 技术

```text
Go
Gin
PostgreSQL
GORM
Docker Compose
zap
Migration Tool
Makefile
```

## 要学习

### Go 工程基础

- module；
- package；
- internal；
- cmd；
- dependency injection；
- interface；
- context；
- error handling；
- graceful shutdown。

### HTTP

- method；
- status code；
- header；
- JSON；
- middleware；
- request context；
- timeout。

## 第一版目录

```text
flashsale/
├── cmd/server/main.go
├── internal/
│   ├── user/
│   ├── item/
│   ├── activity/
│   ├── inventory/
│   └── order/
├── pkg/
│   ├── database/
│   ├── logger/
│   └── response/
├── configs/
├── migrations/
└── docker-compose.yml
```

## 验收

能够独立回答：

> 一个 HTTP 请求进入你的 Go 项目之后经历了什么？

回答应该能从：

```text
TCP
→ HTTP
→ Gin Router
→ Middleware
→ Handler
→ Service
→ Repository
→ PostgreSQL
→ Response
```

一路解释。

---

# 9. Phase 1 ------ 最简单可用版本：单体秒杀系统

> 本阶段禁止 Redis、Kafka、微服务、分布式锁。

## 9.1 业务目标

完成完整业务：

```text
创建用户
↓
创建商品
↓
创建秒杀活动
↓
初始化库存
↓
用户秒杀
↓
库存减少
↓
生成订单
↓
查询订单
```

---

## 9.2 秒杀最简单实现

伪代码：

```go
func (s *Service) Seckill(ctx context.Context, userID, activityID int64) error {
    inventory, err := s.inventoryRepo.Get(ctx, activityID)
    if err != nil {
        return err
    }

    if inventory.Available <= 0 {
        return ErrSoldOut
    }

    inventory.Available--

    if err := s.inventoryRepo.Update(ctx, inventory); err != nil {
        return err
    }

    return s.orderRepo.Create(ctx, ...)
}
```

故意保留问题。

不要提前解决。

---

## 9.3 本阶段学习目标

### API

理解：

```text
REST
DTO
参数校验
统一错误响应
错误码
Middleware
Request ID
```

### DB

学习：

```text
表设计
主键
外键
唯一索引
普通索引
联合索引
SQL
Explain
```

### 事务

让：

```text
扣库存
+
创建订单
```

在一个事务中。

学习：

```text
BEGIN
COMMIT
ROLLBACK
ACID
```

---

## 9.4 第一批故障实验

主动制造：

### Experiment A

创建订单成功前，让程序 panic。

观察：

```text
库存是否已经扣减？
```

### Experiment B

扣库存成功后手动让订单 Insert 失败。

观察：

```text
库存和订单是否一致？
```

### Experiment C

关闭 PostgreSQL。

观察：

```text
API 返回什么？
日志记录什么？
连接什么时候超时？
```

---

## 9.5 Phase 1 面试表达

必须能够讲：

> 第一版系统是一个 Go 单体应用，请求通过 Gin Handler 进入 Service
> 层，Service 负责秒杀业务，Repository 使用 PostgreSQL
> 持久化。库存扣减和订单创建需要满足原子性，所以放在同一个事务里。如果订单创建失败整个事务回滚。

---

# 10. Phase 2 ------ 单机并发：故意制造超卖

这一阶段开始真正进入后端核心。

学习路线：

```text
HTTP
→ API
→ DB
→ 事务
→ 索引
→ 并发
```

---

# 11. Experiment：1000 人抢 100 张票

库存：

```text
100
```

并发请求：

```text
1000
```

理论成功订单：

```text
100
```

第一次压测。

使用：

```text
k6
```

记录：

```text
总请求数
成功数
失败数
库存
订单数
QPS
P50
P95
P99
```

极有可能发现：

```text
订单数量 > 100
```

或者库存数据发生覆盖。

这就是：

```text
超卖
```

---

# 12. 为什么发生超卖

两个请求：

```text
Request A
    ↓
SELECT stock = 1

Request B
    ↓
SELECT stock = 1
```

然后：

```text
A → stock = 0
B → stock = 0
```

两个用户都成功。

核心问题：

```text
read-modify-write
```

不是原子操作。

---

# 13. Solution 1 ------ 悲观锁

使用：

```sql
SELECT available
FROM inventories
WHERE activity_id = $1
FOR UPDATE;
```

学习：

```text
行锁
锁等待
事务
死锁
隔离级别
```

然后重新压测。

记录：

```text
是否还有超卖
QPS
P95
DB CPU
锁等待
```

---

# 14. Solution 2 ------ 条件 UPDATE

进一步优化：

```sql
UPDATE inventories
SET available = available - 1
WHERE activity_id = $1
  AND available > 0;
```

根据 affected rows 判断是否成功。

重点：

```text
把检查 + 修改变成数据库原子操作
```

重新测试。

比较：

```text
SELECT FOR UPDATE
vs
条件 UPDATE
```

---

# 15. Solution 3 ------ 乐观锁

利用：

```text
version
```

SQL：

```sql
UPDATE inventories
SET available = available - 1,
    version = version + 1
WHERE activity_id = $1
  AND available > 0
  AND version = $2;
```

学习：

```text
CAS
乐观锁
冲突重试
```

压测：

```text
低冲突
高冲突
```

比较悲观锁和乐观锁。

---

# 16. Phase 2 索引实验

建立订单查询：

```sql
SELECT *
FROM orders
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT 20;
```

先不建立联合索引。

数据制造：

```text
1,000,000 orders
```

运行：

```text
EXPLAIN
```

然后建立：

```sql
INDEX idx_user_created(user_id, created_at)
```

再次比较。

必须记录：

```text
rows
type
key
Extra
执行时间
```

---

# 16.1 PostgreSQL 专项实验

除了索引实验，本阶段必须专门做 PostgreSQL 内部机制实验。

## Experiment A ------ MVCC

开启两个数据库 Session：

```text
Transaction A
BEGIN
UPDATE inventories ...
不提交

Transaction B
SELECT ...
```

分别在：

```text
READ COMMITTED
REPEATABLE READ
```

下观察读取结果。

目标：

- 理解 Snapshot；
- 理解不同事务看到的 Tuple Version；
- 理解 PostgreSQL MVCC 与"加读锁"不是一回事。

## Experiment B ------ VACUUM / Dead Tuple

对测试表进行大量：

```text
UPDATE
DELETE
```

观察：

```text
pg_stat_user_tables
n_dead_tup
```

然后执行：

```sql
VACUUM;
```

进一步学习：

```text
Autovacuum
Table Bloat
Transaction ID
```

理解 PostgreSQL 为什么不能简单地把 UPDATE 当作"原地修改"。

## Experiment C ------ WAL

观察：

```text
事务提交
↓
WAL
↓
数据页持久化
```

学习：

```text
WAL
Checkpoint
Crash Recovery
Replication
```

重点能够回答：

> 为什么数据库不需要每次事务提交都立刻把所有修改后的数据页刷盘？

## Experiment D ------ EXPLAIN ANALYZE

对订单查询执行：

```sql
EXPLAIN (ANALYZE, BUFFERS)
SELECT *
FROM orders
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT 20;
```

比较建立联合索引前后：

```text
Seq Scan
Index Scan
Bitmap Scan
actual time
rows
loops
shared hit/read
planning time
execution time
```

目标不是只会看"有没有走索引"，而是开始理解 PostgreSQL Planner
为什么选择某个执行计划。

## Experiment E ------ PostgreSQL 索引类型

除 B-Tree 外建立实验表，理解：

```text
B-Tree
GIN
GiST
BRIN
Partial Index
Expression Index
```

不要求秒杀核心链路强行使用所有索引。

要求能够解释：

> 为什么订单查询主要使用 B-Tree，而
> JSONB、全文/数组、范围/地理数据、超大顺序表可能需要其他索引？

---

# 17. Phase 2 输出报告

生成：

```text
docs/phase2-concurrency-report.md
```

记录：

```text
方案
正确性
QPS
P95
P99
实现复杂度
锁竞争
优缺点
适用场景
```

这份报告以后就是面试素材。

---

# 18. Phase 3 ------ 高性能：Redis

此时我们制造新问题：

```text
活动开始前
100 万用户疯狂刷新活动详情
```

接口：

```http
GET /api/v1/activities/:id
```

所有请求都访问 MySQL。

导致：

```text
DB QPS 上升
CPU 上升
连接池耗尽
响应时间上升
```

现在才允许引入 Redis。

---

# 19. Redis Solution 1 ------ Cache Aside

读：

```text
Request
 ↓
Redis
 ↓ miss
PostgreSQL
 ↓
Redis SET
 ↓
Response
```

写：

```text
Update DB
↓
Delete Cache
```

实现活动详情缓存：

```text
activity:{id}
```

TTL：

```text
5 min
```

---

# 20. Redis 必做实验

## 缓存命中率

压测前：

```text
PostgreSQL QPS
```

压测后：

```text
Redis Hit Rate
PostgreSQL QPS
P95
```

比较。

---

## 缓存穿透

请求：

```text
activity_id = random_invalid_id
```

观察大量 DB Miss。

解决：

```text
缓存空值
```

进阶：

```text
Bloom Filter
```

---

## 缓存击穿

制造：

```text
一个超级热门演唱会
```

缓存突然过期。

10000 个请求同时打 PostgreSQL。

解决：

```text
singleflight
```

这是 Go 项目特别适合展示的点。

进阶：

```text
逻辑过期
互斥重建
```

---

## 缓存雪崩

让大量 Key 使用相同 TTL。

然后同时过期。

解决：

```text
TTL + random jitter
```

---

# 21. Redis 库存预扣

现在继续制造问题：

即使数据库防止了超卖，

```text
100000 QPS
```

全部打 PostgreSQL 依旧撑不住。

将秒杀库存预热到 Redis：

```text
seckill:stock:{activityID}
```

---

# 22. Lua 原子扣库存

Lua：

```text
检查库存
↓
判断用户是否已经购买
↓
扣库存
↓
记录用户购买标记
```

作为一个原子操作执行。

学习：

```text
Redis 单线程命令执行模型
Lua
原子操作
幂等
```

---

# 23. 用户重复秒杀

制造：

```text
同一个 user
并发发送 100 次请求
```

要求：

```text
只能成功一次
```

方案演进：

```text
PostgreSQL UNIQUE(user_id, activity_id)
↓
Redis SET / Set
↓
Lua 原子校验
```

最终数据库 UNIQUE 仍然作为最后一道防线。

核心思想：

```text
Redis 提升性能
PostgreSQL 保证最终正确性
```

---

# 24. Phase 3 ------ 限流

制造恶意请求：

```text
单 IP
10000 req/s
```

逐步实现：

### Level 1

Gin Middleware：

```text
固定窗口
```

### Level 2

滑动窗口。

### Level 3

Token Bucket。

### Level 4

Redis 分布式限流。

学习：

```text
Fixed Window
Sliding Window
Leaky Bucket
Token Bucket
```

---

# 25. Phase 3 ------ Nginx

引入：

```text
Client
  ↓
Nginx
  ↓
Go Application
```

学习：

```text
reverse proxy
keepalive
upstream
connection
rate limit
access log
```

先运行一个 Go 实例。

然后：

```text
Go App 1
Go App 2
Go App 3
```

通过 Nginx Load Balance。

---

# 26. Phase 3 ------ MQ：Kafka

制造问题：

Redis 秒杀成功后，如果立即：

```text
写订单
写操作日志
写活动统计
发送通知
```

同步请求仍然很慢。

因此改成：

```text
用户
 ↓
Redis Lua
 ↓
Kafka Producer
 ↓
快速返回
```

Kafka Consumer：

```text
Kafka
 ↓
Order Consumer
 ↓
PostgreSQL
```

---

# 27. Kafka Topic 设计

例如：

```text
seckill.order.create
```

消息：

```json
{
  "event_id": "...",
  "user_id": 10001,
  "activity_id": 20001,
  "quantity": 1,
  "timestamp": 0
}
```

---

# 28. Kafka 必做实验

## 消费者宕机

```text
Producer 正常
Consumer 停止
```

持续产生消息。

观察：

```text
Consumer Lag
```

重新启动 Consumer。

观察：

```text
恢复速度
```

---

## 重复消费

让 Consumer：

```text
订单 INSERT 成功
↓
还没提交 offset
↓
进程崩溃
```

重启。

消息再次消费。

解决：

```text
event_id 唯一键
```

或者：

```text
user_id + activity_id UNIQUE
```

学习：

```text
at-least-once
幂等消费
```

---

## 消息积压

Producer：

```text
10000 msg/s
```

Consumer：

```text
1000 msg/s
```

观察 Lag。

处理：

```text
增加 Partition
增加 Consumer
批量写 DB
```

---

# 29. Phase 3 学习结果

完成：

```text
Redis
→ 缓存
→ Lua
→ 限流
→ Kafka
→ 削峰
→ Nginx
→ Load Balance
```

对应学习阶段：

```text
Redis → 缓存 → MQ → Nginx → 并发 → 限流
```

---

# 30. Phase 4 ------ 分布式

现在已有：

```text
Nginx
  ↓
Go App 1
Go App 2
Go App 3
```

从这里开始，所有：

```text
进程内状态
```

都会成为问题。

---

# 31. 分布式问题 1 ------ 本地锁失效

原来：

```go
var mu sync.Mutex
```

单实例有效。

多个实例：

```text
App 1 mutex
App 2 mutex
App 3 mutex
```

互相完全不知道对方。

这就是学习：

```text
分布式锁
```

的真实原因。

---

# 32. Redis 分布式锁

基础：

```text
SET key value NX PX
```

必须继续研究：

- 锁过期；
- 业务没执行完锁就释放；
- 锁误删；
- UUID owner token；
- watchdog；
- Redis 主从切换；
- Redlock；
- fencing token。

要求：

不要把分布式锁当成"万能锁"。

优先思考：

```text
能不能通过数据库唯一约束？
能不能通过原子操作？
能不能避免共享状态？
```

---

# 33. 分布式问题 2 ------ 全局唯一 ID

订单号不能依赖：

```text
单机自增变量
```

研究方案：

```text
UUID
数据库 auto increment
Redis INCR
Snowflake
```

最终实现：

```text
Snowflake-like ID Generator
```

比较：

```text
唯一性
趋势递增
排序
存储空间
时钟回拨
机器 ID
```

---

# 34. 分布式问题 3 ------ CAP

在真正有多个服务和网络故障之后学习 CAP。

不要死背：

```text
C / A / P
```

需要结合场景：

```text
Redis Cluster 网络分区怎么办？
服务发现节点不可达怎么办？
订单和库存数据短暂不一致怎么办？
```

理解：

```text
Partition 是分布式系统必须考虑的现实
```

真正需要选择的是：

```text
CP vs AP
```

---

# 35. 分布式问题 4 ------ 一致性

Kafka 异步化后：

```text
Redis 已扣库存
↓
Kafka 消息发送失败
↓
MySQL 没有订单
```

产生一致性问题。

研究方案：

```text
重试
本地消息表
Transactional Outbox
CDC
补偿任务
```

---

# 36. Transactional Outbox

订单数据库事务：

```text
BEGIN

INSERT order
INSERT outbox_event

COMMIT
```

后台任务：

```text
scan outbox
↓
publish Kafka
↓
mark sent
```

学习：

```text
本地事务
最终一致性
消息可靠性
```

---

# 37. 分布式事务

学习但不要全部强行实现。

需要理解：

```text
2PC
3PC
TCC
Saga
Outbox
```

项目优先实践：

```text
最终一致性
+
Outbox
+
补偿
+
幂等
```

---

# 38. Phase 4 ------ gRPC

在真正拆微服务之前，先体验 RPC。

从单体中抽：

```text
Inventory Service
```

先做内部服务。

定义：

```protobuf
service InventoryService {
    rpc GetInventory(GetInventoryRequest) returns (GetInventoryResponse);
    rpc DeductInventory(DeductInventoryRequest) returns (DeductInventoryResponse);
}
```

学习：

```text
protobuf
serialization
HTTP/2
deadline
context
metadata
status code
connection
```

---

# 39. RPC 故障实验

## 延迟

Inventory Service：

```text
sleep 3s
```

Order Service：

```text
timeout 500ms
```

观察。

学习：

```text
deadline
timeout
```

---

## 服务宕机

直接 kill Inventory Service。

观察：

```text
错误传播
```

然后引入：

```text
retry
```

再制造新问题：

```text
一次扣库存请求因为重试执行两次
```

学习：

```text
retry 必须结合幂等
```

---

# 40. Phase 4 对应学习路径

```text
负载均衡
→ RPC
→ 分布式锁
→ CAP
→ 一致性
→ 分布式事务
```

---

# 41. Phase 5 ------ 微服务

此时才允许真正拆服务。

目标不是：

```text
“项目有很多文件夹”
```

而是理解：

```text
为什么拆？
拆了之后发生什么？
```

---

# 42. 微服务拆分

拆成：

```text
gateway
user-service
activity-service
inventory-service
order-service
```

暂时不拆 payment。

---

# 43. 微服务代码仓库结构

建议早期仍使用 Monorepo：

```text
flashsale/
├── services/
│   ├── gateway/
│   ├── user/
│   ├── activity/
│   ├── inventory/
│   └── order/
│
├── api/
│   └── proto/
│
├── pkg/
├── deployments/
├── scripts/
└── docs/
```

每个服务：

```text
services/order/
├── cmd/
│   └── main.go
├── internal/
│   ├── handler/
│   ├── service/
│   ├── repository/
│   ├── client/
│   ├── consumer/
│   └── model/
└── config/
```

---

# 44. Gateway

外部：

```text
HTTP REST
```

内部：

```text
gRPC
```

请求：

```text
Client
 ↓
Gateway
 ↓
Order Service
 ↓ gRPC
Inventory Service
```

Gateway 负责：

```text
routing
auth
request id
rate limit
logging
```

不要放核心业务逻辑。

---

# 45. 服务注册与发现

引入：

```text
etcd
```

Service 启动：

```text
register service address
+
lease
```

调用方：

```text
watch service instances
```

学习：

```text
service registry
service discovery
lease
heartbeat
watch
```

进一步延伸：

```text
etcd
↓
Raft
↓
leader election
↓
log replication
↓
quorum
```

---

# 46. 配置中心

前期：

```text
.env
```

单体没有问题。

多个服务后：

```text
每个服务都有配置
```

研究：

```text
central config
dynamic config
version
watch
```

可以利用 etcd 做教学型配置中心。

重点不是做完整 Nacos，而是理解原理。

---

# 47. 熔断

制造：

```text
Inventory Service
80% 请求延迟 5 秒
```

如果 Order Service 不断请求：

```text
goroutine
connection
memory
```

都会堆积。

实现：

```text
Circuit Breaker
```

状态：

```text
CLOSED
OPEN
HALF-OPEN
```

---

# 48. 降级

库存服务异常时：

```text
活动详情仍然可以返回
```

但显示：

```text
库存暂不可用
```

理解：

```text
核心功能
非核心功能
fallback
```

---

# 49. 重试策略

禁止：

```text
无限重试
```

实现：

```text
max retry
+
exponential backoff
+
jitter
```

同时判断：

```text
哪些错误值得 retry？
哪些 API 是 idempotent？
```

---

# 50. 链路追踪

引入：

```text
OpenTelemetry
```

一个请求：

```text
Gateway
  ↓
Order Service
  ↓
Inventory Service
  ↓
Redis
  ↓
Kafka
```

Trace：

```text
trace_id
span_id
parent_span_id
```

必须能够打开 Trace UI，看出：

```text
到底哪个服务慢
```

---

# 51. Metrics

每个服务暴露：

```http
/metrics
```

至少监控：

```text
request_total
request_duration
request_errors
goroutines
GC
DB connections
Redis latency
Kafka lag
```

核心黄金指标：

```text
Traffic
Errors
Latency
Saturation
```

---

# 52. Phase 5 对应学习路径

```text
Gateway
→ 服务发现
→ 配置中心
→ 熔断
→ 降级
→ 重试
→ 链路追踪
```

---

# 53. Phase 6 ------ 系统设计

从此以后：

```text
代码
```

和：

```text
架构设计
```

必须结合。

目标：

面对面试题：

> 设计一个 10 万 QPS 的演唱会抢票系统。

能够自己从零推导。

---

# 54. Capacity Estimation

假设：

```text
注册用户：10,000,000
同时在线：1,000,000
活动开始瞬间请求：500,000 QPS
票数：50,000
```

首先计算：

```text
read QPS
write QPS
bandwidth
storage
Kafka throughput
Redis memory
DB writes
```

不要一上来画架构图。

---

# 55. 最终架构

```text
                    CDN
                     ↓
                  Nginx/LB
                     ↓
                  Gateway
                     ↓
       ┌─────────────┼─────────────┐
       ↓             ↓             ↓
 Activity        Seckill         User
 Service         Service        Service
       ↓             ↓
     Redis       Redis Cluster
                     ↓
                   Kafka
                     ↓
              Order Consumer
                     ↓
                   PostgreSQL
```

再加入：

```text
Prometheus
Grafana
OpenTelemetry
Jaeger/Tempo
```

---

# 56. 秒杀请求最终流程

## 活动开始前

```text
PostgreSQL
 ↓
Preload
 ↓
Redis
```

预热：

```text
活动详情
库存
用户资格
```

---

## 活动开始

```text
Client
 ↓
CDN / Static Page
 ↓
Nginx
 ↓
Gateway Rate Limit
 ↓
Seckill Service
 ↓
Redis Lua
```

Lua：

```text
检查活动状态
检查库存
检查重复购买
扣 Redis 库存
记录购买标记
```

成功后：

```text
Kafka
```

快速返回：

```text
排队中
```

---

## 异步订单

```text
Kafka
 ↓
Order Consumer
 ↓
PostgreSQL Transaction
```

创建：

```text
Order
+
Outbox/Event
```

---

## 最终一致性

后台：

```text
reconciliation worker
```

周期对比：

```text
Redis Stock
PostgreSQL Orders
Kafka Events
```

发现异常：

```text
repair
+
alert
```

---

# 57. 防止系统被打爆的多层保护

```text
CDN
↓
Nginx limit
↓
Gateway limit
↓
User-level limit
↓
Redis atomic validation
↓
Kafka queue
↓
DB
```

理念：

```text
让无效请求尽可能早失败
```

---

# 58. Phase 7 ------ Kubernetes / Cloud Native

这一阶段不是面试初期核心，放后面。

将服务：

```text
docker build
```

部署到 Kubernetes。

学习：

```text
Pod
Deployment
Service
ConfigMap
Secret
Ingress
HPA
Readiness Probe
Liveness Probe
Rolling Update
```

---

# 59. HPA 实验

压测：

```text
1000 QPS
↓
10000 QPS
```

观察：

```text
replica 2
↓
replica 5
↓
replica 10
```

理解：

```text
horizontal scaling
```

---

# 60. Graceful Shutdown

Kubernetes 删除 Pod：

```text
SIGTERM
```

Go 应用需要：

```text
停止接收新请求
等待正在执行的请求
关闭 consumer
关闭 DB connection
退出
```

这是非常好的 Go 工程面试点。

---

# 61. Phase 8 ------ 故障演练

最终项目必须主动"搞坏"。

## Chaos 1

Kill Redis。

观察：

```text
缓存失效
请求是否打爆 DB
```

---

## Chaos 2

Kill Kafka。

观察：

```text
秒杀请求怎么办
消息是否丢失
```

---

## Chaos 3

Kill PostgreSQL。

观察：

```text
consumer backlog
```

---

## Chaos 4

Inventory Service latency：

```text
+ 3 秒
```

观察：

```text
timeout
retry
circuit breaker
```

---

## Chaos 5

随机：

```text
5% packet loss
```

思考：

```text
distributed system ≠ reliable network
```

---

# 62. 项目测试体系

目录：

```text
tests/
├── unit/
├── integration/
├── e2e/
├── load/
└── chaos/
```

---

## Unit Test

测试：

```text
Service business logic
```

---

## Integration Test

测试：

```text
PostgreSQL
Redis
Kafka
```

使用真实容器环境。

---

## E2E

完整：

```text
Create Activity
↓
Start Activity
↓
Seckill
↓
Kafka
↓
Create Order
↓
Query Order
```

---

## Load Test

k6：

```text
100
1000
10000
100000 concurrent requests
```

实际值根据本机能力调整。

重要的是：

```text
比较优化前后
```

不是强行跑到某个数字。

---

# 63. 性能报告必须记录

每次优化前后记录：

```text
QPS
P50
P95
P99
Error Rate
CPU
Memory
DB QPS
DB Connection
Redis QPS
Kafka Lag
```

示例：

```text
Before Redis

QPS: 1,200
P95: 85ms
PostgreSQL QPS: 1,150

After Redis

QPS: 8,700
P95: 12ms
PostgreSQL QPS: 80
Redis Hit Rate: 96%
```

注意：

简历最终只能写你真实测出来的数据。

---

# 64. 每个 Phase 必须写一份技术报告

目录：

```text
docs/
├── architecture.md
├── database.md
├── database/
│   └── mysql-vs-postgresql.md
├── phase1-monolith.md
├── phase2-concurrency.md
├── phase3-cache.md
├── phase3-kafka.md
├── phase4-distributed.md
├── phase5-microservices.md
├── performance.md
└── postmortems/
```

---

# 65. Postmortem

每次制造故障后写：

```text
Incident
Impact
Root Cause
Detection
Resolution
Prevention
Trade-off
```

例如：

```text
Redis 热点 Key 过期导致 DB CPU 100%
```

这会极大提升面试项目表达能力。

---

# 66. 面试官可能沿着项目追问什么

## Go

```text
goroutine 和 thread 区别？
GMP？
channel？
context？
mutex？
RWMutex？
atomic？
GC？
逃逸分析？
```

## HTTP

```text
HTTP/1.1 vs HTTP/2？
Keep Alive？
TCP？
HTTPS？
```

## PostgreSQL

```text
为什么出现超卖？
事务怎么解决？
MVCC？
索引为什么 B+Tree？
联合索引？
锁？
死锁？
```

## Redis

```text
为什么 Redis 快？
缓存一致性？
缓存穿透？
击穿？
雪崩？
Lua 为什么原子？
分布式锁？
```

## Kafka

```text
为什么 Kafka 吞吐高？
Partition？
Consumer Group？
消息丢失？
重复消费？
顺序？
积压？
```

## 分布式

```text
CAP？
一致性？
幂等？
分布式锁？
分布式事务？
Snowflake？
```

## 微服务

```text
为什么拆？
服务发现？
RPC？
超时？
重试？
熔断？
链路追踪？
```

## 系统设计

```text
10 万 QPS 怎么办？
Redis 挂了怎么办？
Kafka 挂了怎么办？
MySQL 挂了怎么办？
热点活动怎么办？
怎么避免超卖？
怎么保证一人一单？
```

---

# 67. 每一个知识点必须遵守"五件套"

学习 Redis 时不能只写：

```text
Redis 是内存数据库
```

必须形成：

```text
1 个业务场景
+
1 份实现代码
+
1 次压测
+
1 个故障实验
+
1 份 trade-off 总结
```

例如：

```text
Redis
→ 活动缓存
→ Cache Aside Code
→ k6 Test
→ Cache Breakdown Experiment
→ Cache Consistency Trade-off
```

Kafka：

```text
Kafka
→ 秒杀请求削峰
→ Producer / Consumer
→ Throughput Test
→ Consumer Crash
→ At-least-once + Idempotency
```

分布式锁：

```text
Distributed Lock
→ 多节点重复任务
→ Redis SET NX PX
→ Concurrent Test
→ Lock Expiration
→ lease / fencing trade-off
```

---

# 68. 项目最终简历表达目标

最终不要写：

> 使用 Gin、Redis、Kafka、MySQL、gRPC 开发高并发秒杀系统。

这种描述没有价值。

应该根据真实实验数据写成：

> 基于 Go/Gin 构建高并发秒杀服务，针对热点活动查询引入 Redis Cache Aside
> 缓存，并通过随机 TTL 与 singleflight 缓解缓存雪崩和热点 Key
> 击穿问题；通过 k6 压测验证优化前后的 QPS、P95 延迟及数据库请求量变化。

第二条：

> 针对秒杀库存超卖问题，对比实现 MySQL 悲观锁、条件原子 UPDATE
> 与乐观锁方案，最终结合 Redis Lua
> 实现库存校验、重复购买校验及库存扣减的原子执行，并使用数据库唯一约束作为最终一致性保护。

第三条：

> 使用 Kafka 将秒杀请求与订单持久化解耦，通过 Consumer Group
> 横向扩展订单消费能力，并使用业务唯一键实现幂等消费；针对 Consumer
> Crash、消息重复和消息积压场景进行了故障演练与监控。

第四条：

> 将单体系统逐步拆分为活动、库存、订单等服务，使用 gRPC
> 进行内部通信，并实现超时、指数退避重试、熔断及 OpenTelemetry
> 分布式链路追踪，分析网络调用对系统可靠性的影响。

注意：

这些只能在对应功能真实完成、真实测试之后写进简历。

---

# 68.1 数据库学习策略：项目 PostgreSQL，面试补 MySQL

本项目采用：

```text
主线实践：PostgreSQL
        +
面试补充：MySQL / InnoDB
```

学习一个数据库概念时尽量做一次横向对比，例如：

主题 PostgreSQL 主线 MySQL 面试补充

---

MVCC Tuple Version + xmin/xmax + Snapshot Undo Log + Read View
日志 WAL Redo Log + Undo Log
默认隔离级别 Read Committed Repeatable Read
幻读/串行化 Serializable SSI 等机制 Next-Key Lock 等机制
索引 B-Tree、GIN、GiST、BRIN 等 InnoDB B+Tree、聚簇/二级索引
数据回收 VACUUM / Autovacuum InnoDB Purge
执行计划 EXPLAIN ANALYZE EXPLAIN / EXPLAIN ANALYZE
悲观并发控制 SELECT FOR UPDATE SELECT FOR UPDATE

最终目标不是争论哪个数据库"更好"，而是能够回答：

> 为什么当前项目选择 PostgreSQL？它在这个项目中怎么工作？如果换成
> MySQL，事务、MVCC、锁和索引实现上会发生哪些重要变化？

---

# 69. 推荐开发顺序

不要并行学习所有东西。

严格执行：

```text
Phase 0
Go 工程结构
↓
Phase 1
Gin + PostgreSQL 单体
↓
Phase 2
事务 + 索引 + 并发 + 数据库锁
↓
Phase 3A
Redis 缓存
↓
Phase 3B
Redis Lua + 秒杀
↓
Phase 3C
限流 + Nginx
↓
Phase 3D
Kafka
↓
Phase 4
多实例 + RPC + 分布式问题
↓
Phase 5
微服务治理
↓
Phase 6
系统设计
↓
Phase 7
Kubernetes
↓
Phase 8
故障演练
```

---

# 70. 第一阶段不要做什么

Phase 1 禁止：

```text
Redis
Kafka
Kubernetes
微服务
etcd
gRPC
Elasticsearch
复杂 DDD
CQRS
Event Sourcing
Service Mesh
```

原因：

如果最基本的：

```text
Handler
Service
Repository
Transaction
SQL
Index
```

还不能独立实现，那么增加基础设施只会隐藏基础薄弱的问题。

---

# 71. 第二阶段不要做什么

不要看到并发问题直接：

```text
上 Redis
```

必须先经历：

```text
错误实现
↓
复现超卖
↓
数据库事务
↓
悲观锁
↓
条件 UPDATE
↓
乐观锁
↓
性能比较
↓
Redis
```

因为面试真正有价值的是：

```text
“我为什么从方案 A 演进到方案 B”
```

而不是：

```text
“我的项目用了 Redis”
```

---

# 72. 最终能力目标

完成本项目后，面对一个后端需求，你应该能够自动想到：

```text
这个业务的数据模型是什么？
↓
API 怎么设计？
↓
事务边界在哪里？
↓
哪些字段需要索引？
↓
会不会出现并发竞争？
↓
是否需要缓存？
↓
缓存一致性怎么办？
↓
是否适合异步？
↓
消息重复怎么办？
↓
服务是否需要拆分？
↓
网络失败怎么办？
↓
如何限流？
↓
如何监控？
↓
如何压测？
↓
如何处理故障？
↓
如何解释 trade-off？
```

这才是本项目真正需要训练的：

# 后端系统思维

而不是：

# 框架 API 记忆

---

# 73. 第一阶段立即执行的任务

项目第一次 Commit 只做这些：

```text
feat: initialize flashsale monolith
```

内容：

1.  创建 Go Module；
2.  初始化 Gin；
3.  Docker Compose 启动 PostgreSQL；
4.  数据库连接；
5.  Migration；
6.  User 表；
7.  Item 表；
8.  Activity 表；
9.  Inventory 表；
10. Order 表；
11. Health Check；
12. 统一 Response；
13. 统一 Error；
14. Zap Logger；
15. Graceful Shutdown。

然后开始：

```text
POST /activities
GET /activities/:id
POST /activities/:id/seckill
GET /orders/:id
```

第一版秒杀代码：

```text
只使用 PostgreSQL
```

不要提前优化。

等第一版真正跑起来之后：

# 第一次任务不是加 Redis。

而是：

# 用 k6 把它打出超卖问题。
