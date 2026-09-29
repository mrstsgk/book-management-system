package config

import (
	"fmt"
	"log/slog"
	"os"
)

type Config struct {
	LogLevel slog.Level
	HTTPPort string
	DB       DBConfig
	Catalog  CatalogConfig
	// AdminToken は書き込み系 API（登録・更新・削除）に必要なトークン。
	AdminToken string
}

type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// CatalogConfig は外部の書籍カタログの接続先。
type CatalogConfig struct {
	OpenBDBaseURL string
}

// Load は環境変数から Config を読み込む。既定値はローカルの docker compose の DB と開発用の管理者トークン
// （デプロイしないので、環境ごとに必須の値を変える仕組みは持たない）。
func Load() (Config, error) {
	var level slog.Level
	if err := level.UnmarshalText([]byte(getenv("LOG_LEVEL", "info"))); err != nil {
		return Config{}, fmt.Errorf("LOG_LEVEL: %w", err)
	}
	return Config{
		LogLevel: level,
		HTTPPort: getenv("HTTP_PORT", "8080"),
		DB: DBConfig{
			Host:     getenv("DB_HOST", "localhost"),
			Port:     getenv("DB_PORT", "5432"),
			User:     getenv("DB_USER", "postgres"),
			Password: getenv("DB_PASSWORD", "postgres"),
			DBName:   getenv("DB_NAME", "book_management"),
			SSLMode:  getenv("DB_SSLMODE", "disable"),
		},
		Catalog: CatalogConfig{
			OpenBDBaseURL: getenv("OPENBD_BASE_URL", "https://api.openbd.jp"),
		},
		AdminToken: getenv("ADMIN_TOKEN", "local-admin-token"),
	}, nil
}

// NewLogger はテキスト形式でログを出す。
func (c Config) NewLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: c.LogLevel}))
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
