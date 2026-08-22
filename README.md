# TicketGo

TicketGo 是按照 `plan.md` 分阶段演进的 Go 高并发抢票系统。当前已完成 Phase 1 PostgreSQL 单体 MVP：提供用户、商品、活动/库存、秒杀和订单闭环，仍故意保留朴素 read-modify-write 的并发缺陷，供 Phase 2 实验。当前不包含 Redis、Kafka、gRPC 或微服务。

## 环境基线

- Go 1.27.0（`go.mod`、Dockerfile、CI 锁定）
- PostgreSQL 17.6
- Docker Desktop / Docker Compose
- GNU Make（可选，命令也可直接执行）

项目会优先使用 `.tools/go` 下的便携 Go；`Makefile` 将 Go 模块与构建缓存写入项目内 `.cache`，不会修改全局 Go 配置。

Windows 上若没有 Go，可先执行以下命令。脚本只下载到当前项目的 `.tools/go`，并校验官方 SHA-256，不安装 MSI、不修改系统 PATH：

```powershell
make bootstrap-go
```

## 本地启动

```sh
cp .env.example .env
docker compose up -d
make migrate-up
make run
```

另开终端验证：

```sh
curl -i http://localhost:8080/health/live
curl -i http://localhost:8080/health/ready
```

Windows PowerShell 可用 `Copy-Item .env.example .env` 代替 `cp`。应用直接读取环境变量；`JWT_SECRET` 必须至少 32 字符。Make 会加载项目根目录 `.env`，生产环境必须显式注入独立的安全配置。

若本机 5432 已被其他 PostgreSQL 占用，可在 `.env` 中同时把 `POSTGRES_PORT` 改为例如 `55432`，并把 `DATABASE_URL` 的端口同步改为 `55432`。不要停止或改写现有数据库实例。

## 常用命令

```sh
make deps
make format
make lint
make test
make build
make run
make migrate-up
make migrate-down
make migrate-create NAME=add_example
make check
```

完整启动链路见 `docs/phase0-bootstrap.md`；Phase 1 的数据库、API、实现与故障记录分别见 `docs/database.md`、`docs/api.md`、`docs/phase1-monolith.md` 和 `docs/postmortems/`。
