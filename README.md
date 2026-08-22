# TicketGo

TicketGo 是按照 `plan.md` 分阶段演进的 Go 高并发抢票系统。当前实现严格停留在 Phase 0：提供可重复启动、可迁移、可观测、可优雅退出的模块化单体骨架，不包含业务 API、Redis、Kafka、gRPC 或微服务。

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

Windows PowerShell 可用 `Copy-Item .env.example .env` 代替 `cp`。应用直接读取环境变量；默认值与 `.env.example` 的本地开发值一致，因此 Phase 0 可不加载 `.env` 文件直接启动。生产环境必须显式注入安全配置。

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

完整启动链路、请求路径、配置和故障行为见 `docs/phase0-bootstrap.md`。
