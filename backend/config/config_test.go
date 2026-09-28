package config_test

import (
	"log/slog"
	"testing"

	"github.com/mrstsgk/book-management-system/backend/config"
)

var dbKeys = []string{"DB_HOST", "DB_PORT", "DB_USER", "DB_PASSWORD", "DB_NAME", "DB_SSLMODE"}

func clearEnv(t *testing.T) {
	t.Helper()
	for _, k := range append([]string{"STAGE", "LOG_LEVEL", "HTTP_PORT"}, dbKeys...) {
		t.Setenv(k, "")
	}
}

func TestLoad(t *testing.T) {
	t.Run("defaults to local stage with local DB and info level", func(t *testing.T) {
		clearEnv(t)
		cfg, err := config.Load()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.Stage != "local" {
			t.Errorf("Stage = %q, want local", cfg.Stage)
		}
		if cfg.LogLevel != slog.LevelInfo {
			t.Errorf("LogLevel = %v, want INFO", cfg.LogLevel)
		}
		if cfg.DB.Host == "" || cfg.DB.User == "" || cfg.DB.DBName == "" {
			t.Errorf("local DB defaults missing: %+v", cfg.DB)
		}
	})

	t.Run("LOG_LEVEL is parsed", func(t *testing.T) {
		clearEnv(t)
		t.Setenv("LOG_LEVEL", "debug")
		cfg, err := config.Load()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.LogLevel != slog.LevelDebug {
			t.Errorf("LogLevel = %v, want DEBUG", cfg.LogLevel)
		}
	})

	t.Run("invalid LOG_LEVEL fails", func(t *testing.T) {
		clearEnv(t)
		t.Setenv("LOG_LEVEL", "verbose")
		if _, err := config.Load(); err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("non-local stage fails when a DB setting is missing", func(t *testing.T) {
		for _, missing := range []string{"DB_HOST", "DB_USER", "DB_PASSWORD", "DB_NAME"} {
			t.Run(missing, func(t *testing.T) {
				clearEnv(t)
				t.Setenv("STAGE", "stg")
				for _, k := range []string{"DB_HOST", "DB_USER", "DB_PASSWORD", "DB_NAME"} {
					if k != missing {
						t.Setenv(k, "x")
					}
				}
				if _, err := config.Load(); err == nil {
					t.Fatalf("expected error when %s is missing", missing)
				}
			})
		}
	})

	t.Run("non-local stage succeeds when DB settings are given", func(t *testing.T) {
		clearEnv(t)
		t.Setenv("STAGE", "prd")
		t.Setenv("DB_HOST", "db.internal")
		t.Setenv("DB_USER", "app")
		t.Setenv("DB_PASSWORD", "secret")
		t.Setenv("DB_NAME", "book_management")
		cfg, err := config.Load()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.DB.Host != "db.internal" {
			t.Errorf("DB.Host = %q, want db.internal", cfg.DB.Host)
		}
	})
}
