package public_test

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"gorm.io/gorm"

	webapi "github.com/mrstsgk/book-management-system/backend/api/gen/web"
	"github.com/mrstsgk/book-management-system/backend/internal/apitest"
	bookdomain "github.com/mrstsgk/book-management-system/backend/internal/book/domain"
	pgbook "github.com/mrstsgk/book-management-system/backend/internal/book/infrastructure/postgres"
	bookpublic "github.com/mrstsgk/book-management-system/backend/internal/book/presentation/http/public"
	bookqry "github.com/mrstsgk/book-management-system/backend/internal/book/usecase/query"
	pgplatform "github.com/mrstsgk/book-management-system/backend/internal/platform/postgres"
)

// connectIntegrationDB はローカルの開発用 DB に接続する。実際の PostgreSQL に対する結合テストなので、
// DB に繋がらなければ skip せず失敗にする（docs/rules/testing.md）。
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

type testServer struct {
	apitest.Unimplemented
	*bookpublic.Handler
}

// newServer は、本物の Query/UseCase を実 DB に配線した閲覧の Handler をサーバーに載せる。
func newServer(t *testing.T, db *gorm.DB) *apitest.Server {
	t.Helper()
	bookQuery := pgbook.NewQuery(db)
	return apitest.NewServer(t, &testServer{Handler: &bookpublic.Handler{
		ReadBookUsecase:    &bookqry.ReadBookUsecaseImpl{Books: bookQuery},
		SearchBooksUsecase: &bookqry.SearchBooksUsecaseImpl{Books: bookQuery},
	}}, apitest.AcceptSession{})
}

// mustCreateBook は Repository へ直接（外部カタログを経由せず）本を 1 冊作る。閲覧の API の準備のため。
func mustCreateBook(t *testing.T, db *gorm.DB, isbn string) *bookdomain.Book {
	t.Helper()
	parsed, err := bookdomain.NewISBN(isbn)
	if err != nil {
		t.Fatal(err)
	}
	bib, _ := bookdomain.NewBibliography("結合テストの書名", "結合テスト著者", "結合テスト出版社", "202609")
	summary, _ := bookdomain.NewSummary("結合テスト用の一言まとめ")
	comment, _ := bookdomain.NewComment("結合テスト用の感想")
	rating, _ := bookdomain.NewRating(4)
	tags, _ := bookdomain.NewTagSelection(nil)
	b := bookdomain.New(parsed, bib, nil, summary, comment, rating, tags)
	if err := pgbook.NewRepository(db).Create(context.Background(), b); err != nil {
		t.Fatalf("create book: %v", err)
	}
	t.Cleanup(func() { db.Exec("DELETE FROM book WHERE id = ?", int64(b.ID)) })
	return b
}

var isbnSeq atomic.Int64

// nextISBN は、テストごとに一意でチェックディジットが正しい ISBN-13 を作る（前回の片付け漏れと衝突しないため）。
func nextISBN(t *testing.T) string {
	t.Helper()
	n := time.Now().UnixNano() + isbnSeq.Add(1)
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

func TestSearchBooks(t *testing.T) {
	db := connectIntegrationDB(t)
	isbn := nextISBN(t)
	b := mustCreateBook(t, db, isbn)
	s := newServer(t, db)

	t.Run("200: 条件に合う本を一覧の項目で返す", func(t *testing.T) {
		rec := s.Do(t, apitest.Request{Method: http.MethodGet, Path: "/api/books?q=" + isbn + "&limit=5"})
		apitest.AssertStatus(t, rec, http.StatusOK)
		got := apitest.Decode[webapi.SearchBooksResponse](t, rec)
		if got.Limit != 5 || got.Offset != 0 {
			t.Fatalf("limit/offset = %d/%d, want 5/0", got.Limit, got.Offset)
		}
		var found *webapi.BookListItemResponse
		for i := range got.Items {
			if got.Items[i].ID == int64(b.ID) {
				found = &got.Items[i]
			}
		}
		if found != nil {
			t.Fatalf("q is matched against title and authors, not ISBN: found %+v", found)
		}
	})

	t.Run("200: 書名で絞り込むと全項目が載る", func(t *testing.T) {
		rec := s.Do(t, apitest.Request{Method: http.MethodGet, Path: "/api/books?q=" + "結合テストの書名" + "&limit=100"})
		apitest.AssertStatus(t, rec, http.StatusOK)
		got := apitest.Decode[webapi.SearchBooksResponse](t, rec)
		for _, it := range got.Items {
			if it.ID == int64(b.ID) {
				if it.Isbn != isbn || it.Title != "結合テストの書名" || it.Authors != "結合テスト著者" || it.Summary != "結合テスト用の一言まとめ" ||
					it.Rating != 4 || len(it.Tags) != 0 || it.CoverURL != nil || it.AmazonURL == nil {
					t.Fatalf("item = %+v", it)
				}
				return
			}
		}
		t.Fatalf("book %d not in %+v", b.ID, got.Items)
	})

	t.Run("400: キーワードが 101 文字なら errors の無い 400（VO の判定）", func(t *testing.T) {
		rec := s.Do(t, apitest.Request{Method: http.MethodGet, Path: "/api/books?q=" + strings.Repeat("a", 101)})
		apitest.AssertStatus(t, rec, http.StatusBadRequest)
	})

	t.Run("500: DB の呼び出しが失敗", func(t *testing.T) {
		rec := s.Do(t, apitest.Request{Method: http.MethodGet, Path: "/api/books", Cancelled: true})
		apitest.AssertStatus(t, rec, http.StatusInternalServerError)
	})
}

// TestSearchBooks_Constraints は、YAML に書いた query の制約 1 つにつき、通る値と弾かれる値を 1 組ずつ確かめる。
func TestSearchBooks_Constraints(t *testing.T) {
	db := connectIntegrationDB(t)
	s := newServer(t, db)
	tests := []struct {
		name, query, field, rule string
		pass                     bool
	}{
		{"limit 1 は通る", "limit=1", "limit", "minimum", true},
		{"limit 0 は弾く", "limit=0", "limit", "minimum", false},
		{"limit 100 は通る", "limit=100", "limit", "maximum", true},
		{"limit 101 は弾く", "limit=101", "limit", "maximum", false},
		{"limit 数字以外は弾く", "limit=abc", "limit", "type", false},
		{"offset 0 は通る", "offset=0", "offset", "minimum", true},
		{"offset -1 は弾く", "offset=-1", "offset", "minimum", false},
		{"tagId 1 は通る", "tagId=1", "tagId", "minimum", true},
		{"tagId 0 は弾く", "tagId=0", "tagId", "minimum", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := s.Do(t, apitest.Request{Method: http.MethodGet, Path: "/api/books?" + tt.query})
			if tt.pass {
				apitest.AssertStatus(t, rec, http.StatusOK)
				return
			}
			apitest.AssertFieldErrors(t, rec, apitest.FieldError(tt.field, tt.rule))
		})
	}
}

func TestReadBook(t *testing.T) {
	db := connectIntegrationDB(t)
	isbn := nextISBN(t)
	b := mustCreateBook(t, db, isbn)
	s := newServer(t, db)

	t.Run("200: 詳細の全項目を返す", func(t *testing.T) {
		rec := s.Do(t, apitest.Request{Method: http.MethodGet, Path: "/api/books/" + strconv.FormatInt(int64(b.ID), 10)})
		apitest.AssertStatus(t, rec, http.StatusOK)
		got := apitest.Decode[webapi.BookResponse](t, rec)
		if got.Isbn != isbn || got.Title != "結合テストの書名" || got.Authors != "結合テスト著者" || got.Publisher != "結合テスト出版社" ||
			got.PublishedOn != "202609" || got.Summary != "結合テスト用の一言まとめ" || got.Comment != "結合テスト用の感想" || got.Rating != 4 ||
			got.TitleOverride != nil || len(got.Tags) != 0 || got.Version != 1 || got.CoverURL != nil || got.CoverSource != nil || got.AmazonURL == nil {
			t.Fatalf("response = %+v", got)
		}
	})

	t.Run("400: id が 0 なら minimum", func(t *testing.T) {
		rec := s.Do(t, apitest.Request{Method: http.MethodGet, Path: "/api/books/0"})
		apitest.AssertFieldErrors(t, rec, apitest.FieldError("id", "minimum"))
	})

	t.Run("400: id が数字以外なら type", func(t *testing.T) {
		rec := s.Do(t, apitest.Request{Method: http.MethodGet, Path: "/api/books/abc"})
		apitest.AssertFieldErrors(t, rec, apitest.FieldError("id", "type"))
	})

	t.Run("404: 存在しない ID は bookNotFound", func(t *testing.T) {
		rec := s.Do(t, apitest.Request{Method: http.MethodGet, Path: "/api/books/999999999"})
		apitest.AssertCode(t, rec, http.StatusNotFound, "bookNotFound")
	})

	t.Run("500: DB の呼び出しが失敗", func(t *testing.T) {
		rec := s.Do(t, apitest.Request{Method: http.MethodGet, Path: "/api/books/1", Cancelled: true})
		apitest.AssertStatus(t, rec, http.StatusInternalServerError)
	})
}
