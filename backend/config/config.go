package config

import (
	"fmt"
	"log/slog"
	"os"
)

const (
	stageLocal          = "local"
	adminTokenMinLength = 32
)

type Config struct {
	Stage    string
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

// CatalogConfig は外部の書籍カタログの接続先。楽天のキーは任意で、無ければ openBD だけを使う
// （openBD に書影が無い本は書影なしになる）。
type CatalogConfig struct {
	OpenBDBaseURL        string
	RakutenBaseURL       string
	RakutenApplicationID string
	RakutenAccessKey     string
}

// RakutenEnabled は楽天のアプリ ID とアクセスキーが両方設定されているかを返す。
func (c CatalogConfig) RakutenEnabled() bool {
	return c.RakutenApplicationID != "" && c.RakutenAccessKey != ""
}

// Load は環境変数から Config を読み込む。non-local では DB_HOST・DB_USER・
// DB_PASSWORD・DB_NAME・ADMIN_TOKEN（32文字以上）の明示指定が必須（DB_PORT は 5432、
// DB_SSLMODE は disable を常にデフォルト値として使う）。
func Load() (Config, error) {
	stage := getenv("STAGE", stageLocal)
	local := stage == stageLocal

	var level slog.Level
	if err := level.UnmarshalText([]byte(getenv("LOG_LEVEL", "info"))); err != nil {
		return Config{}, fmt.Errorf("LOG_LEVEL: %w", err)
	}

	// 既定値はローカルの docker compose の DB 用だけ。それ以外の環境で設定が漏れたら、
	// 黙って localhost に向けずに起動を止める。
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
	adminToken := getenv("ADMIN_TOKEN", localDefault("local-admin-token"))
	for k, v := range map[string]string{
		"DB_HOST": db.Host, "DB_USER": db.User, "DB_PASSWORD": db.Password, "DB_NAME": db.DBName,
		"ADMIN_TOKEN": adminToken,
	} {
		if v == "" {
			return Config{}, fmt.Errorf("%s is required when STAGE=%s", k, stage)
		}
	}
	// 公開する環境では推測されにくい長さを求める（ローカルの既定値は開発用なので対象外）
	if !local && len(adminToken) < adminTokenMinLength {
		return Config{}, fmt.Errorf("ADMIN_TOKEN must be at least %d characters when STAGE=%s", adminTokenMinLength, stage)
	}

	return Config{
		Stage:    stage,
		LogLevel: level,
		HTTPPort: getenv("HTTP_PORT", "8080"),
		DB:       db,
		Catalog:  catalog,

		AdminToken: adminToken,
	}, nil
}

// NewLogger はローカルでは読みやすいテキスト、それ以外ではログ集約向けの JSON でログを出す。
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
