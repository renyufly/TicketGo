# TicketGo 实现进度

## 当前 Phase

- 当前阶段：**Phase 1B——React 前端演示页面**。
- 当前状态：**已完成，Phase 1B 阶段门禁已通过**。
- 下一阶段：Phase 2——单机并发、数据库锁与 PostgreSQL 实验；尚未开始，Phase 1 的朴素并发缺陷仍保留。

## 运行项目所需指令

以下命令均在项目根目录 `D:\Code\TicketGo` 执行。

### 1. 首次运行与项目内环境准备

确认 Docker Desktop 已启动，然后执行：

```powershell
Set-Location D:\Code\TicketGo

# 仅在项目内安装锁定的 Go 1.27.0；不会安装 MSI 或修改系统 PATH
make bootstrap-go

# 只在 .env 不存在时创建，避免覆盖已有本地配置
if (-not (Test-Path .env)) {
    Copy-Item .env.example .env
}
```

打开 `.env`，确认 `JWT_SECRET` 已替换为至少 32 字符的本地随机值。不要把真实密钥提交到 Git：

```dotenv
JWT_SECRET=replace-with-your-own-random-secret-at-least-32-characters
AUTH_TOKEN_TTL=24h
ALLOW_ADMIN_SELF_REGISTRATION=false
```

如果本机 5432 已被其他 PostgreSQL 占用，在 `.env` 中同步修改映射端口与连接地址：

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

可确认 Phase 1 migration 已应用到 version 2：

```powershell
docker compose exec postgres psql -U ticketgo -d ticketgo -c "SELECT version, dirty FROM schema_migrations;"
```

预期结果为 `version=2`、`dirty=false`。

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

### 4. 启动 Phase 1B React 演示页面

另开终端执行；依赖与缓存只写入项目目录：

```powershell
Set-Location D:\Code\TicketGo
make web-install
make web-dev
```

打开 `http://localhost:5173`。Vite 会把 `/api` 和 `/health` 代理到 Gin 8080。

仅需演示匿名 admin 注册时，在根 `.env` 设置 `ALLOW_ADMIN_SELF_REGISTRATION=true`，并创建 `web/.env.local`：

```dotenv
VITE_ALLOW_ADMIN_REGISTRATION=true
```

两个开关必须同时开启并重启前后端。前端开关只控制显示，后端开关才决定是否允许创建 admin；生产环境禁止开启。随后可完全通过页面执行 admin 建商品/活动 → 退出切换 customer → 秒杀 → 查单 → 取消 → 查看库存回补 → 查看系统状态的闭环。

### 5. 跑通 Phase 1 API 命令行闭环（可选）

以下指令在另一个 PowerShell 终端执行。先注册管理员账号，再通过仅限本地开发的数据库命令提升角色；提升后必须重新登录，因为旧 JWT 中的角色不会作为最终授权依据，但重新登录便于验证完整流程。

```powershell
Set-Location D:\Code\TicketGo
$base = "http://localhost:8080/api/v1"

# 4.1 注册管理员账号
$adminEmail = "admin@ticketgo.local"
$adminPassword = "TicketGo-Admin-123"
$adminRegisterBody = @{
    email = $adminEmail
    password = $adminPassword
} | ConvertTo-Json
Invoke-RestMethod -Method Post -Uri "$base/users" -ContentType "application/json" -Body $adminRegisterBody

# 4.2 本地开发环境提升管理员角色；生产环境必须使用受控运维流程
docker compose exec postgres psql -U ticketgo -d ticketgo -c "UPDATE users SET role='admin', updated_at=CURRENT_TIMESTAMP WHERE email='admin@ticketgo.local';"

# 4.3 登录并准备 Authorization Header
$adminLogin = Invoke-RestMethod -Method Post -Uri "$base/login" -ContentType "application/json" -Body $adminRegisterBody
$adminHeaders = @{ Authorization = "Bearer $($adminLogin.data.access_token)" }

# 4.4 创建商品
$itemBody = @{
    name = "TicketGo Phase 1 Concert"
    description = "Phase 1 serial business loop"
    price_cents = 10000
    status = "active"
} | ConvertTo-Json
$item = Invoke-RestMethod -Method Post -Uri "$base/items" -Headers $adminHeaders -ContentType "application/json" -Body $itemBody

# 4.5 创建正在进行的活动与库存
$activityBody = @{
    item_id = [int64]$item.data.id
    name = "TicketGo Phase 1 Sale"
    price_cents = 8000
    starts_at = (Get-Date).ToUniversalTime().AddMinutes(-1).ToString("o")
    ends_at = (Get-Date).ToUniversalTime().AddHours(1).ToString("o")
    status = "active"
    total = 10
} | ConvertTo-Json
$activity = Invoke-RestMethod -Method Post -Uri "$base/activities" -Headers $adminHeaders -ContentType "application/json" -Body $activityBody

# 4.6 注册并登录普通购票用户
$buyerBody = @{
    email = "buyer@ticketgo.local"
    password = "TicketGo-Buyer-123"
} | ConvertTo-Json
Invoke-RestMethod -Method Post -Uri "$base/users" -ContentType "application/json" -Body $buyerBody
$buyerLogin = Invoke-RestMethod -Method Post -Uri "$base/login" -ContentType "application/json" -Body $buyerBody
$buyerHeaders = @{ Authorization = "Bearer $($buyerLogin.data.access_token)" }

# 4.7 秒杀、查询订单并取消；取消会回补库存
$order = Invoke-RestMethod -Method Post -Uri "$base/activities/$($activity.data.id)/seckill" -Headers $buyerHeaders -ContentType "application/json" -Body '{"quantity":1}'
Invoke-RestMethod -Method Get -Uri "$base/orders?limit=20&offset=0" -Headers $buyerHeaders
Invoke-RestMethod -Method Get -Uri "$base/orders/$($order.data.id)" -Headers $buyerHeaders
Invoke-RestMethod -Method Post -Uri "$base/orders/$($order.data.id)/cancel" -Headers $buyerHeaders
```

若重复执行上述示例，注册接口会因邮箱唯一约束返回 409。可更换邮箱，或只在确认不需要这些本地演示数据时手动清理。

### 6. 运行质量、集成与 E2E 测试

```powershell
# 格式化、vet、普通测试和构建
make check

# 显式启用真实 PostgreSQL 集成测试与 HTTP E2E；地址必须与 .env 一致
$env:TEST_DATABASE_URL = "postgres://ticketgo:ticketgo_local_password@localhost:55432/ticketgo?sslmode=disable"
make test
Remove-Item Env:TEST_DATABASE_URL

# Gin 以 admin 本地演示开关启动时，运行真实前后端 API 演示闭环
make web-e2e
```

未设置 `TEST_DATABASE_URL` 时，涉及真实数据库的测试会安全跳过；Phase 1 门禁验证必须显式设置该变量并确认测试全部通过。

### 7. 停止项目

先在运行服务的终端按 `Ctrl+C` 停止 Go 服务，再执行：

```powershell
make compose-down
```

该命令停止并删除 Compose 容器与项目网络，但保留 PostgreSQL 具名数据卷。下次启动继续执行 `make compose-up` 和 `make run` 即可。

`make migrate-down` 会删除 Phase 1 的全部业务表和其中数据，不应在当前开发数据库上随意执行。需要验证 migration 回滚时，应使用一次性空数据库；Phase 1 已通过临时 PostgreSQL 完成过安全的 up → down → up 验证。

仅当明确确认当前数据库数据可以删除时，才执行：

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

| 检查项                  | 结果 | 说明                                                                                  |
| ----------------------- | ---- | ------------------------------------------------------------------------------------- |
| `make check`            | 通过 | 已依次完成 format、`go vet ./...`、`go test ./...`、build                             |
| 单元/路由测试           | 通过 | 覆盖配置校验、request ID、readiness 依赖失败以及 live/ready 语义分离                  |
| `docker compose config` | 通过 | Compose 配置可解析；Docker 用户配置文件有权限警告，但不影响配置解析                   |
| golang-migrate CLI      | 通过 | 首次发现 `go run` 未编入 PostgreSQL driver；Makefile 增加 `-tags postgres` 后验证通过 |
| PostgreSQL 启动失败行为 | 通过 | 账号认证失败时，服务在连接超时内退出并输出结构化错误                                  |
| Compose PostgreSQL      | 通过 | `postgres:17.6-alpine` 在宿主机 55432 启动并达到 healthy                              |
| 空库 up/down/up         | 通过 | 三步均成功；最终 `schema_migrations` 为 `version=1, dirty=false`                      |
| live/ready 实际 HTTP    | 通过 | 正常时 live=200、ready=200；统一 JSON 和请求 ID 正确                                  |
| PostgreSQL 故障与恢复   | 通过 | 容器停止时 live=200、ready=503；恢复后 ready=200                                      |
| SIGTERM 实际进程验证    | 通过 | Linux 服务容器收到 Docker SIGTERM 后约 436ms 干净退出，exit code=0、OOM=false         |
| Docker image build      | 通过 | 多阶段 Dockerfile 成功构建 `ticketgo:phase0`                                          |

### 当前进度与后续动作

Phase 0 的实现与阶段门禁现已全部完成，可以进入 Phase 1，但本次没有提前实现任何 Phase 1 业务。项目 PostgreSQL 容器当前保持运行并处于 healthy，映射端口为 55432；具名数据卷保留 migration 状态。仅用于 SIGTERM 的临时 TicketGo 容器已删除，本地测试服务进程也已清理，8080 端口已释放。Docker 镜像 `ticketgo:phase0` 保留供后续启动验证使用。

验证过程中发现并修复了两项实际工程问题：

1. GNU Make 在 Windows 使用 `cmd.exe`，类 Unix 环境变量前缀不可用；现已按 Windows/Linux 分支处理，并为 Windows 构建产物增加 `.exe` 后缀。
2. golang-migrate 通过 `go run` 构建时需要显式启用 PostgreSQL build tag；现已固定 `-tags postgres`，避免 `unknown driver postgres`。

## 2026-08-22：Phase 1 PostgreSQL 单体 MVP

### 执行依据与范围

- 严格执行 `plan.md` Phase 1，并以 `Idea.md` 的抢票业务背景、渐进演进和“五件套”原则为参考。
- 只使用 Gin、PostgreSQL、bcrypt、JWT 和现有日志/配置骨架；没有引入 Redis、Kafka、gRPC、etcd、微服务、进程内锁或其他 Phase 2+ 组件。
- 秒杀实现刻意采用 PostgreSQL read-modify-write 朴素方案，保留并发 lost update 窗口，未提前掩盖 Phase 2 要复现的问题。

### 环境与依赖处理过程

1. 继续使用 Phase 0 已放在项目 `.tools/go` 的 Go 1.27.0，以及项目 `.cache/go-mod`、`.cache/go-build`，未修改系统 PATH 或全局 Go 环境。
2. 继续使用 TicketGo Compose 中锁定的 PostgreSQL 17.6 Alpine，宿主机端口为 55432；没有安装或修改全局 PostgreSQL。
3. 密码哈希直接使用依赖树中已有并已锁定的 `golang.org/x/crypto/bcrypt v0.37.0`，`go mod tidy` 只将其从 indirect 调整为 direct，没有新增全局工具。
4. migration CLI 缓存缺少版本元数据时，经授权仅补齐项目 `.cache`，随后成功应用 `000002_phase1_monolith`。
5. `.env.example` 新增 `JWT_SECRET` 与 `AUTH_TOKEN_TTL`；密钥至少 32 字符，真实密钥不进入 Git。

### 实现过程

1. 新增 users、items、activities、inventories、orders、seckill_records 六张业务表及成对 migration；加入 PK、FK、NOT NULL、CHECK、UNIQUE 和查询索引。
2. 固化金额最小单位整数、UTC/RFC3339 时间、Activity/Order 状态、库存等式、一人一活动唯一记录，以及“取消回补库存但不可重新参加”的规则。
3. 各业务模块按 model/dto → repository → service → handler 分层；Repository 接收 context，Service 持有业务规则与事务边界，Handler 不执行核心 SQL。
4. 实现注册、登录、当前用户、商品创建/列表/详情、活动与库存创建/列表/详情、秒杀、订单列表/详情/取消 API。
5. 密码使用 bcrypt 强哈希；JWT 使用 HS256、过期时间和最短密钥校验；管理写接口与普通用户接口通过中间件区分权限。
6. 秒杀事务依次执行库存读取与校验、绝对值库存回写、订单创建、秒杀记录创建、commit；任一步失败均 rollback。
7. 取消事务锁定订单，只允许 pending 取消，并在同一事务内回补库存、更新订单与秒杀记录。
8. 列表固定倒序、limit 最大 100、offset 最大 10000；订单查询始终按当前用户隔离。
9. 新增 `docs/database.md`、`docs/api.md`、`docs/phase1-monolith.md` 和两份 Phase 1 postmortem，并更新 README。

### 测试与故障实验

| 检查项               | 结果 | 说明                                                                                               |
| -------------------- | ---- | -------------------------------------------------------------------------------------------------- |
| 单元测试             | 通过 | 覆盖 JWT 过期、配置密钥、商品价格、活动时间、非法数量、重复用户、未开始/已结束/非 active、库存不足 |
| PostgreSQL 集成测试  | 通过 | 覆盖提交、取消回补、唯一约束、FK/CHECK 防线与事务回滚                                              |
| 库存后注入错误       | 通过 | `total=2, available=2, sold=0`，订单数 0                                                           |
| 订单 CHECK 约束失败  | 通过 | quantity=0 被数据库拒绝，此前库存更新完整回滚                                                      |
| 重复用户             | 通过 | 返回 409，唯一秒杀记录阻止重复，库存修改回滚                                                       |
| HTTP E2E             | 通过 | 完整注册到取消闭环，覆盖 400/401/403/404/409/500 映射                                              |
| migration            | 通过 | 主项目应用到 version=2；另用无持久卷临时 PostgreSQL 完成空库 up → down → up，随后删除临时容器      |
| PostgreSQL 停机/恢复 | 通过 | 停机前 ready=200；停机中 live=200、ready=503；恢复 healthy 后 ready=200，无需重启应用              |
| `make format`        | 通过 | Go 源码格式化完成                                                                                  |
| `go vet ./...`       | 通过 | 无 vet 问题                                                                                        |
| `go test ./...`      | 通过 | 设置 TEST_DATABASE_URL 后真实单元、集成、E2E 全部通过                                              |
| `go build`           | 通过 | 生成 `bin/ticketgo.exe`                                                                            |

### 当前进度与 Phase 2 边界

Phase 1 的串行业务闭环、事务原子性、错误映射、request_id 日志关联、真实数据库集成测试与故障实验均已完成。PostgreSQL 容器已恢复 healthy，8080 上的临时验收服务已停止。

当前实现不能声称高并发下不超卖：两个事务可能读取同一库存旧值并绝对值回写，造成 lost update。下一步应严格进入 Phase 2，先用固定 k6 场景稳定复现，再比较 `SELECT FOR UPDATE`、条件原子 UPDATE 和 version CAS；在完成实测决策前不加入 Redis。

## 2026-08-22：Phase 1B React 前端演示页面

### 执行依据与范围

- 严格执行 `plan.md` 的 Phase 1B，并参考 `Idea.md` 的抢票领域模型、初始 API、单体事务边界和渐进演进原则。
- 本阶段只增加 React 演示能力与受控 admin 本地注册开关，不改变库存、一人一单、权限和订单状态机的后端最终裁决地位。
- 未加入 Redis、Kafka、Next.js、SSR、BFF、微服务或 Phase 2 并发修复；Phase 1 的朴素 read-modify-write 缺陷继续保留。

### 环境与依赖处理过程

1. 盘点发现系统已有 Node.js 24.11.1 和 npm 11.6.2，因此没有安装或升级任何全局 Node/npm/Go 工具。
2. 前端依赖安装到 `web/node_modules`，npm 缓存固定为项目 `.cache/npm`；Node/npm 版本锁定到 `web/.node-version`、`package.json`、CI 和 `web/Dockerfile`。
3. 查询并锁定 React 19.2.8、Vite 8.2.2、TypeScript 5.9.3 等精确依赖；初次选择 TypeScript 7.0.2 时发现与 typescript-eslint peer range 不兼容，改为受支持的 5.9.3，没有使用 `--force` 或忽略依赖约束。
4. jsdom 30 要求 Node 24.15.0，而项目锁定 Node 24.11.1；因此选择兼容当前 Node 的 jsdom 29.1.1，消除 engine 警告。
5. 新增 `web/go.mod` 作为 module 边界，避免根 Go 命令误扫描 `node_modules` 内第三方示例 Go 包。

### 运行项目所需指令

以下命令均在 Windows PowerShell 中执行，不会修改全局 Go、Node 或 npm 版本。

#### 1. 准备项目内环境与配置

```powershell
Set-Location D:\Code\TicketGo

# 仅当项目内便携 Go 不存在时执行；只写入 .tools/go 和 .cache
if (-not (Test-Path .tools\go\bin\go.exe)) {
    make bootstrap-go
}

# 不覆盖已经存在的本地环境配置
if (-not (Test-Path .env)) {
    Copy-Item .env.example .env
}

# 使用 package-lock.json 安装到 web/node_modules，npm 缓存写入 .cache/npm
make web-install
```

打开根目录 `.env`，至少确认 JWT 密钥和数据库端口正确。当前本机 PostgreSQL 映射使用 55432：

```dotenv
JWT_SECRET=replace-with-your-own-random-secret-at-least-32-characters
POSTGRES_PORT=55432
DATABASE_URL=postgres://ticketgo:ticketgo_local_password@localhost:55432/ticketgo?sslmode=disable
ALLOW_ADMIN_SELF_REGISTRATION=false
```

若需要在页面直接注册 admin 进行完整本地演示，把根 `.env` 的后端开关改为：

```dotenv
ALLOW_ADMIN_SELF_REGISTRATION=true
```

并创建 `web/.env.local`；若文件已经存在则不要覆盖，直接确认其中的值：

```powershell
if (-not (Test-Path web\.env.local)) {
    Set-Content -Path web\.env.local -Value 'VITE_ALLOW_ADMIN_REGISTRATION=true'
}
```

这两个 admin 开关只允许在本地演示环境同时开启；生产环境必须保持关闭。前端开关只控制选项显示，Gin 后端仍执行最终权限校验。

#### 2. 启动 PostgreSQL 并应用 migration

```powershell
Set-Location D:\Code\TicketGo
make compose-up
make migrate-up
docker compose ps
```

预期 `ticketgo-postgres-1` 状态为 `healthy`，migration 版本为 2。

#### 3. 启动 Gin 后端

在第一个 PowerShell 终端执行并保持运行：

```powershell
Set-Location D:\Code\TicketGo
make run
```

后端默认地址为 `http://localhost:8080`。可在另一个终端检查：

```powershell
curl.exe -i http://localhost:8080/health/live
curl.exe -i http://localhost:8080/health/ready
```

#### 4. 启动 React 前端

在第二个 PowerShell 终端执行并保持运行：

```powershell
Set-Location D:\Code\TicketGo
make web-dev
```

浏览器打开 `http://localhost:5173`。Vite 会把 `/api` 和 `/health` 代理到 Gin 8080，无需为本地双端口演示开启后端 CORS。

页面完整演示顺序：

```text
注册并登录 admin
→ 创建商品
→ 创建 active 活动与库存
→ 退出并切换 customer
→ 浏览活动并秒杀
→ 查看订单
→ 取消订单
→ 返回活动详情确认库存回补
→ 打开系统状态页确认 live/ready/PostgreSQL 均为 ok
```

#### 5. 运行质量与真实 E2E 门禁

```powershell
Set-Location D:\Code\TicketGo

# Go 与 React 的格式、lint、单测、类型检查和生产构建
make check

# 连接真实 PostgreSQL 的 Go integration/E2E
$env:TEST_DATABASE_URL = "postgres://ticketgo:ticketgo_local_password@localhost:55432/ticketgo?sslmode=disable"
make test
Remove-Item Env:TEST_DATABASE_URL

# Gin 已启动且允许本地 admin 注册时，验证真实 API 演示闭环
make web-e2e
```

#### 6. 停止项目

分别在 Gin 和 Vite 终端按 `Ctrl+C`，确认 8080 与 5173 已释放。随后停止 Compose 服务：

```powershell
Set-Location D:\Code\TicketGo
make compose-down
```

`make compose-down` 会保留 PostgreSQL 具名数据卷。不要为了停止服务执行 `make migrate-down`；该命令会删除业务表及其中数据。

### 实现过程

1. 建立 React + Vite + TypeScript strict 工程，接入 React Router、TanStack Query、React Hook Form、Zod、Vitest、Testing Library、ESLint 和 Prettier。
2. 统一 API Client 自动附加 Bearer Token，解包成功 envelope，保留稳定错误码和 request_id；网络、401、403、503、售罄、冲突和活动不可用均有用户可读反馈。
3. 完成注册、登录、身份恢复、当前用户、退出/安全账号切换；Token 仅保存在 sessionStorage，只记忆最近邮箱，不保存密码或多个 JWT。
4. 完成商品列表/详情/创建、活动列表/详情/创建与库存、秒杀、订单分页/详情/取消，以及 live/ready 系统状态页面，覆盖 Phase 1 的全部 API。
5. 金额全程使用整数 cents，仅展示时格式化；本地 datetime 输入转换为 RFC3339/UTC；秒杀与取消后主动失效订单和活动查询，不做本地乐观库存。
6. 后端注册 DTO 增加 customer/admin 枚举和缺省 customer；新增默认 false 的 `ALLOW_ADMIN_SELF_REGISTRATION`，关闭时 admin 请求稳定返回 403，前端显示开关不能替代后端校验。
7. 增加前端静态 Nginx 镜像、项目内 Make 命令、GitHub Actions 前端门禁、真实 Gin/PostgreSQL 演示脚本，并更新 README 和阶段文档。

### 验证结果

| 检查项                       | 结果 | 说明                                                                               |
| ---------------------------- | ---- | ---------------------------------------------------------------------------------- |
| 前端 format/ESLint/typecheck | 通过 | Prettier、ESLint、TypeScript strict 均无问题                                       |
| Vitest                       | 通过 | 2 个测试文件、3 个测试，覆盖 envelope、稳定错误与 request_id 展示                  |
| production build             | 通过 | Vite 转换 177 个模块，生成纯静态 dist                                              |
| Go vet/build                 | 通过 | 后端原有质量门禁继续通过                                                           |
| Go unit/integration/E2E      | 通过 | 连接真实 PostgreSQL，覆盖普通注册、admin 开关开/关、非法 role、事务与完整 API 闭环 |
| 真实 API 演示脚本            | 通过 | admin/customer 注册登录、商品、活动、秒杀、查单、取消和库存回补完整通过            |
| 真实浏览器页面闭环           | 通过 | admin 创建商品/活动，切换 customer 秒杀与取消；库存从 1/2 回补到 2/2               |
| 系统状态页面                 | 通过 | live=ok、ready=ok、PostgreSQL=ok，5 秒自动刷新                                     |
| 浏览器运行时日志             | 通过 | 控制台无 error/warn，普通用户导航不显示管理入口                                    |
| 前端 Docker build            | 通过 | 成功构建 `ticketgo-web:phase1b`，Node 与 Nginx 镜像精确锁定                        |

### 当前进度与下一阶段边界

Phase 1B 的源码、lockfile、项目内环境、页面闭环、后端安全开关、测试、CI、Docker 和文档均已完成，阶段门禁通过。验收期间创建了少量 `example.test` 本地演示账号、商品、活动和已取消订单；它们只存在于项目 PostgreSQL 数据卷，不包含真实个人信息。

本轮没有修改全局 Go/Node/npm 版本。PostgreSQL 容器保持 healthy；最终交付时临时 Gin/Vite 验收服务已停止，避免占用 8080/5173。下一步仍应按计划进入 Phase 2，用 k6 稳定复现 lost update/超卖，再比较 PostgreSQL 锁与原子更新方案，不能把前端按钮禁用误认为并发正确性修复。

### 前端页面解释

商品/ -> 资源目录 ：商品 / 门票

先创建一个基础商品：
商品名称：XXX 2026 巴黎演唱会门票
描述：巴黎站内场门票
基础价格：¥1,280
状态：active

> 它只说明“卖的是什么”，此时还没有销售时间和库存，所以用户不能秒杀

活动/ -> 实时库存：秒杀活动

管理员再基于这个商品创建一个活动：
活动名称：巴黎站首轮限量抢票
关联商品：XXX 2026 巴黎演唱会门票
活动价格：¥980
开始时间：2026-08-23 20:00
结束时间：2026-08-23 20:10
总库存：100 张
可用库存：100 张
已售：0 张
状态：active

> 这个活动说明“什么时候卖、卖多少钱、有多少库存”

用户操作

活动开始后，普通用户进入“秒杀活动”详情页：
购买数量：1
→ 点击“立即秒杀”
→ 可用库存变为 99
→ 已售变为 1
→ 生成 ¥980 的 pending 订单 (订单已创建，但还没有完成支付)
（位于 订单/）

如果用户取消订单：
订单状态：pending → cancelled
可用库存：99 → 100
已售数量：1 → 0

同一个商品还可以创建多个活动：
XXX 2026 巴黎演唱会门票
├── 早鸟抢票：¥880，库存 50
├── 首轮抢票：¥980，库存 100
└── 最终补票：¥1,180，库存 20

> 因此，“商品 / 门票”相当于商品档案；“秒杀活动”才是用户实际参加和下单的销售场次
