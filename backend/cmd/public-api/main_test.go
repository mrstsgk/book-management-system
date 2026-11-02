package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"

	publicapi "github.com/mrstsgk/book-management-system/backend/api/gen/public"
	"github.com/mrstsgk/book-management-system/backend/config"
	bookdomain "github.com/mrstsgk/book-management-system/backend/internal/book/domain"
	pgbook "github.com/mrstsgk/book-management-system/backend/internal/book/infrastructure/postgres"
	"github.com/mrstsgk/book-management-system/backend/internal/platform/httpserver"
	pgplatform "github.com/mrstsgk/book-management-system/backend/internal/platform/postgres"
	tagdomain "github.com/mrstsgk/book-management-system/backend/internal/tag/domain"
)

// run・buildEcho は非公開なので、このテストは package main に置く。

func TestRun_FailsFastOnInvalidConfig(t *testing.T) {
	t.Setenv("STAGE", "local")
	t.Setenv("LOG_LEVEL", "verbose")

	if err := run(); err == nil {
		t.Fatal("expected an error for an invalid LOG_LEVEL")
	}
}

func TestRun_FailsWhenDBIsUnreachable(t *testing.T) {
	t.Setenv("STAGE", "local")
	t.Setenv("LOG_LEVEL", "error")
	t.Setenv("DB_HOST", "127.0.0.1")
	t.Setenv("DB_PORT", "1")

	if err := run(); err == nil {
		t.Fatal("expected an error when the DB is unreachable")
	}
}

// buildEcho に Fake を渡して、全ルートが Handler → UseCase を通って応答すること（依存の nil・登録漏れ）と、
// 認証・入力検証の middleware が spec から掛かっていることを確かめる。DB と外部カタログには繋がない。

type stubBookQuery struct{}

func (stubBookQuery) FindDetailByID(context.Context, bookdomain.ID) (*bookdomain.BookDetail, error) {
	return &bookdomain.BookDetail{ID: 1, ISBN: "9784873118703", Title: "書名", Authors: "著者", Publisher: "出版社", PublishedOn: "202601",
		Summary: "まとめ", Tags: []string{}, Comment: "感想", Rating: 4, Version: 1}, nil
}
func (stubBookQuery) Search(context.Context, bookdomain.ListCondition, bookdomain.ListRange) (*bookdomain.BookList, error) {
	return &bookdomain.BookList{Items: []*bookdomain.BookListItem{}}, nil
}
func (stubBookQuery) ExistsByISBN(context.Context, bookdomain.ISBN) (bool, error) { return false, nil }

type stubTagQuery struct{}

func (stubTagQuery) FindAll(context.Context) (*tagdomain.TagList, error) {
	return &tagdomain.TagList{Items: []*tagdomain.TagListItem{}}, nil
}
func (stubTagQuery) ExistsAll(context.Context, []tagdomain.ID) (bool, error) { return true, nil }
func (stubTagQuery) ExistsByName(context.Context, string, tagdomain.ID) (bool, error) {
	return false, nil
}
func (stubTagQuery) CountBooks(context.Context) (*tagdomain.TagBookCounts, error) {
	return &tagdomain.TagBookCounts{Items: []*tagdomain.TagBookCount{}}, nil
}

func newTestEcho(t *testing.T) *echo.Echo {
	t.Helper()
	spec, err := publicapi.GetSpec()
	if err != nil {
		t.Fatal(err)
	}
	e, err := buildEcho(spec, deps{BookQuery: stubBookQuery{}, TagQuery: stubTagQuery{}})
	if err != nil {
		t.Fatal(err)
	}
	return e
}

func send(e *echo.Echo, method, path, origin string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, nil)
	req.Header.Set("Origin", origin)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}

func TestBuildEcho_すべてのルートがHandlerとUseCaseを通って応答する(t *testing.T) {
	silenceLog(t)
	e := newTestEcho(t)
	for _, path := range []string{"/api/books", "/api/books/1", "/api/tags", "/api/tags/counts"} {
		t.Run(path, func(t *testing.T) {
			// 閲覧はセッションも Origin も要らない（別のサイトからの読み込みでも 200）
			rec := send(e, http.MethodGet, path, "http://evil.example")
			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d, want 200 (body=%s)", rec.Code, rec.Body.String())
			}
		})
	}
}

func TestBuildEcho_管理のAPIは公開しない(t *testing.T) {
	silenceLog(t)
	rec := send(newTestEcho(t), http.MethodPost, "/api/admin/books", "http://example.com")
	if rec.Code != http.StatusNotFound && rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want 404/405: public-api must not serve admin routes (body=%s)", rec.Code, rec.Body.String())
	}
}

func TestBuildEcho_specに違反するリクエストは入力検証で400になる(t *testing.T) {
	silenceLog(t)
	rec := send(newTestEcho(t), http.MethodGet, "/api/books?limit=0", "http://example.com")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400 (body=%s)", rec.Code, rec.Body.String())
	}
	var got httpserver.ErrorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if want := []httpserver.FieldError{{Field: "limit", Rule: "minimum"}}; !reflect.DeepEqual(got.Errors, want) {
		t.Fatalf("errors = %+v, want %+v", got.Errors, want)
	}
}

func silenceLog(t *testing.T) {
	t.Helper()
	orig := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, nil)))
	t.Cleanup(func() { slog.SetDefault(orig) })
}

func TestNewCatalog_ReturnsTheOpenBDCatalog(t *testing.T) {
	if newCatalog(config.CatalogConfig{OpenBDBaseURL: "http://openbd.test"}) == nil {
		t.Fatal("newCatalog returned nil")
	}
}

// migratedTempDB は共有の開発 DB を空にせずに「空の DB での起動」を試すため、一時的なデータベースを作って
// backend/migrations の up を順に流す。ローカルの PostgreSQL に繋がらなければ失敗する。
func migratedTempDB(t *testing.T) *gorm.DB {
	t.Helper()
	base := pgplatform.Config{Host: "localhost", Port: "5432", User: "postgres", Password: "postgres", DBName: "book_management", SSLMode: "disable"}
	admin, err := pgplatform.Connect(base)
	if err != nil {
		t.Fatalf("local Postgres not reachable (run `make db-up` first; DB を使うテストは skip しない。docs/rules/testing.md): %v", err)
	}
	t.Cleanup(func() { _ = pgplatform.Close(admin) })

	name := fmt.Sprintf("book_management_seedtest_%d", time.Now().UnixNano())
	if err := admin.Exec("CREATE DATABASE " + name).Error; err != nil {
		t.Fatalf("cannot create a temporary database: %v", err)
	}
	t.Cleanup(func() { admin.Exec("DROP DATABASE IF EXISTS " + name + " WITH (FORCE)") })

	cfg := base
	cfg.DBName = name
	db, err := pgplatform.Connect(cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = pgplatform.Close(db) })

	files, err := filepath.Glob(filepath.Join("..", "..", "migrations", "*.up.sql"))
	if err != nil || len(files) == 0 {
		t.Fatalf("no migrations found: %v", err)
	}
	sort.Strings(files)
	for _, f := range files {
		sql, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		if err := db.Exec(string(sql)).Error; err != nil {
			t.Fatalf("migrate %s: %v", filepath.Base(f), err)
		}
	}
	return db
}

func TestSeedIfEmpty_OnAnEmptyDatabaseSeedsOnceAcrossRestarts(t *testing.T) {
	db := migratedTempDB(t)
	ctx := context.Background()
	books, query := pgbook.NewRepository(db), pgbook.NewQuery(db)
	offline := &fakeSeedCatalog{err: fmt.Errorf("offline")}

	for start := 1; start <= 2; start++ {
		if err := seedIfEmpty(ctx, books, query, offline); err != nil {
			t.Fatalf("start %d: %v", start, err)
		}
		var count int64
		if err := db.Raw("SELECT COUNT(*) FROM book").Scan(&count).Error; err != nil {
			t.Fatal(err)
		}
		if count != int64(len(sampleBooks)) {
			t.Fatalf("start %d: %d books, want %d", start, count, len(sampleBooks))
		}
	}
}
