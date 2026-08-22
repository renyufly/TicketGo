# Phase 1 API

所有业务接口位于 `/api/v1`。错误响应固定包含 `code`、`message`、`request_id`；成功响应包含 `data` 和 `request_id`。除注册和登录外均需 `Authorization: Bearer <token>`。

| 方法与路径 | 权限 | 说明 |
| --- | --- | --- |
| `POST /users` | 公开 | 注册普通用户 |
| `POST /login` | 公开 | 获取 HS256 JWT |
| `GET /users/me` | 登录 | 当前用户 |
| `POST /items` | admin | 创建商品 |
| `GET /items`、`GET /items/:id` | 登录 | 商品列表/详情 |
| `POST /activities` | admin | 同事务创建活动与库存 |
| `GET /activities`、`GET /activities/:id` | 登录 | 活动列表/详情 |
| `POST /activities/:id/seckill` | 登录 | 秒杀并创建 pending 订单 |
| `GET /orders`、`GET /orders/:id` | 登录 | 仅查询自己的订单 |
| `POST /orders/:id/cancel` | 登录 | 取消自己的 pending 订单并回补库存 |

列表参数为 `limit`（默认 20，最大 100）与 `offset`（默认 0，最大 10000），排序固定为 `created_at DESC, id DESC`。

本阶段没有对外开放创建管理员的入口。开发环境可在注册后明确执行 `UPDATE users SET role='admin' WHERE email='...'`，再重新登录获取含 admin 角色的新 token；生产环境应由受控运维流程配置。

示例请求：

```powershell
$base = "http://localhost:8080/api/v1"
Invoke-RestMethod -Method Post -Uri "$base/users" -ContentType "application/json" -Body '{"email":"user@example.com","password":"password-123"}'
$login = Invoke-RestMethod -Method Post -Uri "$base/login" -ContentType "application/json" -Body '{"email":"user@example.com","password":"password-123"}'
$headers = @{ Authorization = "Bearer $($login.data.access_token)" }
Invoke-RestMethod -Method Get -Uri "$base/users/me" -Headers $headers
Invoke-RestMethod -Method Post -Uri "$base/activities/1/seckill" -Headers $headers -ContentType "application/json" -Body '{"quantity":1}'
Invoke-RestMethod -Method Get -Uri "$base/orders?limit=20&offset=0" -Headers $headers
```
