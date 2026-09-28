package config

import (
	"fmt"
	"log/slog"
	"os"
)

const stageLocal = "local"

type Config struct {
	Stage    string
	LogLevel slog.Level
	HTTPPort string
	DB       DBConfig
	Catalog  CatalogConfig
}

type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// CatalogConfig points at the external book catalogs. The Rakuten keys are optional:
// without them only openBD is used (and books it has no cover for show none).
type CatalogConfig struct {
	OpenBDBaseURL        string
	RakutenBaseURL       string
	RakutenApplicationID string
	RakutenAccessKey     string
}

// RakutenEnabled reports whether both Rakuten keys are set.
func (c CatalogConfig) RakutenEnabled() bool {
	return c.RakutenApplicationID != "" && c.RakutenAccessKey != ""
}

// Load は環境変数から Config を読み込む。non-local では DB_HOST・DB_USER・
// DB_PASSWORD・DB_NAME の明示指定が必須（DB_PORT は 5432、DB_SSLMODE は
// disable を常にデフォルト値として使う）。
func Load() (Config, error) {
	stage := getenv("STAGE", stageLocal)
	local := stage == stageLocal

	var level slog.Level
	if err := level.UnmarshalText([]byte(getenv("LOG_LEVEL", "info"))); err != nil {
		return Config{}, fmt.Errorf("LOG_LEVEL: %w", err)
	}

	// Defaults exist only for the local docker-compose DB; elsewhere a missing
	// setting must stop startup instead of silently pointing at localhost.
	localDefault := func(v string) string {
		if local {
			return v
		}
		return ""
	}
	db := DBConfig{
		Host:     getenv("DB_HOST", localDefault("localhost")),
		Port:     getenv("DB_PORT", "5432"),
		User:     getenv("DB_USER", localDefault("postgres")),
		Password: getenv("DB_PASSWORD", localDefault("postgres")),
		DBName:   getenv("DB_NAME", localDefault("book_management")),
		SSLMode:  getenv("DB_SSLMODE", "disable"),
	}
	catalog := CatalogConfig{
		OpenBDBaseURL:        getenv("OPENBD_BASE_URL", "https://api.openbd.jp"),
		RakutenBaseURL:       getenv("RAKUTEN_BASE_URL", "https://openapi.rakuten.co.jp"),
		RakutenApplicationID: getenv("RAKUTEN_APPLICATION_ID", ""),
		RakutenAccessKey:     getenv("RAKUTEN_ACCESS_KEY", ""),
	}
	for k, v := range map[string]string{
		"DB_HOST": db.Host, "DB_USER": db.User, "DB_PASSWORD": db.Password, "DB_NAME": db.DBName,
	} {
		if v == "" {
			return Config{}, fmt.Errorf("%s is required when STAGE=%s", k, stage)
		}
	}

	return Config{
		Stage:    stage,
		LogLevel: level,
		HTTPPort: getenv("HTTP_PORT", "8080"),
		DB:       db,
		Catalog:  catalog,
	}, nil
}

// NewLogger returns text logs for local reading and JSON elsewhere (log aggregation).
func (c Config) NewLogger() *slog.Logger {
	opts := &slog.HandlerOptions{Level: c.LogLevel}
	if c.Stage == stageLocal {
		return slog.New(slog.NewTextHandler(os.Stdout, opts))
	}
	return slog.New(slog.NewJSONHandler(os.Stdout, opts))
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
