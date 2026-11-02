package postgres

import (
	"context"
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// プールの上限はプロセスごと。プロセス数 × maxOpenConns が PostgreSQL の max_connections を下回るようにする。
const (
	maxOpenConns    = 20
	maxIdleConns    = 10
	connMaxLifetime = time.Hour
	pingTimeout     = 5 * time.Second
)

type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// Connect はプールを開いて 1 度 ping し、DB 設定の誤りを最初のリクエストではなく起動時に失敗させる。
func Connect(cfg Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=Asia/Tokyo",
		cfg.Host, cfg.User, cfg.Password, cfg.DBName, cfg.Port, cfg.SSLMode,
	)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
		// GORM 自身の ping にはタイムアウトが無いので、下でタイムアウト付きの ping を行う。
		DisableAutomaticPing: true,
		NowFunc: func() time.Time {
			return time.Now().In(time.FixedZone("Asia/Tokyo", 9*60*60))
		},
	})
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get sql.DB: %w", err)
	}
	sqlDB.SetMaxOpenConns(maxOpenConns)
	sqlDB.SetMaxIdleConns(maxIdleConns)
	sqlDB.SetConnMaxLifetime(connMaxLifetime)

	ctx, cancel := context.WithTimeout(context.Background(), pingTimeout)
	defer cancel()
	if err := sqlDB.PingContext(ctx); err != nil {
		_ = sqlDB.Close()
		return nil, fmt.Errorf("ping db %s:%s/%s: %w", cfg.Host, cfg.Port, cfg.DBName, err)
	}
	return db, nil
}

// Close はプールを解放する（graceful shutdown で呼ぶ）。
func Close(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
