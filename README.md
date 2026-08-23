# TicketGo

TicketGo 是按照 `plan.md` 分阶段演进的 Go 高并发抢票系统。当前已完成 Phase 2：稳定复现 Phase 1 的 lost update 超卖，使用相同 k6 场景比较 PostgreSQL 悲观锁、条件原子 UPDATE 与 version CAS，并以实测选择条件原子 UPDATE 作为默认库存扣减方案。项目仍是 PostgreSQL 模块化单体，尚未引入 Redis、Kafka、gRPC、Next.js、BFF 或微服务。

## 环境基线

- Go 1.27.0（项目 `.tools/go`、`go.mod`、CI 和根 Dockerfile 锁定）
- PostgreSQL 17.6
- k6 v2.2.0（项目 `.tools/k6`，SHA-256 校验）
- Node.js 24.11.1、npm 11.6.2（`web/.node-version`、`package.json`、CI 和前端 Dockerfile 锁定）
- React 19.2.8、Vite 8.2.2、TypeScript 5.9.3
- Docker Desktop / Docker Compose；GNU Make 可选

Go/k6 工具链、Go/npm 缓存和前端依赖都优先放在项目目录内：`.tools/go`、`.tools/k6`、`.cache` 和 `web/node_modules`。项目不会修改全局 Go/Node/k6 版本。若 Windows 没有 Go 或 k6，可运行 `make bootstrap-go` / `make bootstrap-k6`；脚本只下载便携版到本项目。

## 本地启动

```powershell
Set-Location D:\Code\TicketGo
if (-not (Test-Path .env)) { Copy-Item .env.example .env }
make bootstrap-go
make compose-up
make migrate-up
make run
```

另开终端安装并启动前端：

```powershell
Set-Location D:\Code\TicketGo
make web-install
make web-dev
```

浏览器打开 `http://localhost:5173`。Vite 会把 `/api` 和 `/health` 代理到 `http://localhost:8080`，本地开发无需扩大后端 CORS。

## 本地管理员演示开关

默认禁止匿名注册 admin。仅为本地演示时，在根 `.env` 设置：

```dotenv
ALLOW_ADMIN_SELF_REGISTRATION=true
```

并创建 `web/.env.local`：

```dotenv
VITE_ALLOW_ADMIN_REGISTRATION=true
```

前端变量只决定是否显示 admin 选项；Gin 后端仍执行最终权限校验。生产环境禁止开启这两个开关，未来真实部署应改为受控初始化、邀请或运维流程。

## 质量门禁

```powershell
# Go 格式、vet、单测/build，加上前端 format/lint/typecheck/test/build
make check

# 连接真实 PostgreSQL 的 Go integration/E2E
$env:TEST_DATABASE_URL = "postgres://ticketgo:ticketgo_local_password@localhost:55432/ticketgo?sslmode=disable"
make test
Remove-Item Env:TEST_DATABASE_URL

# Gin 已以 admin 演示开关启动时，验证真实 API 完整演示闭环
make web-e2e
```

## Phase 2 并发与数据库实验

默认主链路配置如下；`naive` 只允许在受控本地实验中使用：

```dotenv
SECKILL_INVENTORY_STRATEGY=atomic
SECKILL_LOCK_TIMEOUT=500ms
SECKILL_STATEMENT_TIMEOUT=2500ms
SECKILL_OPTIMISTIC_MAX_RETRIES=5
SECKILL_OPTIMISTIC_BACKOFF=2ms
```

确认 Docker PostgreSQL healthy 后，可重复运行固定的 1000 用户抢 100 库存实验，以及百万订单/PostgreSQL 内部机制实验：

```powershell
make bootstrap-k6
make phase2-reproduce
```

正式三轮均值：悲观锁 1083.77 QPS、条件原子 UPDATE 1561.46 QPS、乐观锁 1047.48 QPS；三种方案均严格成功 100 单且不超卖。完整参数、逐轮数据、索引代价与选型依据见 `docs/phase2-concurrency-report.md`、`docs/adr/0001-inventory-concurrency-control.md` 和 `docs/database/postgresql-internals.md`。

前端静态镜像可用 `docker build -f web/Dockerfile -t ticketgo-web:phase1b web` 构建。前端说明见 `docs/phase1b-frontend.md`；Phase 2 报告与 PostgreSQL 实验见上述文档；早期实现见 `docs/phase0-bootstrap.md`、`docs/database.md`、`docs/api.md`、`docs/phase1-monolith.md`。
