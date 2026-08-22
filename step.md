# TicketGo 实现进度

## 当前 Phase

- 当前阶段：**Phase 0——工程准备与可启动骨架**。
- 当前状态：**已完成，全部 Phase 0 阶段门禁已通过**。
- 下一阶段：Phase 1——PostgreSQL 单体 MVP；尚未开始，本仓库当前不包含 Phase 1 业务实现。

## 运行项目所需指令

以下命令均在项目根目录 `D:\Code\TicketGo` 执行。

### 1. 首次运行准备

确认 Docker Desktop 已启动，然后执行：

```powershell
Set-Location D:\Code\TicketGo

# 仅在项目内安装锁定的 Go 版本；不会修改系统 PATH
make bootstrap-go

# 首次运行时创建本地环境配置
Copy-Item .env.example .env
```

如果本机 5432 已被其他 PostgreSQL 占用，需要在 `.env` 中设置：

```dotenv
POSTGRES_PORT=55432
DATABASE_URL=postgres://ticketgo:ticketgo_local_password@localhost:55432/ticketgo?sslmode=disable
```

当前开发环境的 `.env` 已采用 55432，无需重复修改。

### 2. 启动 PostgreSQL 并执行 migration

```powershell
make compose-up
make migrate-up
```

可用以下命令检查数据库容器状态：

```powershell
docker compose ps
```

预期 `ticketgo-postgres-1` 为 `healthy`。

### 3. 启动 TicketGo 服务

```powershell
make run
```

服务默认监听：

```text
http://localhost:8080
```

保持当前终端运行，在另一个 PowerShell 终端执行健康检查：

```powershell
curl.exe -i http://localhost:8080/health/live
curl.exe -i http://localhost:8080/health/ready
```

两个接口在 PostgreSQL 正常时都应返回 HTTP 200；ready 响应还应显示 PostgreSQL 状态为 `ok`。

### 4. 运行代码质量检查

```powershell
make check
```

该命令依次执行格式化、`go vet ./...`、`go test ./...` 和构建。

### 5. 停止项目

先在运行服务的终端按 `Ctrl+C` 停止 Go 服务，再执行：

```powershell
make compose-down
```

该命令停止并删除 Compose 容器与项目网络，但保留 PostgreSQL 具名数据卷。下次启动继续执行 `make compose-up` 和 `make run` 即可。

如需验证 migration 回滚流程：

```powershell
make migrate-down
make migrate-up
```

## 2026-08-22：Phase 0 工程准备与可启动骨架

### 执行依据与范围

- 严格按 `plan.md` 的 Phase 0 执行，并参考 `Idea.md` 的项目背景、技术主线、HTTP 请求路径和渐进式演进原则。
- 本阶段只建立模块化单体骨架，不实现用户、商品、活动、库存、订单等业务，不引入 Redis、Kafka、gRPC、etcd 或微服务。
- 数据库主线固定为 PostgreSQL；migration 工具选定为 golang-migrate，后续不与 Goose 混用。

### 环境处理过程

1. 盘点发现系统 PATH 中没有 Go；Docker CLI/Compose、GNU Make 和 Git 已存在，但 Docker Desktop daemon 未运行。
2. 查询 Go 官方发布信息，锁定 2026-08-19 发布的稳定版 Go 1.27.0。
3. 未进行全局安装：从 `go.dev` 下载 Windows amd64 ZIP 到项目 `.tools`，按官方 SHA-256 校验成功后解压为 `.tools/go`，没有修改系统 PATH。
4. Go module cache 与 build cache 均固定在项目 `.cache`；`.tools`、`.cache` 已加入 `.gitignore`。
5. 新增 `scripts/bootstrap-go.ps1` 和 `make bootstrap-go`，便于 Windows 新环境重复创建同样的项目内 Go 工具链。
6. Gin、pgx、zap 与 golang-migrate 依赖均下载到项目缓存并在 `go.mod`/`go.sum` 中锁定；未安装全局 Go CLI 工具。
7. 经用户确认后启动 Docker Desktop，并拉取 PostgreSQL、Go 和 Alpine 的锁定镜像用于真实门禁验证。因宿主机 5432 已被现有 PostgreSQL 占用，项目 `.env` 使用 `POSTGRES_PORT=55432`，未停止或修改现有实例。

### 已实现内容

- 初始化 Go module、`cmd/server/main.go`、`internal` 与 `pkg` 模块边界，并为 Phase 1 领域目录增加占位包说明。
- 实现严格环境变量配置加载与校验；错误配置快速失败。
- 使用 `database/sql + pgx` 建立 PostgreSQL 连接池，设置最大连接、空闲连接、连接寿命、空闲寿命、连接超时和查询超时；启动时 ping，但不自动建表。
- 使用 Gin 建立 Router，提供统一成功/错误 JSON、稳定错误码和 request ID。
- 实现 request ID、结构化访问日志、panic recovery、请求 context 超时中间件。
- 实现 `/health/live` 与 `/health/ready`：liveness 不访问依赖，readiness 带超时检查 PostgreSQL，失败返回 503 且不泄露内部数据库错误。
- HTTP Server 显式配置 read、read-header、write、idle timeout；捕获 SIGINT/SIGTERM 并在宽限期内优雅关闭 HTTP、数据库连接池和 logger。
- 新增 Docker Compose PostgreSQL 17.6（持久卷、UTC、健康检查、本地开发账号）和多阶段 Dockerfile（Go 1.27.0）。
- 新增 golang-migrate bootstrap migration。Phase 0 migration 只用于建立/验证 `schema_migrations` 流程，不提前创建业务表。
- 新增跨 Windows/Linux 的 Makefile 命令：bootstrap、依赖、格式化、vet、测试、构建、运行、Compose 和 migration。
- 新增 GitHub Actions 基础 CI，执行格式检查、vet、测试和构建。
- 新增 `README.md` 与 `docs/phase0-bootstrap.md`，记录从零启动、一次 HTTP 请求路径、配置、迁移、故障行为和验收步骤。

### 验证结果

| 检查项 | 结果 | 说明 |
| --- | --- | --- |
| `make check` | 通过 | 已依次完成 format、`go vet ./...`、`go test ./...`、build |
| 单元/路由测试 | 通过 | 覆盖配置校验、request ID、readiness 依赖失败以及 live/ready 语义分离 |
| `docker compose config` | 通过 | Compose 配置可解析；Docker 用户配置文件有权限警告，但不影响配置解析 |
| golang-migrate CLI | 通过 | 首次发现 `go run` 未编入 PostgreSQL driver；Makefile 增加 `-tags postgres` 后验证通过 |
| PostgreSQL 启动失败行为 | 通过 | 账号认证失败时，服务在连接超时内退出并输出结构化错误 |
| Compose PostgreSQL | 通过 | `postgres:17.6-alpine` 在宿主机 55432 启动并达到 healthy |
| 空库 up/down/up | 通过 | 三步均成功；最终 `schema_migrations` 为 `version=1, dirty=false` |
| live/ready 实际 HTTP | 通过 | 正常时 live=200、ready=200；统一 JSON 和请求 ID 正确 |
| PostgreSQL 故障与恢复 | 通过 | 容器停止时 live=200、ready=503；恢复后 ready=200 |
| SIGTERM 实际进程验证 | 通过 | Linux 服务容器收到 Docker SIGTERM 后约 436ms 干净退出，exit code=0、OOM=false |
| Docker image build | 通过 | 多阶段 Dockerfile 成功构建 `ticketgo:phase0` |

### 当前进度与后续动作

Phase 0 的实现与阶段门禁现已全部完成，可以进入 Phase 1，但本次没有提前实现任何 Phase 1 业务。项目 PostgreSQL 容器当前保持运行并处于 healthy，映射端口为 55432；具名数据卷保留 migration 状态。仅用于 SIGTERM 的临时 TicketGo 容器已删除，本地测试服务进程也已清理，8080 端口已释放。Docker 镜像 `ticketgo:phase0` 保留供后续启动验证使用。

验证过程中发现并修复了两项实际工程问题：

1. GNU Make 在 Windows 使用 `cmd.exe`，类 Unix 环境变量前缀不可用；现已按 Windows/Linux 分支处理，并为 Windows 构建产物增加 `.exe` 后缀。
2. golang-migrate 通过 `go run` 构建时需要显式启用 PostgreSQL build tag；现已固定 `-tags postgres`，避免 `unknown driver postgres`。
