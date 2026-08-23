\set ON_ERROR_STOP on
\timing on

DROP TABLE IF EXISTS phase2_index_types_lab;
CREATE TABLE phase2_index_types_lab (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    email TEXT NOT NULL,
    status TEXT NOT NULL,
    metadata JSONB NOT NULL,
    sale_window TSTZRANGE NOT NULL,
    created_at TIMESTAMPTZ NOT NULL
);

INSERT INTO phase2_index_types_lab(user_id,email,status,metadata,sale_window,created_at)
SELECT n,
       'Buyer' || n || '@Example.Test',
       CASE WHEN n % 20 = 0 THEN 'cancelled' ELSE 'pending' END,
       jsonb_build_object('city',CASE WHEN n%3=0 THEN 'Paris' ELSE 'Lyon' END,'tags',jsonb_build_array('ticket',(n%10)::text)),
       tstzrange(CURRENT_TIMESTAMP+(n||' seconds')::interval,CURRENT_TIMESTAMP+((n+3600)||' seconds')::interval,'[)'),
       TIMESTAMPTZ '2026-01-01 00:00:00+00' + (n || ' seconds')::interval
FROM generate_series(1,100000) n;

CREATE INDEX phase2_btree_user_idx ON phase2_index_types_lab USING btree(user_id);
CREATE INDEX phase2_gin_metadata_idx ON phase2_index_types_lab USING gin(metadata);
CREATE INDEX phase2_gist_window_idx ON phase2_index_types_lab USING gist(sale_window);
CREATE INDEX phase2_brin_created_idx ON phase2_index_types_lab USING brin(created_at);
CREATE INDEX phase2_partial_pending_idx ON phase2_index_types_lab(created_at DESC) WHERE status='pending';
CREATE INDEX phase2_expression_email_idx ON phase2_index_types_lab(LOWER(email));
ANALYZE phase2_index_types_lab;

SELECT indexrelname,pg_size_pretty(pg_relation_size(indexrelid)) AS index_size
FROM pg_stat_user_indexes
WHERE relname='phase2_index_types_lab'
ORDER BY indexrelname;

EXPLAIN (ANALYZE,BUFFERS) SELECT * FROM phase2_index_types_lab WHERE user_id=4242;
EXPLAIN (ANALYZE,BUFFERS) SELECT COUNT(*) FROM phase2_index_types_lab WHERE metadata @> '{"city":"Paris"}';
EXPLAIN (ANALYZE,BUFFERS) SELECT COUNT(*) FROM phase2_index_types_lab WHERE sale_window @> CURRENT_TIMESTAMP+INTERVAL '2 hours';
EXPLAIN (ANALYZE,BUFFERS) SELECT COUNT(*) FROM phase2_index_types_lab WHERE created_at BETWEEN TIMESTAMPTZ '2026-01-02' AND TIMESTAMPTZ '2026-01-03';
EXPLAIN (ANALYZE,BUFFERS) SELECT * FROM phase2_index_types_lab WHERE status='pending' ORDER BY created_at DESC LIMIT 20;
EXPLAIN (ANALYZE,BUFFERS) SELECT * FROM phase2_index_types_lab WHERE LOWER(email)='buyer4242@example.test';

