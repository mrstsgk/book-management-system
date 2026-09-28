package config_test

import (
	"log/slog"
	"testing"

	"github.com/mrstsgk/book-management-system/backend/config"
)

var dbKeys = []string{"DB_HOST", "DB_PORT", "DB_USER", "DB_PASSWORD", "DB_NAME", "DB_SSLMODE"}

var catalogKeys = []string{"OPENBD_BASE_URL", "RAKUTEN_BASE_URL", "RAKUTEN_APPLICATION_ID", "RAKUTEN_ACCESS_KEY"}

func clearEnv(t *testing.T) {
	t.Helper()
	for _, k := range append(append([]string{"LOG_LEVEL", "HTTP_PORT"}, dbKeys...), append(catalogKeys, "ADMIN_TOKEN")...) {
		t.Setenv(k, "")
	}
}

func TestLoad(t *testing.T) {
	t.Run("未設定ならローカルのDBと開発用の管理者トークンとinfoレベルを使う", func(t *testing.T) {
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
		want := config.CatalogConfig{OpenBDBaseURL: "https://api.openbd.jp", RakutenBaseURL: "https://openapi.rakuten.co.jp"}
		if cfg.Catalog != want {
			t.Errorf("catalog defaults = %+v, want %+v", cfg.Catalog, want)
		}
		if cfg.Catalog.RakutenEnabled() {
			t.Error("Rakuten must be disabled without its keys")
		}
		if cfg.AdminToken != "local-admin-token" {
			t.Errorf("AdminToken = %q, want the local development default", cfg.AdminToken)
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

func TestCatalogConfig_RakutenEnabled(t *testing.T) {
	tests := []struct {
		name        string
		appID, key  string
		wantEnabled bool
	}{
		{name: "アプリIDとアクセスキーの両方があれば有効", appID: "app", key: "key", wantEnabled: true},
		{name: "アクセスキーが無ければ無効", appID: "app", key: "", wantEnabled: false},
		{name: "アプリIDが無ければ無効", appID: "", key: "key", wantEnabled: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearEnv(t)
			t.Setenv("RAKUTEN_APPLICATION_ID", tt.appID)
			t.Setenv("RAKUTEN_ACCESS_KEY", tt.key)
			cfg, err := config.Load()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got := cfg.Catalog.RakutenEnabled(); got != tt.wantEnabled {
				t.Fatalf("RakutenEnabled() = %v, want %v", got, tt.wantEnabled)
			}
		})
	}
}

func TestLoad_CatalogBaseURLsCanBeOverridden(t *testing.T) {
	clearEnv(t)
	t.Setenv("OPENBD_BASE_URL", "http://openbd.test")
	t.Setenv("RAKUTEN_BASE_URL", "http://rakuten.test")
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Catalog.OpenBDBaseURL != "http://openbd.test" || cfg.Catalog.RakutenBaseURL != "http://rakuten.test" {
		t.Fatalf("catalog = %+v", cfg.Catalog)
	}
}

func TestLoad_AdminTokenCanBeOverridden(t *testing.T) {
	clearEnv(t)
	t.Setenv("ADMIN_TOKEN", "my-token")
	cfg, err := config.Load()
	if err != nil || cfg.AdminToken != "my-token" {
		t.Fatalf("got (%q, %v), want my-token", cfg.AdminToken, err)
	}
}
