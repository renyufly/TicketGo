# PostgreSQL 17.6 内部机制实验

## MVCC 与隔离级别

`phase2lab internals` 用两个真实连接完成相同实验：事务 A 先读到 10，事务 B 更新为 20 并提交，事务 A 再读。

| 隔离级别 | 第一次读 | B 提交后第二次读 | 解释 |
| --- | ---: | ---: | --- |
| READ COMMITTED | 10 | 20 | 每条语句取得新的已提交快照 |
| REPEATABLE READ | 10 | 10 | 整个事务保持第一次查询建立的快照 |

普通 SELECT 通过 tuple visibility 读符合快照的版本，不等同于 `FOR UPDATE`。后者除读取可见版本外还获取行级锁，用于协调后续写入。

## Dead Tuple、VACUUM 与 Autovacuum

实验表插入 100,000 行，更新全部行并删除其中 50,000 行。`pg_stat_user_tables.n_dead_tup` 在 VACUUM 前约为 150,000，`VACUUM (ANALYZE)` 后为 0。

PostgreSQL UPDATE 通常创建新 tuple version，旧版本要等不再被任何快照需要后才能回收。VACUUM 回收可复用空间并维护 visibility map，但普通 VACUUM 通常不把文件立即缩回操作系统。长事务会长期保留旧快照，阻碍回收并放大 bloat；Autovacuum 阈值应结合热点表更新量监控。

## WAL、数据页与 Checkpoint

实验建表、插入 50,000 行并更新约三分之一，WAL LSN 从 `0/42687908` 推进到 `0/43941F00`，生成 19,637,752 bytes WAL；手动 CHECKPOINT 后 checkpoint LSN 为 `0/43941F58`。

事务提交首先保证对应 WAL 记录达到持久化要求。脏数据页可由 background writer/checkpointer 随后批量写回；崩溃恢复从 checkpoint 起重放 WAL，把已提交但尚未落到数据文件的更改恢复出来。因此数据库无需在每次提交时随机刷写所有修改页。代价是 WAL 写入、checkpoint I/O 峰值、恢复时间和复制延迟都必须监控。

## 百万订单 EXPLAIN ANALYZE

无 `(user_id, created_at DESC, id DESC)` 时，Planner 选择 3 路 Parallel Seq Scan、各过滤约 333,300 行，再 top-N sort；执行 43.832ms，buffers hit=8,842/read=3,746。

建立 39MiB 联合 B-Tree 后选择 Index Scan，只访问 26 个 buffer，执行 0.239ms，约快 183 倍。代价也经实测确认：表总大小 149MiB → 188MiB；另一个 10 万行插入对照中，写耗时 170.956ms → 353.787ms，空间约 9.7MiB → 15MiB。

业务 `orders` 在 Phase 1 已有同列索引，因此 Phase 2 使用隔离实验表完成 before/after，避免为了演示而破坏主业务查询。原始计划见 `docs/results/phase2/order-index-explain.txt`。

## 索引类型实验

| 类型 | 实验查询 | 实际计划/规模 | 适用边界 |
| --- | --- | --- | --- |
| B-Tree | `user_id = 4242` | Index Scan，0.133ms，2.2MiB | 等值、排序、范围；订单主查询首选 |
| GIN | JSONB `@>` city | Bitmap Index/Heap Scan，14.378ms，744KiB | JSONB、数组、全文的包含关系；写放大较高 |
| GiST | `tstzrange @> timestamp` | Bitmap Index/Heap Scan，0.950ms，7.7MiB | 范围、几何、近邻等可扩展搜索 |
| BRIN | 顺序时间范围 | lossy Bitmap Scan，2.577ms，24KiB | 超大、物理顺序与值高度相关的表；体积极小但需 recheck |
| Partial | pending 状态倒序 | Index Scan + LIMIT，0.067ms，2.0MiB | 只索引高价值子集，谓词必须与查询匹配 |
| Expression | `LOWER(email)` | Index Scan，0.067ms，3.9MiB | 规范化表达式查询；表达式必须一致 |

Planner 是否选索引取决于选择性、统计信息、相关性、成本参数、缓存与返回行数，不应只用“建了索引就必须走”判断。原始输出见 `docs/results/phase2/index-types-explain.txt`。
