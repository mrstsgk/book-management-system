package book_test

import (
	"testing"

	"gorm.io/gorm"

	pgcommon "github.com/mrstsgk/book-management-system/backend/internal/infrastructure/postgres/common"
)

// connectTestDB はローカルの開発用 DB（backend/docker-compose.yml、マイグレーション済み）に接続する。
// 実際の PostgreSQL に対する契約テスト（単体テストではない）なので、DB に繋がらなければ
// 失敗ではなく skip する（docs/rules/testing.md）。
func connectTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := pgcommon.Connect(pgcommon.Config{
		Host:     "localhost",
		Port:     "5432",
		User:     "postgres",
		Password: "postgres",
		DBName:   "book_management",
		SSLMode:  "disable",
	})
	if err != nil {
		t.Skipf("skipping: local Postgres not reachable (run `make db-up migrate-up` first): %v", err)
	}
	if err := db.Exec("SELECT 1 FROM book LIMIT 0").Error; err != nil {
		t.Skipf("skipping: schema not migrated (run `make migrate-up` first): %v", err)
	}
	return db
}
