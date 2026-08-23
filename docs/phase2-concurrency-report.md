# Phase 2 单机并发与 PostgreSQL 实验报告

## 1. 问题与目标

本阶段严格按 `plan.md` Phase 2 执行：先稳定复现 Phase 1 的 lost update，再用同一数据与负载比较悲观锁、条件原子 UPDATE 和 version CAS，最后以实测选出主链路方案。未加入 Redis、Kafka、进程内锁或 Phase 2B 订单生命周期。

核心不变量：库存非负、`available + sold = total`、订单数量/数量合计不超过总库存、同一用户同一活动最多一单、失败事务不留下部分结果。

## 2. 环境与固定参数

- 主机：AMD Ryzen 7 8845HS，8 核/16 线程，32 GiB 内存，Windows 11 Pro 64 位（10.0.26200）；
- Go 1.27.0（项目 `.tools/go`），k6 v2.2.0（项目 `.tools/k6`），PostgreSQL 17.6 Alpine；
- PostgreSQL 容器未设置 CPU/内存上限，shared memory 64 MiB；宿主端口 55432；
- 应用连接池：max open 25、max idle 10；HTTP request timeout 3s；statement timeout 2500ms；lock timeout 500ms；
- 热点场景：1000 个唯一用户、100 份库存、1000 次请求、200 VU、每种方案 3 轮；每轮先删除上轮 Phase 2 数据并完整重建；
- k6 阈值：响应必须为 201、稳定售罄 409 或受控繁忙 503；意外响应为 0；P95 < 3s、P99 < 5s。

最初尝试 1000 VU 瞬时建连时，Windows 本机 TCP accept/backlog 出现大量 `connection refused`，该轮属于压测入口容量干扰并已排除。正式对比保持 1000 个不同用户和 1000 次请求不变，固定 200 VU；所有方案完全一致。

## 3. 超卖复现与事务交错

实验策略在库存 SELECT 后加入 75ms 受控延迟，放大 Phase 1 已存在的竞争窗口。三轮均出现 1000 个成功订单，库存行最终仅显示 `sold=40/41`、`available=60/59`。CHECK 约束没有失效；失效的是“订单总量不超过 total”这一跨表不变量。

```text
Transaction A                   Transaction B
SELECT available=100,sold=0     SELECT available=100,sold=0
计算 available=99,sold=1       计算 available=99,sold=1
UPDATE 写入 99/1               等待同一行
COMMIT                          获得锁后仍写入旧计算值 99/1
INSERT order A                  INSERT order B
结果：库存只减少 1，但产生 2 个订单
```

多批请求重复这一过程，导致订单数量远大于库存记录的 sold。原始请求 ID 与 SQL 对应日志保存在本地忽略目录 `docs/results/phase2/raw/`，可由脚本重新生成。

## 4. 三种并发方案结果

下表为正式三轮均值；正确性数据每轮完全一致。Docker CPU 是另一次相同参数代表性采样的峰值，允许超过 100%（表示使用多个逻辑核），只用于本机方案间参考，不外推为生产容量。

| 方案 | 成功/售罄 | 正确性 | QPS | P50 | P95 | P99 | PostgreSQL CPU 峰值 | 平均峰值锁等待者 | 最长活动事务峰值均值 |
| --- | --- | --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 朴素基线 | 1000/0 | 失败，超卖 | 292.77 | 546.59 ms | 1302.73 ms | 1730.20 ms | 66.59% | 15.33 | 177.839 ms |
| `FOR UPDATE` | 100/900 | 通过 | 1083.77 | 124.69 ms | 459.00 ms | 570.24 ms | 130.77% | 24.00 | 354.664 ms |
| 条件原子 UPDATE | 100/900 | 通过 | **1561.46** | **66.66 ms** | **401.82 ms** | **457.16 ms** | 92.17% | 23.67 | 266.703 ms |
| version CAS | 100/900 | 通过 | 1047.48 | 68.18 ms | 714.50 ms | 775.02 ms | 78.30% | 24.00 | 194.015 ms |

所有三种修复方案的每一轮均满足：`available=0`、`sold=100`、订单=100、秒杀记录=100、重复用户=0、意外错误率=0。完整逐轮数字见 `docs/results/phase2/concurrency-summary.json`。

### 乐观锁低冲突与热点对比

无竞争对照使用 1 VU、100 用户、100 库存，3 轮均无 CAS 重试，P95 为 6.11–6.92ms。热点场景中，对所有 1000 请求统计的平均重试为 0.102–0.104；由于 900 个请求读取到售罄后无需 CAS，折算到 100 个成功请求约为每单 1.02–1.04 次额外尝试。说明乐观锁在无竞争时简单快速，但热点单行上重试放大了数据库工作与尾延迟。

### 锁等待失败行为

真实 PostgreSQL 集成测试先持有库存行锁，再以 50ms lock timeout 调用悲观锁秒杀。请求稳定转换为领域错误 `ErrBusy`/HTTP `concurrency_busy`，事务回滚，没有无限等待。PostgreSQL 错误 `55P03`、`57014` 和 `40P01` 都被映射为可控的 503。

## 5. 决策与 PostgreSQL 瓶颈

主链路选用条件原子 UPDATE，理由见 ADR-0001。它在保证正确性的三种方案中吞吐最高，P95/P99 最低，SQL 与失败判定也最直接。

本阶段已经证明热点瓶颈仍在 PostgreSQL：25 个连接的池在正式采样中达到 25 个活动会话，热点阶段有 23–24 个会话等待同一库存行，悲观锁三轮“最长活动事务峰值”的均值为 354.664ms；各方案代表轮次 CPU 峰值为 66.59%–130.77%，但吞吐受单行串行化而不是主机总 CPU 限制。增加应用并发不能消除这一行级竞争。Phase 3 才有理由把大量无效竞争提前挡在 Redis，同时 PostgreSQL 条件 UPDATE 与约束仍保留为最终防线。

## 6. 索引与 PostgreSQL 专项实验摘要

- 百万订单查询：联合索引前 Parallel Seq Scan + sort，执行 43.832ms、读取/命中 12,588 个 buffer；索引后 Index Scan，执行 0.239ms、26 个 buffer；
- 索引空间：表总大小从 149MiB 增至 188MiB，联合索引 39MiB；
- 10 万行写入：无联合索引 170.956ms，有联合索引 353.787ms，后者约 2.07 倍；总空间从 9.7MiB 增至 15MiB；
- READ COMMITTED 的同一事务第二次读取看到其他事务提交值 20；REPEATABLE READ 仍看到初始快照 10；
- 10 万行 UPDATE + 5 万行 DELETE 产生约 150,000 个 dead tuple，手动 VACUUM 后统计为 0；
- WAL 实验生成 19,637,752 bytes WAL，CHECKPOINT LSN 前移；提交依赖 WAL 持久化，不要求每次同步刷完所有数据页；
- B-Tree、GIN、GiST、BRIN、Partial、Expression 索引均在隔离表完成实际 EXPLAIN。

详见 `docs/database/postgresql-internals.md`、`docs/results/phase2/order-index-explain.txt` 与 `docs/results/phase2/index-types-explain.txt`。

## 7. 复现、故障恢复与下一阶段

```powershell
make bootstrap-k6
make phase2-load
make phase2-postgres-labs
```

`phase2lab prepare` 每轮只清理 `[phase2]` 实验活动与 `phase2-load-*` 测试用户；数据库专项脚本只重建 `phase2_*_lab` 表。若实验中断，重新运行同一命令即可从一致起点恢复。负载脚本停止临时服务但保留 PostgreSQL 容器和实验结果。

Phase 2 门禁通过，可以进入 Phase 2B；尚未实现支付、超时关单、Redis 或后续阶段组件。
