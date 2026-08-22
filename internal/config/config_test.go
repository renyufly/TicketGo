package config

import "testing"

func TestLoadDefaults(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://example")
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
}

func TestLoadRejectsInvalidDuration(t *testing.T) {
	t.Setenv("HTTP_REQUEST_TIMEOUT", "eventually")
	if _, err := Load(); err == nil {
		t.Fatal("Load() error = nil, want validation error")
	}
}

func TestLoadRejectsOversizedIdlePool(t *testing.T) {
	t.Setenv("DB_MAX_OPEN_CONNS", "2")
	t.Setenv("DB_MAX_IDLE_CONNS", "3")
	if _, err := Load(); err == nil {
		t.Fatal("Load() error = nil, want validation error")
	}
}
