package book_test

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"gorm.io/gorm"

	domainbook "github.com/mrstsgk/book-management-system/backend/internal/domain/book"
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

// uniqueCoverTestISBN は書影の契約テスト用に、開発用 DB にまだ無い ISBN（9780000007000〜7099 の範囲）を返す。
// チェックディジットが決まるので範囲内に10通りしか無く、固定にすると前回のクリーンアップ漏れで衝突するため、空いているものを探す。
func uniqueCoverTestISBN(t *testing.T, db *gorm.DB) string {
	t.Helper()
	start := int(time.Now().UnixNano() % 10)
	for i := range 10 {
		first12 := fmt.Sprintf("97800000070%d", (start+i)%10)
		isbn := first12 + checkDigit13(first12)
		var n int64
		if err := db.Raw("SELECT COUNT(*) FROM book WHERE isbn = ?", isbn).Scan(&n).Error; err != nil {
			t.Fatal(err)
		}
		if n == 0 {
			return isbn
		}
	}
	t.Fatal("no free ISBN left in 9780000007000-9780000007099; delete leftover rows from earlier runs")
	return ""
}

func checkDigit13(first12 string) string {
	sum := 0
	for i, r := range first12 {
		d := int(r - '0')
		if i%2 == 1 {
			d *= 3
		}
		sum += d
	}
	return strconv.Itoa((10 - sum%10) % 10)
}

func mustGoogleBooksCover(t *testing.T, id string) *domainbook.Cover {
	t.Helper()
	c, err := domainbook.NewGoogleBooksCover("https://books.google.com/books/content?id="+id+"&printsec=frontcover&img=1&zoom=1", "https://books.google.co.jp/books?id="+id)
	if err != nil {
		t.Fatal(err)
	}
	return &c
}

// mustCreateTag はテスト用のタグを1件挿入し、採番されたIDを返す。挿入に失敗したらここで即座に落とす
// （呼び出し側で気づかず後続の外部キー違反などで原因が分かりにくくなるのを防ぐため）。
func mustCreateTag(t *testing.T, db *gorm.DB, name string) int64 {
	t.Helper()
	var id int64
	if err := db.Raw("INSERT INTO tag (name, version) VALUES (?, 1) RETURNING id", name).Scan(&id).Error; err != nil {
		t.Fatalf("insert tag %q: %v", name, err)
	}
	return id
}
