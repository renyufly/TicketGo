# MySQL/InnoDB 与 PostgreSQL 并发机制对比

本项目运行数据库固定为 PostgreSQL；以下只用于面试与设计比较，不建立 MySQL 业务链路。

| 主题 | PostgreSQL 17 | MySQL 8 / InnoDB | TicketGo 结论 |
| --- | --- | --- | --- |
| 默认隔离级别 | READ COMMITTED | REPEATABLE READ | 必须显式说明，不把另一数据库默认行为套用过来 |
| MVCC 版本位置 | 新 tuple version 留在 heap，VACUUM 回收 | 聚簇索引记录 + undo log 构造旧版本 | 两者都不是简单原地覆盖，但清理与膨胀机制不同 |
| 锁定读 | `SELECT ... FOR UPDATE` | `SELECT ... FOR UPDATE` | 都能串行化热点行，语法相似不代表锁范围完全相同 |
| 范围锁/幻读 | PostgreSQL 普通行锁不使用 InnoDB next-key lock 模型；SERIALIZABLE 使用 SSI | RR 下 next-key/gap lock 常用于范围修改与幻读控制 | 分析死锁时必须依据实际引擎与查询计划 |
| 条件扣减 | `UPDATE ... WHERE available >= n RETURNING ...` | `UPDATE ... WHERE available >= n` 后检查 affected rows | 两者都适合把检查与修改合成原子写；PostgreSQL 可直接 RETURNING |
| 乐观锁 | version 条件 UPDATE + `RETURNING` | version 条件 UPDATE + affected rows | 热点下都可能重试放大，必须限制次数和退避 |
| 执行计划 | `EXPLAIN (ANALYZE, BUFFERS)` | `EXPLAIN ANALYZE` | 字段名称不同；本项目报告使用 PostgreSQL 的 actual time/loops/buffers |
| 索引组织 | heap + 独立 B-Tree，二级索引指向 tuple | 主键聚簇，二级索引叶子包含主键 | 主键宽度对 InnoDB 二级索引代价尤其敏感 |
| 后台清理 | VACUUM/Autovacuum 回收 dead tuple | purge thread 清理 undo/旧版本 | 长事务都会拖慢旧版本回收，但观测指标与运维手段不同 |
| WAL/redo | WAL + checkpoint + crash recovery | redo log + checkpoint + crash recovery | 都用顺序日志避免每次提交同步刷完全部数据页 |

Phase 2 选条件原子 UPDATE 的核心依据是数据库原子谓词更新与实测，而不是某个数据库品牌特有技巧。迁移到 MySQL 时需要重新验证隔离级别、affected rows 语义、死锁码、索引组织与执行计划，不能只替换驱动。

