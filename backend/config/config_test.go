package config_test

import (
	"log/slog"
	"testing"

	"github.com/mrstsgk/book-management-system/backend/config"
)

var dbKeys = []string{"DB_HOST", "DB_PORT", "DB_USER", "DB_PASSWORD", "DB_NAME", "DB_SSLMODE"}

var catalogKeys = []string{"OPENBD_BASE_URL"}

func clearEnv(t *testing.T) {
	t.Helper()
	for _, k := range append(append([]string{"LOG_LEVEL", "HTTP_PORT"}, dbKeys...), append(catalogKeys, "ADMIN_ID", "ADMIN_PASSWORD_HASH")...) {
		t.Setenv(k, "")
	}
}

func TestLoad(t *testing.T) {
	t.Run("未設定ならローカルのDBとinfoレベルを使い、管理者の既定値は置かない", func(t *testing.T) {
		clearEnv(t)
		cfg, err := config.Load()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.LogLevel != slog.LevelInfo {
			t.Errorf("LogLevel = %v, want INFO", cfg.LogLevel)
		}
		wantDB := config.DBConfig{Host: "localhost", Port: "5432", User: "postgres", Password: "postgres", DBName: "book_management", SSLMode: "disable"}
		if cfg.DB != wantDB {
			t.Errorf("DB = %+v, want %+v", cfg.DB, wantDB)
		}
		want := config.CatalogConfig{OpenBDBaseURL: "https://api.openbd.jp"}
		if cfg.Catalog != want {
			t.Errorf("catalog defaults = %+v, want %+v", cfg.Catalog, want)
		}
		if cfg.Admin.ID != "" || cfg.Admin.PasswordHash != "" {
			t.Errorf("Admin = %+v, want no default (a missing setting must not open write access)", cfg.Admin)
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

	t.Run("DBの設定は環境変数で上書きできる", func(t *testing.T) {
		clearEnv(t)
		t.Setenv("DB_HOST", "db.test")
		t.Setenv("DB_NAME", "other")
		cfg, err := config.Load()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.DB.Host != "db.test" || cfg.DB.DBName != "other" || cfg.DB.User != "postgres" {
			t.Errorf("DB = %+v, want overridden host/name with default user", cfg.DB)
		}
	})
}

func TestLoad_CatalogBaseURLsCanBeOverridden(t *testing.T) {
	clearEnv(t)
	t.Setenv("OPENBD_BASE_URL", "http://openbd.test")
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Catalog.OpenBDBaseURL != "http://openbd.test" {
		t.Fatalf("catalog = %+v", cfg.Catalog)
	}
}

func TestLoad_Admin(t *testing.T) {
	clearEnv(t)
	t.Setenv("ADMIN_ID", "me")
	t.Setenv("ADMIN_PASSWORD_HASH", "$2a$12$x")
	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Admin.ID != "me" || cfg.Admin.PasswordHash != "$2a$12$x" {
		t.Fatalf("got %+v", cfg.Admin)
	}
}
