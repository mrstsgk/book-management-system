package config_test

import (
	"log/slog"
	"testing"

	"github.com/mrstsgk/book-management-system/backend/config"
)

var dbKeys = []string{"DB_HOST", "DB_PORT", "DB_USER", "DB_PASSWORD", "DB_NAME", "DB_SSLMODE"}

var catalogKeys = []string{"OPENBD_BASE_URL", "RAKUTEN_BASE_URL", "RAKUTEN_APPLICATION_ID", "RAKUTEN_ACCESS_KEY"}

var requiredKeys = []string{"DB_HOST", "DB_USER", "DB_PASSWORD", "DB_NAME", "ADMIN_TOKEN"}

const strongToken = "0123456789abcdef0123456789abcdef" // 32文字

func clearEnv(t *testing.T) {
	t.Helper()
	for _, k := range append(append([]string{"STAGE", "LOG_LEVEL", "HTTP_PORT"}, dbKeys...), append(catalogKeys, "ADMIN_TOKEN")...) {
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

	t.Run("non-local stage fails when a required setting is missing", func(t *testing.T) {
		for _, missing := range requiredKeys {
			t.Run(missing, func(t *testing.T) {
				clearEnv(t)
				t.Setenv("STAGE", "stg")
				for _, k := range requiredKeys {
					if k != missing {
						t.Setenv(k, strongToken)
					}
				}
				if _, err := config.Load(); err == nil {
					t.Fatalf("expected error when %s is missing", missing)
				}
			})
		}
	})

	t.Run("non-local stage succeeds when required settings are given", func(t *testing.T) {
		clearEnv(t)
		t.Setenv("STAGE", "prd")
		t.Setenv("DB_HOST", "db.internal")
		t.Setenv("DB_USER", "app")
		t.Setenv("DB_PASSWORD", "secret")
		t.Setenv("DB_NAME", "book_management")
		t.Setenv("ADMIN_TOKEN", strongToken)
		cfg, err := config.Load()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.DB.Host != "db.internal" {
			t.Errorf("DB.Host = %q, want db.internal", cfg.DB.Host)
		}
		if cfg.Catalog.RakutenEnabled() {
			t.Error("Rakuten must stay optional outside local too")
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

func TestLoad_AdminToken(t *testing.T) {
	setRequired := func(t *testing.T, token string) {
		t.Helper()
		clearEnv(t)
		t.Setenv("STAGE", "prd")
		t.Setenv("DB_HOST", "db.internal")
		t.Setenv("DB_USER", "app")
		t.Setenv("DB_PASSWORD", "secret")
		t.Setenv("DB_NAME", "book_management")
		t.Setenv("ADMIN_TOKEN", token)
	}

	t.Run("non-localで32文字ちょうどは有効", func(t *testing.T) {
		setRequired(t, strongToken)
		cfg, err := config.Load()
		if err != nil || cfg.AdminToken != strongToken {
			t.Fatalf("got (%q, %v)", cfg.AdminToken, err)
		}
	})

	t.Run("non-localで31文字はエラー", func(t *testing.T) {
		setRequired(t, strongToken[:31])
		if _, err := config.Load(); err == nil {
			t.Fatal("expected an error for a short ADMIN_TOKEN")
		}
	})

	t.Run("localでは短いトークンも指定できる", func(t *testing.T) {
		clearEnv(t)
		t.Setenv("ADMIN_TOKEN", "dev")
		cfg, err := config.Load()
		if err != nil || cfg.AdminToken != "dev" {
			t.Fatalf("got (%q, %v)", cfg.AdminToken, err)
		}
	})
}
