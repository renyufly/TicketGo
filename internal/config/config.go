package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"ticketgo/internal/order"
	"ticketgo/pkg/database"
)

type Config struct {
	Environment string
	HTTP        HTTPConfig
	Database    database.Config
	Auth        AuthConfig
	Seckill     order.ConcurrencyConfig
}

type AuthConfig struct {
	JWTSecret                  string
	TokenTTL                   time.Duration
	AllowAdminSelfRegistration bool
}

type HTTPConfig struct {
	Address           string
	ReadTimeout       time.Duration
	ReadHeaderTimeout time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
	RequestTimeout    time.Duration
	ShutdownTimeout   time.Duration
}

func Load() (Config, error) {
	cfg := Config{
		Environment: env("APP_ENV", "development"),
		HTTP: HTTPConfig{
			Address: env("HTTP_ADDRESS", ":8080"),
		},
		Database: database.Config{
			URL: env("DATABASE_URL", "postgres://ticketgo:ticketgo_local_password@localhost:5432/ticketgo?sslmode=disable"),
		},
		Auth:    AuthConfig{JWTSecret: env("JWT_SECRET", "")},
		Seckill: order.DefaultConcurrencyConfig(),
	}

	var err error
	if cfg.HTTP.ReadTimeout, err = duration("HTTP_READ_TIMEOUT", "5s"); err != nil {
		return Config{}, err
	}
	if cfg.HTTP.ReadHeaderTimeout, err = duration("HTTP_READ_HEADER_TIMEOUT", "2s"); err != nil {
		return Config{}, err
	}
	if cfg.HTTP.WriteTimeout, err = duration("HTTP_WRITE_TIMEOUT", "10s"); err != nil {
		return Config{}, err
	}
	if cfg.HTTP.IdleTimeout, err = duration("HTTP_IDLE_TIMEOUT", "60s"); err != nil {
		return Config{}, err
	}
	if cfg.HTTP.RequestTimeout, err = duration("HTTP_REQUEST_TIMEOUT", "3s"); err != nil {
		return Config{}, err
	}
	if cfg.HTTP.ShutdownTimeout, err = duration("HTTP_SHUTDOWN_TIMEOUT", "10s"); err != nil {
		return Config{}, err
	}
	if cfg.Database.MaxOpenConns, err = integer("DB_MAX_OPEN_CONNS", 25); err != nil {
		return Config{}, err
	}
	if cfg.Database.MaxIdleConns, err = integer("DB_MAX_IDLE_CONNS", 10); err != nil {
		return Config{}, err
	}
	if cfg.Database.ConnMaxLifetime, err = duration("DB_CONN_MAX_LIFETIME", "30m"); err != nil {
		return Config{}, err
	}
	if cfg.Database.ConnMaxIdleTime, err = duration("DB_CONN_MAX_IDLE_TIME", "5m"); err != nil {
		return Config{}, err
	}
	if cfg.Database.ConnectTimeout, err = duration("DB_CONNECT_TIMEOUT", "5s"); err != nil {
		return Config{}, err
	}
	if cfg.Database.QueryTimeout, err = duration("DB_QUERY_TIMEOUT", "3s"); err != nil {
		return Config{}, err
	}
	if cfg.Auth.TokenTTL, err = duration("AUTH_TOKEN_TTL", "24h"); err != nil {
		return Config{}, err
	}
	if cfg.Auth.AllowAdminSelfRegistration, err = boolean("ALLOW_ADMIN_SELF_REGISTRATION", false); err != nil {
		return Config{}, err
	}
	if cfg.Seckill.Strategy, err = order.ParseInventoryStrategy(env("SECKILL_INVENTORY_STRATEGY", string(order.StrategyAtomic))); err != nil {
		return Config{}, err
	}
	if cfg.Seckill.NaiveDelay, err = nonNegativeDuration("SECKILL_NAIVE_DELAY", "0s"); err != nil {
		return Config{}, err
	}
	if cfg.Seckill.LockTimeout, err = duration("SECKILL_LOCK_TIMEOUT", "500ms"); err != nil {
		return Config{}, err
	}
	if cfg.Seckill.StatementTimeout, err = duration("SECKILL_STATEMENT_TIMEOUT", "2500ms"); err != nil {
		return Config{}, err
	}
	if cfg.Seckill.OptimisticRetries, err = nonNegativeInteger("SECKILL_OPTIMISTIC_MAX_RETRIES", 5); err != nil {
		return Config{}, err
	}
	if cfg.Seckill.OptimisticBackoff, err = nonNegativeDuration("SECKILL_OPTIMISTIC_BACKOFF", "2ms"); err != nil {
		return Config{}, err
	}

	if cfg.Database.URL == "" {
		return Config{}, errors.New("DATABASE_URL must not be empty")
	}
	if cfg.Database.MaxIdleConns > cfg.Database.MaxOpenConns {
		return Config{}, errors.New("DB_MAX_IDLE_CONNS must not exceed DB_MAX_OPEN_CONNS")
	}
	if len(cfg.Auth.JWTSecret) < 32 {
		return Config{}, errors.New("JWT_SECRET must contain at least 32 characters")
	}
	return cfg, nil
}

func env(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func duration(key, fallback string) (time.Duration, error) {
	value := env(key, fallback)
	parsed, err := time.ParseDuration(value)
	if err != nil || parsed <= 0 {
		return 0, fmt.Errorf("%s must be a positive Go duration, got %q", key, value)
	}
	return parsed, nil
}

func nonNegativeDuration(key, fallback string) (time.Duration, error) {
	value := env(key, fallback)
	parsed, err := time.ParseDuration(value)
	if err != nil || parsed < 0 {
		return 0, fmt.Errorf("%s must be a non-negative Go duration, got %q", key, value)
	}
	return parsed, nil
}

func integer(key string, fallback int) (int, error) {
	value := env(key, strconv.Itoa(fallback))
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return 0, fmt.Errorf("%s must be a positive integer, got %q", key, value)
	}
	return parsed, nil
}

func nonNegativeInteger(key string, fallback int) (int, error) {
	value := env(key, strconv.Itoa(fallback))
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < 0 {
		return 0, fmt.Errorf("%s must be a non-negative integer, got %q", key, value)
	}
	return parsed, nil
}

func boolean(key string, fallback bool) (bool, error) {
	value := env(key, strconv.FormatBool(fallback))
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return false, fmt.Errorf("%s must be a boolean, got %q", key, value)
	}
	return parsed, nil
}
