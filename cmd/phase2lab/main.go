package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"ticketgo/internal/auth"
)

type dataset struct {
	ActivityID int64    `json:"activity_id"`
	Total      int      `json:"total"`
	Tokens     []string `json:"tokens"`
}

type verification struct {
	ActivityID       int64 `json:"activity_id"`
	Total            int64 `json:"total"`
	Available        int64 `json:"available"`
	Sold             int64 `json:"sold"`
	Orders           int64 `json:"orders"`
	OrderedQuantity  int64 `json:"ordered_quantity"`
	SeckillRecords   int64 `json:"seckill_records"`
	DuplicateBuyers  int64 `json:"duplicate_buyers"`
	InventorySafe    bool  `json:"inventory_safe"`
	OrderLimitSafe   bool  `json:"order_limit_safe"`
	OneOrderPerBuyer bool  `json:"one_order_per_buyer"`
	Safe             bool  `json:"safe"`
	Oversold         bool  `json:"oversold"`
}

func main() {
	if len(os.Args) < 2 {
		fatal(errors.New("usage: phase2lab prepare|verify [flags]"))
	}
	var err error
	switch os.Args[1] {
	case "prepare":
		err = prepare(os.Args[2:])
	case "verify":
		err = verify(os.Args[2:])
	case "internals":
		err = internals(os.Args[2:])
	case "monitor":
		err = monitor(os.Args[2:])
	default:
		err = fmt.Errorf("unknown command %q", os.Args[1])
	}
	if err != nil {
		fatal(err)
	}
}

func monitor(args []string) error {
	fs := flag.NewFlagSet("monitor", flag.ContinueOnError)
	dsn := fs.String("database-url", env("DATABASE_URL", "postgres://ticketgo:ticketgo_local_password@localhost:5432/ticketgo?sslmode=disable"), "PostgreSQL DSN")
	output := fs.String("output", "tests/load/.generated/phase2-monitor.csv", "CSV output")
	stopFile := fs.String("stop-file", "tests/load/.generated/phase2-monitor.stop", "sentinel file")
	interval := fs.Duration("interval", 50*time.Millisecond, "sample interval")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *interval <= 0 {
		return errors.New("interval must be positive")
	}
	db, err := openDB(*dsn)
	if err != nil {
		return err
	}
	defer db.Close()
	file, err := os.Create(*output)
	if err != nil {
		return err
	}
	defer file.Close()
	if _, err := fmt.Fprintln(file, "timestamp_utc,lock_waiters,active_sessions,oldest_active_transaction_ms"); err != nil {
		return err
	}
	ticker := time.NewTicker(*interval)
	defer ticker.Stop()
	for {
		if _, err := os.Stat(*stopFile); err == nil {
			return nil
		} else if !errors.Is(err, os.ErrNotExist) {
			return err
		}
		ctx, cancel := context.WithTimeout(context.Background(), *interval)
		var lockWaiters, active int
		var oldestTransactionMS float64
		err := db.QueryRowContext(ctx, `SELECT COUNT(*) FILTER (WHERE wait_event_type='Lock'),COUNT(*) FILTER (WHERE state='active'),COALESCE(MAX(EXTRACT(EPOCH FROM (clock_timestamp()-xact_start))*1000) FILTER (WHERE state='active' AND xact_start IS NOT NULL),0) FROM pg_stat_activity WHERE datname=current_database() AND pid<>pg_backend_pid()`).Scan(&lockWaiters, &active, &oldestTransactionMS)
		cancel()
		if err == nil {
			if _, err := fmt.Fprintf(file, "%s,%d,%d,%.3f\n", time.Now().UTC().Format(time.RFC3339Nano), lockWaiters, active, oldestTransactionMS); err != nil {
				return err
			}
		}
		<-ticker.C
	}
}

type internalsReport struct {
	PostgreSQLVersion string          `json:"postgresql_version"`
	MVCC              []mvccResult    `json:"mvcc"`
	DeadTuples        deadTupleResult `json:"dead_tuples"`
	WAL               walResult       `json:"wal"`
}

type mvccResult struct {
	Isolation  string `json:"isolation"`
	FirstRead  int    `json:"first_read"`
	SecondRead int    `json:"second_read_after_other_commit"`
}

type deadTupleResult struct {
	BeforeVacuum int64 `json:"before_vacuum"`
	AfterVacuum  int64 `json:"after_vacuum"`
}

type walResult struct {
	BeforeLSN      string `json:"before_lsn"`
	AfterLSN       string `json:"after_lsn"`
	GeneratedBytes string `json:"generated_bytes"`
	CheckpointLSN  string `json:"checkpoint_lsn"`
}

func internals(args []string) error {
	fs := flag.NewFlagSet("internals", flag.ContinueOnError)
	dsn := fs.String("database-url", env("DATABASE_URL", "postgres://ticketgo:ticketgo_local_password@localhost:5432/ticketgo?sslmode=disable"), "PostgreSQL DSN")
	output := fs.String("output", "docs/results/phase2/postgresql-internals.json", "JSON output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	db, err := openDB(*dsn)
	if err != nil {
		return err
	}
	defer db.Close()
	report := internalsReport{}
	if err := db.QueryRow(`SHOW server_version`).Scan(&report.PostgreSQLVersion); err != nil {
		return err
	}
	for _, isolation := range []struct {
		name  string
		level sql.IsolationLevel
	}{{"READ COMMITTED", sql.LevelReadCommitted}, {"REPEATABLE READ", sql.LevelRepeatableRead}} {
		result, err := observeMVCC(db, isolation.name, isolation.level)
		if err != nil {
			return err
		}
		report.MVCC = append(report.MVCC, result)
	}
	if report.DeadTuples, err = observeDeadTuples(db); err != nil {
		return err
	}
	if report.WAL, err = observeWAL(db); err != nil {
		return err
	}
	encoded, _ := json.MarshalIndent(report, "", "  ")
	if err := os.MkdirAll(filepath.Dir(*output), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(*output, encoded, 0o644); err != nil {
		return err
	}
	fmt.Println(string(encoded))
	return nil
}

func observeMVCC(db *sql.DB, name string, level sql.IsolationLevel) (mvccResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if _, err := db.ExecContext(ctx, `DROP TABLE IF EXISTS phase2_mvcc_lab; CREATE TABLE phase2_mvcc_lab(id INT PRIMARY KEY,value INT NOT NULL); INSERT INTO phase2_mvcc_lab VALUES(1,10)`); err != nil {
		return mvccResult{}, err
	}
	tx, err := db.BeginTx(ctx, &sql.TxOptions{Isolation: level, ReadOnly: true})
	if err != nil {
		return mvccResult{}, err
	}
	defer tx.Rollback()
	result := mvccResult{Isolation: name}
	if err := tx.QueryRowContext(ctx, `SELECT value FROM phase2_mvcc_lab WHERE id=1`).Scan(&result.FirstRead); err != nil {
		return mvccResult{}, err
	}
	if _, err := db.ExecContext(ctx, `UPDATE phase2_mvcc_lab SET value=20 WHERE id=1`); err != nil {
		return mvccResult{}, err
	}
	if err := tx.QueryRowContext(ctx, `SELECT value FROM phase2_mvcc_lab WHERE id=1`).Scan(&result.SecondRead); err != nil {
		return mvccResult{}, err
	}
	return result, tx.Commit()
}

func observeDeadTuples(db *sql.DB) (deadTupleResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if _, err := db.ExecContext(ctx, `DROP TABLE IF EXISTS phase2_vacuum_lab; CREATE TABLE phase2_vacuum_lab(id BIGINT PRIMARY KEY,payload TEXT NOT NULL); INSERT INTO phase2_vacuum_lab SELECT n,repeat('x',100) FROM generate_series(1,100000) n; UPDATE phase2_vacuum_lab SET payload=repeat('y',100); DELETE FROM phase2_vacuum_lab WHERE id % 2 = 0; ANALYZE phase2_vacuum_lab`); err != nil {
		return deadTupleResult{}, err
	}
	_, _ = db.ExecContext(ctx, `SELECT pg_stat_force_next_flush()`)
	result := deadTupleResult{}
	if err := db.QueryRowContext(ctx, `SELECT n_dead_tup FROM pg_stat_user_tables WHERE relname='phase2_vacuum_lab'`).Scan(&result.BeforeVacuum); err != nil {
		return deadTupleResult{}, err
	}
	if _, err := db.ExecContext(ctx, `VACUUM (ANALYZE) phase2_vacuum_lab`); err != nil {
		return deadTupleResult{}, err
	}
	_, _ = db.ExecContext(ctx, `SELECT pg_stat_force_next_flush()`)
	if err := db.QueryRowContext(ctx, `SELECT n_dead_tup FROM pg_stat_user_tables WHERE relname='phase2_vacuum_lab'`).Scan(&result.AfterVacuum); err != nil {
		return deadTupleResult{}, err
	}
	return result, nil
}

func observeWAL(db *sql.DB) (walResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	result := walResult{}
	if err := db.QueryRowContext(ctx, `SELECT pg_current_wal_lsn()::text`).Scan(&result.BeforeLSN); err != nil {
		return walResult{}, err
	}
	if _, err := db.ExecContext(ctx, `DROP TABLE IF EXISTS phase2_wal_lab; CREATE TABLE phase2_wal_lab(id BIGINT PRIMARY KEY,payload TEXT NOT NULL); INSERT INTO phase2_wal_lab SELECT n,repeat(md5(n::text),4) FROM generate_series(1,50000) n; UPDATE phase2_wal_lab SET payload=payload || 'changed' WHERE id % 3 = 0`); err != nil {
		return walResult{}, err
	}
	if err := db.QueryRowContext(ctx, `SELECT pg_current_wal_lsn()::text,pg_wal_lsn_diff(pg_current_wal_lsn(),$1::pg_lsn)::text`, result.BeforeLSN).Scan(&result.AfterLSN, &result.GeneratedBytes); err != nil {
		return walResult{}, err
	}
	if _, err := db.ExecContext(ctx, `CHECKPOINT`); err != nil {
		return walResult{}, err
	}
	if err := db.QueryRowContext(ctx, `SELECT checkpoint_lsn::text FROM pg_control_checkpoint()`).Scan(&result.CheckpointLSN); err != nil {
		return walResult{}, err
	}
	return result, nil
}

func prepare(args []string) error {
	fs := flag.NewFlagSet("prepare", flag.ContinueOnError)
	dsn := fs.String("database-url", env("DATABASE_URL", "postgres://ticketgo:ticketgo_local_password@localhost:5432/ticketgo?sslmode=disable"), "PostgreSQL DSN")
	secret := fs.String("jwt-secret", os.Getenv("JWT_SECRET"), "JWT signing secret")
	users := fs.Int("users", 1000, "unique buyers")
	stock := fs.Int("stock", 100, "activity inventory")
	output := fs.String("output", "tests/load/.generated/phase2-users.json", "dataset output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *users <= 0 || *stock <= 0 || *stock > *users {
		return errors.New("users and stock must be positive, and stock must not exceed users")
	}
	if len(*secret) < 32 {
		return errors.New("jwt-secret must contain at least 32 characters")
	}

	db, err := openDB(*dsn)
	if err != nil {
		return err
	}
	defer db.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err := cleanup(tx, ctx); err != nil {
		return err
	}
	var itemID, activityID int64
	if err := tx.QueryRowContext(ctx, `INSERT INTO items(name,description,price_cents,status) VALUES('[phase2] concurrency load item','generated by phase2lab',10000,'active') RETURNING id`).Scan(&itemID); err != nil {
		return err
	}
	if err := tx.QueryRowContext(ctx, `INSERT INTO activities(item_id,name,price_cents,starts_at,ends_at,status) VALUES($1,'[phase2] 1000 buyers / 100 tickets',8000,CURRENT_TIMESTAMP-INTERVAL '1 minute',CURRENT_TIMESTAMP+INTERVAL '2 hours','active') RETURNING id`, itemID).Scan(&activityID); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO inventories(activity_id,total,available) VALUES($1,$2,$2)`, activityID, *stock); err != nil {
		return err
	}

	manager := auth.NewManager(*secret, 4*time.Hour)
	data := dataset{ActivityID: activityID, Total: *stock, Tokens: make([]string, 0, *users)}
	stamp := time.Now().UTC().Format("20060102T150405.000000000")
	for i := 0; i < *users; i++ {
		var userID int64
		email := fmt.Sprintf("phase2-load-%s-%04d@example.test", stamp, i)
		if err := tx.QueryRowContext(ctx, `INSERT INTO users(email,password_hash,role,status) VALUES($1,'phase2-no-login','customer','active') RETURNING id`, email).Scan(&userID); err != nil {
			return err
		}
		token, err := manager.Issue(userID, "customer")
		if err != nil {
			return err
		}
		data.Tokens = append(data.Tokens, token)
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(*output), 0o755); err != nil {
		return err
	}
	encoded, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(*output, encoded, 0o600); err != nil {
		return err
	}
	fmt.Printf("prepared activity=%d users=%d stock=%d dataset=%s\n", activityID, *users, *stock, *output)
	return nil
}

func verify(args []string) error {
	fs := flag.NewFlagSet("verify", flag.ContinueOnError)
	dsn := fs.String("database-url", env("DATABASE_URL", "postgres://ticketgo:ticketgo_local_password@localhost:5432/ticketgo?sslmode=disable"), "PostgreSQL DSN")
	datasetPath := fs.String("dataset", "tests/load/.generated/phase2-users.json", "prepared dataset")
	expect := fs.String("expect", "safe", "safe or oversold")
	output := fs.String("output", "", "optional JSON output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	raw, err := os.ReadFile(*datasetPath)
	if err != nil {
		return err
	}
	var data dataset
	if err := json.Unmarshal(raw, &data); err != nil {
		return err
	}
	db, err := openDB(*dsn)
	if err != nil {
		return err
	}
	defer db.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	v := verification{ActivityID: data.ActivityID}
	if err := db.QueryRowContext(ctx, `SELECT total,available,sold FROM inventories WHERE activity_id=$1`, data.ActivityID).Scan(&v.Total, &v.Available, &v.Sold); err != nil {
		return err
	}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*),COALESCE(SUM(quantity),0) FROM orders WHERE activity_id=$1`, data.ActivityID).Scan(&v.Orders, &v.OrderedQuantity); err != nil {
		return err
	}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM seckill_records WHERE activity_id=$1`, data.ActivityID).Scan(&v.SeckillRecords); err != nil {
		return err
	}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM (SELECT user_id FROM orders WHERE activity_id=$1 GROUP BY user_id HAVING COUNT(*) > 1) duplicates`, data.ActivityID).Scan(&v.DuplicateBuyers); err != nil {
		return err
	}
	v.InventorySafe = v.Available >= 0 && v.Sold >= 0 && v.Available+v.Sold == v.Total
	v.OrderLimitSafe = v.OrderedQuantity <= v.Total && v.Orders <= v.Total
	v.OneOrderPerBuyer = v.DuplicateBuyers == 0 && v.SeckillRecords == v.Orders
	v.Safe = v.InventorySafe && v.OrderLimitSafe && v.OneOrderPerBuyer
	v.Oversold = v.OrderedQuantity > v.Total || v.Orders > v.Total
	encoded, _ := json.MarshalIndent(v, "", "  ")
	fmt.Println(string(encoded))
	if *output != "" {
		if err := os.MkdirAll(filepath.Dir(*output), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(*output, encoded, 0o644); err != nil {
			return err
		}
	}
	switch *expect {
	case "safe":
		if !v.Safe {
			return errors.New("safe invariants failed")
		}
	case "oversold":
		if !v.Oversold {
			return errors.New("overselling was not reproduced")
		}
	default:
		return fmt.Errorf("unsupported expectation %q", *expect)
	}
	return nil
}

func cleanup(tx *sql.Tx, ctx context.Context) error {
	statements := []string{
		`DELETE FROM seckill_records WHERE activity_id IN (SELECT id FROM activities WHERE name LIKE '[phase2]%')`,
		`DELETE FROM orders WHERE activity_id IN (SELECT id FROM activities WHERE name LIKE '[phase2]%')`,
		`DELETE FROM inventories WHERE activity_id IN (SELECT id FROM activities WHERE name LIKE '[phase2]%')`,
		`DELETE FROM activities WHERE name LIKE '[phase2]%'`,
		`DELETE FROM items WHERE name LIKE '[phase2]%'`,
		`DELETE FROM users WHERE email LIKE 'phase2-load-%@example.test'`,
	}
	for _, statement := range statements {
		if _, err := tx.ExecContext(ctx, statement); err != nil {
			return err
		}
	}
	return nil
}

func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "phase2lab:", err)
	os.Exit(1)
}
