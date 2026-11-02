package config

import (
	"fmt"
	"log/slog"
	"os"
)

const stageLocal = "local"

type Config struct {
	// Stage は動かす環境（既定は local）。local 以外では DB の接続先の指定が必須で、ログは JSON になる。
	Stage    string
	LogLevel slog.Level
	HTTPPort string
	DB       DBConfig
	Catalog  CatalogConfig
	// Admin は管理画面のログインに使う ID とパスワードの bcrypt ハッシュ。既定値は無い（未設定ならログインできない）。
	Admin AdminConfig
}

// AdminConfig は管理者 1 人分の資格情報。ハッシュは `go run ./cmd/hashpw` で作る。
type AdminConfig struct {
	ID           string
	PasswordHash string
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

// Load は環境変数から Config を読み込む。defaultHTTPPort は実行サービスごとの既定の待ち受けポート
// （public-api と admin-api を同じホストで並べて動かすため）。
// DB の既定値はローカルの docker compose の DB だけに置き、local 以外では DB_HOST・DB_USER・DB_PASSWORD・DB_NAME を必須にする
// （設定漏れのまま localhost に繋ぎに行くより、起動時に止まる方が気づけるため）。
// 管理者の ID・パスワードには既定値を置かない（設定漏れで書き込めるようにしないため）。
func Load(defaultHTTPPort string) (Config, error) {
	stage := getenv("STAGE", stageLocal)
	local := stage == stageLocal

	var level slog.Level
	if err := level.UnmarshalText([]byte(getenv("LOG_LEVEL", "info"))); err != nil {
		return Config{}, fmt.Errorf("LOG_LEVEL: %w", err)
	}

	dbDefault := func(v string) string {
		if local {
			return v
		}
		return ""
	}
	db := DBConfig{
		Host:     getenv("DB_HOST", dbDefault("localhost")),
		Port:     getenv("DB_PORT", "5432"),
		User:     getenv("DB_USER", dbDefault("postgres")),
		Password: getenv("DB_PASSWORD", dbDefault("postgres")),
		DBName:   getenv("DB_NAME", dbDefault("book_management")),
		SSLMode:  getenv("DB_SSLMODE", "disable"),
	}
	for _, kv := range [][2]string{{"DB_HOST", db.Host}, {"DB_USER", db.User}, {"DB_PASSWORD", db.Password}, {"DB_NAME", db.DBName}} {
		if kv[1] == "" {
			return Config{}, fmt.Errorf("%s is required when STAGE=%s", kv[0], stage)
		}
	}

	return Config{
		Stage:    stage,
		LogLevel: level,
		HTTPPort: getenv("HTTP_PORT", defaultHTTPPort),
		DB:       db,
		Catalog: CatalogConfig{
			OpenBDBaseURL: getenv("OPENBD_BASE_URL", "https://api.openbd.jp"),
		},
		Admin: AdminConfig{
			ID:           os.Getenv("ADMIN_ID"),
			PasswordHash: os.Getenv("ADMIN_PASSWORD_HASH"),
		},
	}, nil
}

// NewLogger は local では読みやすいテキスト、それ以外では集計しやすい JSON でログを出す。
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
