\set ON_ERROR_STOP on
\timing on

DROP TABLE IF EXISTS phase2_write_no_index;
DROP TABLE IF EXISTS phase2_write_with_index;
DROP TABLE IF EXISTS phase2_order_index_lab;
CREATE TABLE phase2_order_index_lab (
    id BIGSERIAL PRIMARY KEY,
    order_no TEXT NOT NULL UNIQUE,
    user_id BIGINT NOT NULL,
    activity_id BIGINT NOT NULL,
    quantity BIGINT NOT NULL,
    total_price_cents BIGINT NOT NULL,
    status TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL
);

INSERT INTO phase2_order_index_lab(order_no,user_id,activity_id,quantity,total_price_cents,status,created_at)
SELECT
    'P2-' || LPAD(n::text, 12, '0'),
    1 + (n % 10000),
    1 + (n % 100),
    1,
    8000,
    CASE WHEN n % 10 = 0 THEN 'cancelled' ELSE 'pending' END,
    TIMESTAMPTZ '2026-01-01 00:00:00+00' + (n || ' seconds')::interval
FROM generate_series(1, 1000000) n;
ANALYZE phase2_order_index_lab;

SELECT 'before_index' AS stage,pg_size_pretty(pg_total_relation_size('phase2_order_index_lab')) AS total_size;
EXPLAIN (ANALYZE, BUFFERS, FORMAT TEXT)
SELECT * FROM phase2_order_index_lab
WHERE user_id = 4242
ORDER BY created_at DESC,id DESC
LIMIT 20;

CREATE INDEX phase2_order_user_created_idx
ON phase2_order_index_lab(user_id,created_at DESC,id DESC);
ANALYZE phase2_order_index_lab;

SELECT 'after_index' AS stage,
       pg_size_pretty(pg_relation_size('phase2_order_user_created_idx')) AS index_size,
       pg_size_pretty(pg_total_relation_size('phase2_order_index_lab')) AS total_size;
EXPLAIN (ANALYZE, BUFFERS, FORMAT TEXT)
SELECT * FROM phase2_order_index_lab
WHERE user_id = 4242
ORDER BY created_at DESC,id DESC
LIMIT 20;

DROP TABLE IF EXISTS phase2_write_no_index;
DROP TABLE IF EXISTS phase2_write_with_index;
CREATE TABLE phase2_write_no_index (LIKE phase2_order_index_lab INCLUDING DEFAULTS INCLUDING CONSTRAINTS);
CREATE TABLE phase2_write_with_index (LIKE phase2_order_index_lab INCLUDING DEFAULTS INCLUDING CONSTRAINTS);
CREATE INDEX phase2_write_user_created_idx ON phase2_write_with_index(user_id,created_at DESC,id DESC);

EXPLAIN (ANALYZE, BUFFERS, FORMAT TEXT)
INSERT INTO phase2_write_no_index(order_no,user_id,activity_id,quantity,total_price_cents,status,created_at)
SELECT 'NOIDX-' || n,1+(n%10000),1+(n%100),1,8000,'pending',CURRENT_TIMESTAMP+(n||' milliseconds')::interval
FROM generate_series(1,100000) n;

EXPLAIN (ANALYZE, BUFFERS, FORMAT TEXT)
INSERT INTO phase2_write_with_index(order_no,user_id,activity_id,quantity,total_price_cents,status,created_at)
SELECT 'IDX-' || n,1+(n%10000),1+(n%100),1,8000,'pending',CURRENT_TIMESTAMP+(n||' milliseconds')::interval
FROM generate_series(1,100000) n;

SELECT relname,pg_size_pretty(pg_total_relation_size(relid)) AS total_size
FROM pg_catalog.pg_statio_user_tables
WHERE relname IN ('phase2_write_no_index','phase2_write_with_index')
ORDER BY relname;
