# Phase 1B：React 前端演示页面

## 1. 目标与边界

本阶段在已通过门禁的 Phase 1 Gin/PostgreSQL 单体上增加轻量浏览器演示层，覆盖注册、登录、当前用户、商品、活动/库存、秒杀、订单、取消以及 live/ready。前端只改善可操作性和反馈，不拥有库存、价格、权限、一人一单或订单状态机的最终判断权。

继续明确延期：不加入 Redis、Kafka、Next.js、SSR、BFF、支付、OAuth、复杂 RBAC 或 Phase 2 并发修复。同步秒杀成功后直接返回订单，不模拟异步排队语义。

## 2. 技术与项目内环境

- Node.js 24.11.1、npm 11.6.2；版本锁定在 `web/.node-version`、`web/package.json`、CI 和 `web/Dockerfile`。
- React 19.2.8、React Router 7.18.2、TanStack Query 5.102.0、React Hook Form 7.86.0、Zod 4.4.3。
- Vite 8.2.2、TypeScript 5.9.3、Vitest 4.1.11、React Testing Library、ESLint 10 和 Prettier 3。
- npm 依赖写入 `web/node_modules`，缓存写入项目 `.cache/npm`；没有执行全局 npm 安装，也没有修改系统 PATH。
- `web/go.mod` 只作为 Go module 边界，防止根模块的 `go test ./...` 误扫描 `node_modules` 中第三方示例 Go 包，不参与前端运行。

## 3. 结构与数据流

`src/api` 定义 TypeScript 契约、统一 envelope/API error 和金额/时间格式化；`src/auth` 管理单一 JWT、身份恢复与退出；`src/features` 对应 users/items/activities/orders API；`src/pages` 提供页面；`src/router` 执行登录与 admin 页面访问控制。

API Client 自动附加 Bearer Token，成功时解包 `data/request_id`，失败时保留 `code/message/request_id`。401 会清除 session Token；403、503、库存不足、重复参与和活动不可用等稳定错误会转换为用户可读提示，追踪编号可以复制。数据库错误或堆栈不会显示在页面。

TanStack Query 只缓存服务端数据。秒杀成功后失效 activity/order 查询；取消成功后失效订单和活动库存查询。没有本地乐观扣库存。金额始终保持整数 cents；时间表单按浏览器本地时区输入，提交前转为 RFC3339/UTC。

## 4. 认证与 admin 演示开关

JWT 只保存在当前标签页的 `sessionStorage`，不会进入 URL、日志或错误提示；浏览器只用 `localStorage` 记忆最近邮箱，不保存密码或多个 Token。退出会清除当前 JWT，再通过正常登录完成账号切换。

后端 `POST /api/v1/users` 的 `role` 只允许 `customer/admin`，缺省为 customer。`ALLOW_ADMIN_SELF_REGISTRATION=false` 是安全默认值；请求 admin 时返回稳定 403。只有后端开关为 true 才能真正创建 admin。`VITE_ALLOW_ADMIN_REGISTRATION=true` 只显示 admin 单选项，不能越过 Gin 与数据库权限校验。

生产环境必须保持匿名 admin 注册关闭，并删除或替换为受控初始化/邀请流程。

## 5. 启动与演示

后端按 README 启动后，另开终端：

```powershell
Set-Location D:\Code\TicketGo
make web-install
make web-dev
```

本地 admin 演示需同时配置根 `.env` 的 `ALLOW_ADMIN_SELF_REGISTRATION=true` 与 `web/.env.local` 的 `VITE_ALLOW_ADMIN_REGISTRATION=true`，然后重启 Gin 和 Vite。打开 `http://localhost:5173`，依次执行：

```text
注册并登录 admin
→ 创建商品
→ 创建 active 活动和库存
→ 退出并清除当前 JWT
→ 注册并登录 customer
→ 浏览活动详情并秒杀
→ 查看订单
→ 取消订单
→ 返回活动详情确认 available 回补、sold 减少
→ 系统状态页确认 live 与 PostgreSQL ready
```

## 6. 测试与部署

`npm run check` 依次执行 format check、ESLint、strict typecheck、Vitest 和 production build。组件/API 测试覆盖标准 envelope、稳定错误/request_id 显示。`npm run test:e2e` 连接真实 Gin/PostgreSQL，验证 health、admin/customer 注册登录、创建商品/活动、秒杀、查单、取消和库存回补；运行前后端必须启用 admin 本地开关。

真实浏览器验收已完成同一页面闭环，并确认普通用户导航不显示管理入口、订单取消后活动库存从 `1/2` 回补到 `2/2`、系统状态页显示 live/ready/PostgreSQL `ok`，浏览器控制台无 error/warn。

生产构建是纯静态文件。`web/Dockerfile` 使用锁定 Node 构建，再由 Nginx 提供 SPA fallback，并把 `/api`、`/health` 转发到名为 `ticketgo` 的 Gin 服务；部署平台需要提供该 DNS/网络关系。

## 7. 已知取舍

- 页面按钮状态和库存展示可能短暂滞后，提交结果以后端为准。
- Phase 1 朴素 read-modify-write 的 lost update 风险仍存在，前端的请求中禁用只防止误点，不宣称解决超卖。
- BIGSERIAL ID 在当前小规模演示中使用 number，类型层不对 ID 做算术；超过 JavaScript 安全整数范围后需统一改字符串。
- 当前为学习项目演示级 Token 存储；正式公网应用需结合威胁模型评估 HttpOnly Cookie、CSRF 与更完整的会话策略。
