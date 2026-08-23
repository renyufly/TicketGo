package config

import "testing"

func TestLoadDefaults(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://example")
	t.Setenv("JWT_SECRET", "0123456789abcdef0123456789abcdef")
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.HTTP.Address != ":8080" {
		t.Fatalf("Address = %q", cfg.HTTP.Address)
	}
	if cfg.Database.MaxOpenConns != 25 {
		t.Fatalf("MaxOpenConns = %d", cfg.Database.MaxOpenConns)
	}
	if cfg.Auth.AllowAdminSelfRegistration {
		t.Fatal("AllowAdminSelfRegistration = true, want secure default false")
	}
	if cfg.Seckill.Strategy != "atomic" {
		t.Fatalf("Seckill.Strategy = %q, want atomic", cfg.Seckill.Strategy)
	}
}

func TestLoadRejectsUnknownInventoryStrategy(t *testing.T) {
	t.Setenv("JWT_SECRET", "0123456789abcdef0123456789abcdef")
	t.Setenv("SECKILL_INVENTORY_STRATEGY", "redis-before-phase3")
	if _, err := Load(); err == nil {
		t.Fatal("Load() error = nil, want inventory strategy validation error")
	}
}

func TestLoadRejectsInvalidDuration(t *testing.T) {
	t.Setenv("JWT_SECRET", "0123456789abcdef0123456789abcdef")
	t.Setenv("HTTP_REQUEST_TIMEOUT", "eventually")
	if _, err := Load(); err == nil {
		t.Fatal("Load() error = nil, want validation error")
	}
}

func TestLoadRejectsOversizedIdlePool(t *testing.T) {
	t.Setenv("JWT_SECRET", "0123456789abcdef0123456789abcdef")
	t.Setenv("DB_MAX_OPEN_CONNS", "2")
	t.Setenv("DB_MAX_IDLE_CONNS", "3")
	if _, err := Load(); err == nil {
		t.Fatal("Load() error = nil, want validation error")
	}
}

func TestLoadRejectsWeakJWTSecret(t *testing.T) {
	t.Setenv("JWT_SECRET", "short")
	if _, err := Load(); err == nil {
		t.Fatal("Load() error = nil, want JWT secret validation error")
	}
}

func TestLoadAdminSelfRegistrationFlag(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://example")
	t.Setenv("JWT_SECRET", "0123456789abcdef0123456789abcdef")
	t.Setenv("ALLOW_ADMIN_SELF_REGISTRATION", "true")
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.Auth.AllowAdminSelfRegistration {
		t.Fatal("AllowAdminSelfRegistration = false, want true")
	}

	t.Setenv("ALLOW_ADMIN_SELF_REGISTRATION", "sometimes")
	if _, err := Load(); err == nil {
		t.Fatal("Load() error = nil, want boolean validation error")
	}
}
