package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type Config struct {
	URL             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
	ConnectTimeout  time.Duration
	QueryTimeout    time.Duration
}

type DB struct {
	*sql.DB
	queryTimeout time.Duration
}

func Open(parent context.Context, cfg Config) (*DB, error) {
	db, err := sql.Open("pgx", cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("open connection pool: %w", err)
	}
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	db.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)

	ctx, cancel := context.WithTimeout(parent, cfg.ConnectTimeout)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping database within %s: %w", cfg.ConnectTimeout, err)
	}
	return &DB{DB: db, queryTimeout: cfg.QueryTimeout}, nil
}

func (db *DB) QueryTimeout() time.Duration { return db.queryTimeout }
func (db *DB) SQL() *sql.DB                { return db.DB }
