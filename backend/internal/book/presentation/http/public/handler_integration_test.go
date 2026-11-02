package public_test

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	"gorm.io/gorm"

	bookdomain "github.com/mrstsgk/book-management-system/backend/internal/book/domain"
	pgbook "github.com/mrstsgk/book-management-system/backend/internal/book/infrastructure/postgres"
	bookpublic "github.com/mrstsgk/book-management-system/backend/internal/book/presentation/http/public"
	bookqry "github.com/mrstsgk/book-management-system/backend/internal/book/usecase/query"
	pgplatform "github.com/mrstsgk/book-management-system/backend/internal/platform/postgres"
)

// connectIntegrationDB はローカルの開発用 DB に接続する。実際の PostgreSQL に対する結合テストなので、
// DB に繋がらなければ skip せず失敗にする（契約テストと同じ理由。docs/rules/testing.md）。
func connectIntegrationDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := pgplatform.Connect(pgplatform.Config{
		Host: "localhost", Port: "5432", User: "postgres", Password: "postgres",
		DBName: "book_management", SSLMode: "disable",
	})
	if err != nil {
		t.Fatalf("local Postgres not reachable (run `make db-up migrate-up` first; DB を使うテストは skip しない。docs/rules/testing.md): %v", err)
	}
	if err := db.Exec("SELECT 1 FROM book LIMIT 0").Error; err != nil {
		t.Fatalf("schema not migrated (run `make migrate-up` first): %v", err)
	}
	return db
}

// newIntegrationHandler は本物の Repository/Query/UseCase を実DBに配線した Handler を返す（AdminOnly は serve() が付ける）。
func newIntegrationHandler(db *gorm.DB) *bookpublic.Handler {
	bookQuery := pgbook.NewQuery(db)
	return &bookpublic.Handler{
		ReadBookUsecase:    &bookqry.ReadBookUsecaseImpl{Books: bookQuery},
		SearchBooksUsecase: &bookqry.SearchBooksUsecaseImpl{Books: bookQuery},
	}
}

// integrationCatalogEntry は Fake カタログが返す書誌・書影を組み立てる。
func integrationCatalogEntry(t *testing.T, isbn string) *bookdomain.CatalogEntry {
	t.Helper()
	parsed, err := bookdomain.NewISBN(isbn)
	if err != nil {
		t.Fatal(err)
	}
	bib, err := bookdomain.NewBibliography("結合テストの書名", "結合テスト著者", "結合テスト出版社", "202609")
	if err != nil {
		t.Fatal(err)
	}
	return &bookdomain.CatalogEntry{ISBN: parsed, Bibliography: bib, Cover: nil}
}

// mustCreateIntegrationBook は Repository へ直接（外部カタログを経由せず）本を1冊作る。
// 事前データの用意が目的で、登録APIそのもののテストではないため。
func mustCreateIntegrationBook(t *testing.T, db *gorm.DB, isbn string) *bookdomain.Book {
	t.Helper()
	entry := integrationCatalogEntry(t, isbn)
	summary, err := bookdomain.NewSummary("結合テスト用の一言まとめ")
	if err != nil {
		t.Fatal(err)
	}
	comment, err := bookdomain.NewComment("結合テスト用の感想")
	if err != nil {
		t.Fatal(err)
	}
	rating, err := bookdomain.NewRating(4)
	if err != nil {
		t.Fatal(err)
	}
	tags, err := bookdomain.NewTagSelection(nil)
	if err != nil {
		t.Fatal(err)
	}
	b := bookdomain.New(entry.ISBN, entry.Bibliography, entry.Cover, summary, comment, rating, tags)
	if err := pgbook.NewRepository(db).Create(context.Background(), b); err != nil {
		t.Fatalf("create book: %v", err)
	}
	t.Cleanup(func() { db.Exec("DELETE FROM book WHERE id = ?", int64(b.ID)) })
	return b
}

func itoa(v int64) string { return strconv.FormatInt(v, 10) }

// integrationISBNSeq は nextIntegrationISBN の一意性を同一プロセス内の複数呼び出しでも保証するための連番。
var integrationISBNSeq atomic.Int64

// nextIntegrationISBN は結合テストごとに一意な、チェックディジットが正しい ISBN-13 を生成する。
// 固定リテラルだと、前回実行のクリーンアップ漏れ（プロセスの強制終了など）で同じ ISBN の行が
// 開発用 DB に残っていた場合、Repository がユニーク制約違反を ErrConflict に変換してしまい、
// Handler の検証（本テストの目的）に届く前にテストが落ちてしまうため。
func nextIntegrationISBN(t *testing.T) string {
	t.Helper()
	n := time.Now().UnixNano() + integrationISBNSeq.Add(1)
	first12 := fmt.Sprintf("978%09d", n%1_000_000_000)
	sum := 0
	for i, r := range first12 {
		d := int(r - '0')
		if i%2 == 1 {
			d *= 3
		}
		sum += d
	}
	return first12 + strconv.Itoa((10-sum%10)%10)
}

func TestBookHandlerIntegration_List(t *testing.T) {
	db := connectIntegrationDB(t)
	mustCreateIntegrationBook(t, db, nextIntegrationISBN(t))
	h := newIntegrationHandler(db)

	t.Run("200: ハッピーパス", func(t *testing.T) {
		rec := serve(t, h, http.MethodGet, "/api/books", "", false)
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200 (body=%s)", rec.Code, rec.Body.String())
		}
		got := decode[bookpublic.ListResponse](t, rec)
		if got.Total < 1 || len(got.Items) < 1 {
			t.Fatalf("body = %+v, want at least 1 item", got)
		}
	})

	t.Run("400: qが不正", func(t *testing.T) {
		rec := serve(t, h, http.MethodGet, "/api/books?tagId=0", "", false)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want 400 (body=%s)", rec.Code, rec.Body.String())
		}
	})
}

func TestBookHandlerIntegration_Get(t *testing.T) {
	db := connectIntegrationDB(t)
	isbn := nextIntegrationISBN(t)
	b := mustCreateIntegrationBook(t, db, isbn)
	h := newIntegrationHandler(db)

	t.Run("200: ハッピーパス", func(t *testing.T) {
		rec := serve(t, h, http.MethodGet, "/api/books/"+itoa(int64(b.ID)), "", false)
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200 (body=%s)", rec.Code, rec.Body.String())
		}
		got := decode[bookpublic.Response](t, rec)
		if got.ISBN != isbn {
			t.Fatalf("isbn = %q, want %s", got.ISBN, isbn)
		}
	})

	t.Run("404: 存在しないID", func(t *testing.T) {
		rec := serve(t, h, http.MethodGet, "/api/books/999999999", "", false)
		if rec.Code != http.StatusNotFound {
			t.Fatalf("status = %d, want 404 (body=%s)", rec.Code, rec.Body.String())
		}
	})
}
