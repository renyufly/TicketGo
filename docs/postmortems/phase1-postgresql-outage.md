# Phase 1 Postmortem：PostgreSQL 停机

## Incident / Impact

2026-08-22 在服务运行期间停止 `ticketgo-postgres-1`，随后立即恢复。数据库停机时业务数据库请求不可用，但进程与 HTTP 栈仍存活。

## Detection

- 停机前：live=200（约 1.5ms），ready=200（约 2.3ms）。
- 停机中：live=200（约 0.9ms），ready=503（约 18.7ms）。
- 恢复且容器 healthy 后：ready=200（约 28.9ms）。
- warning 日志包含 dependency=postgresql 和内部连接错误；客户端只收到稳定的 `dependency_unavailable`，不会看到数据库详情。访问日志携带 request_id。

## Root Cause / Resolution

人为停止唯一 PostgreSQL 容器导致连接拒绝；`docker compose start postgres` 后连接池自动建立新连接，应用无需重启即恢复 readiness。

## Prevention / Trade-off

Phase 1 通过独立 live/ready、查询超时和连接池恢复避免“假健康”和无限等待，但单数据库仍是单点。高可用、重试与降级属于后续阶段；此时不引入隐藏的重试，以免放大故障流量。
