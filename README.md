# TicketGo

TicketGo 是按照 `plan.md` 分阶段演进的 Go 高并发抢票系统。当前已完成 Phase 1B：在 Phase 1 PostgreSQL 单体 MVP 上增加 React 演示页面，可通过浏览器完成管理员建商品/活动、账号切换、普通用户秒杀、查单和取消回补库存的完整闭环。后端仍故意保留朴素 read-modify-write 并发缺陷，供 Phase 2 实验；当前未引入 Redis、Kafka、gRPC、Next.js、BFF 或微服务。

## 环境基线

- Go 1.27.0（项目 `.tools/go`、`go.mod`、CI 和根 Dockerfile 锁定）
- PostgreSQL 17.6
- Node.js 24.11.1、npm 11.6.2（`web/.node-version`、`package.json`、CI 和前端 Dockerfile 锁定）
- React 19.2.8、Vite 8.2.2、TypeScript 5.9.3
- Docker Desktop / Docker Compose；GNU Make 可选

Go 工具链、Go/npm 缓存和前端依赖都优先放在项目目录内：`.tools/go`、`.cache` 和 `web/node_modules`。项目不会修改全局 Go/Node 版本。若 Windows 没有 Go，可运行 `make bootstrap-go`；脚本只下载便携版到本项目。

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

前端静态镜像可用 `docker build -f web/Dockerfile -t ticketgo-web:phase1b web` 构建。详细设计、启动、安全边界与演示流程见 `docs/phase1b-frontend.md`；Phase 0/1 的数据库、API 与后端实现说明见 `docs/phase0-bootstrap.md`、`docs/database.md`、`docs/api.md`、`docs/phase1-monolith.md`。
