# Phase 0：工程骨架与启动链路

## 范围与技术决策

Phase 0 只解决工程启动问题。HTTP 使用 Gin，日志使用 zap，数据库访问使用 `database/sql + pgx`，迁移统一使用 golang-migrate。GORM 留到 Phase 1 的 CRUD 场景再引入；业务表、认证、秒杀、缓存和消息队列均不属于本阶段。

选用 Go 1.27.0 与 PostgreSQL 17.6 并锁定在工程文件中。开发数据库账号只用于本地 Compose，不能用于生产环境。

## 启动链路

1. `docker compose up -d` 启动 PostgreSQL，具名卷保存数据，健康检查使用 `pg_isready`。
2. `make migrate-up` 通过 golang-migrate 应用 `migrations/`。首次执行会创建 `schema_migrations`；Phase 0 的 bootstrap migration 不创建业务表。
3. `make run` 加载并验证环境变量，创建有界数据库连接池，在启动 HTTP 服务前执行带超时的 ping。
4. 配置非法或 PostgreSQL 无法连接时，进程快速失败并输出可操作的结构化错误日志；应用不会自动修改 schema。

## 一次 HTTP 请求的路径

```text
TCP -> net/http -> Gin Router
    -> Request ID -> Access Log -> Panic Recovery -> Request Timeout
    -> Health Handler -> PostgreSQL Ping（仅 readiness）
    -> 统一 JSON Response
```

`/health/live` 只证明进程与 HTTP 栈存活，不访问外部依赖。`/health/ready` 使用独立查询超时 ping PostgreSQL；失败时返回 503、稳定错误码和 request ID，内部数据库错误只写日志而不暴露给客户端。

## 配置与连接池

所有配置来自环境变量并带严格类型校验，示例见 `.env.example`。连接池显式设置最大打开连接数、最大空闲连接数、连接最大寿命和最大空闲时间。所有数据库调用必须接收 `context.Context` 并受调用级超时约束。

HTTP 服务显式设置 read、read-header、write、idle 和请求处理超时。收到 SIGINT/SIGTERM 后停止接收新请求，在 `HTTP_SHUTDOWN_TIMEOUT` 宽限期内等待在途请求，然后关闭连接池并刷新 logger。

## Migration 验证

在空库执行以下序列：

```sh
make migrate-up
make migrate-down
make migrate-up
```

期望三步均成功。任何正式表结构变更都必须新增成对的 up/down migration，禁止依赖应用启动时自动建表。

## Phase 0 验收

```sh
make check
docker compose config
docker compose up -d
make migrate-up
make run
curl -i http://localhost:8080/health/live
curl -i http://localhost:8080/health/ready
```

停止 PostgreSQL 后，已运行服务的 liveness 应保持 200，readiness 应返回 503。向服务发送 SIGTERM 时，应在宽限期内记录干净退出日志。

