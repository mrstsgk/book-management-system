package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"

	"github.com/mrstsgk/book-management-system/backend/config"
	domainbook "github.com/mrstsgk/book-management-system/backend/internal/domain/book"
	pgbook "github.com/mrstsgk/book-management-system/backend/internal/infrastructure/postgres/book"
	pgcommon "github.com/mrstsgk/book-management-system/backend/internal/infrastructure/postgres/common"
	httpcommon "github.com/mrstsgk/book-management-system/backend/internal/presentation/http/common"
)

// run は非公開なので、このテストは package main に置く。

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

func TestRegisterRoutes_ExposesBookAndCatalogAPI(t *testing.T) {
	e := echo.New()
	registerRoutes(e, nil, nil, "token")

	got := map[string]bool{}
	for _, r := range e.Routes() {
		got[r.Method+" "+r.Path] = true
	}
	for _, want := range []string{
		"GET /api/books",
		"GET /api/books/:id",
		"POST /api/books",
		"PUT /api/books/:id",
		"DELETE /api/books/:id",
		"GET /api/catalog/:isbn",
		"GET /api/tags",
		"GET /api/tags/counts",
		"POST /api/tags",
		"PUT /api/tags/:id",
		"DELETE /api/tags/:id",
	} {
		if !got[want] {
			t.Errorf("route %q is not registered (got %v)", want, got)
		}
	}
	// 楽天は撤去した（書影は openBD から取る）ので、楽天由来の情報を消す API も残さない
	if got["DELETE /api/books/:id/rakuten"] {
		t.Error("the Rakuten removal route must no longer be registered")
	}
}

// TestRegisterRoutes_WiresTagQueryIntoBookUsecases は main.go が RegisterUsecaseImpl と UpdateUsecaseImpl の両方の
// Tags に本物の tag.Query を渡していることを、実際のリクエストで確かめる（配線漏れなら nil interface の
// メソッド呼び出しでパニックする）。実DBが要るため繋がらなければ skip する。
func TestRegisterRoutes_WiresTagQueryIntoBookUsecases(t *testing.T) {
	db, err := pgcommon.Connect(pgcommon.Config{
		Host: "localhost", Port: "5432", User: "postgres", Password: "postgres",
		DBName: "book_management", SSLMode: "disable",
	})
	if err != nil {
		t.Skipf("skipping: local Postgres not reachable (run `make db-up migrate-up` first): %v", err)
	}

	e := httpcommon.NewEcho()
	registerRoutes(e, db, nil, "token")

	body := `{"isbn":"4873118700","summary":"要約","tagIds":[999999999],"comment":"良書","rating":5}`
	req := httptest.NewRequest(http.MethodPost, "/api/books", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(echo.HeaderAuthorization, "Bearer token")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	// 存在しないタグIDは400（実在確認までTagsの配線が届いていないとここでpanicする）。
	// タグ検証はカタログ問い合わせより前に行われるため、bookCatalog（nil）には到達しない。
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400 (body=%s)", rec.Code, rec.Body.String())
	}

	// 更新経路（UpdateUsecaseImpl.Tags）も同様に配線されていることを確かめる。
	// 更新はまず対象の本を読むので、実在する本の行を1件作っておく。
	var bookID int64
	if err := db.Raw(
		"INSERT INTO book (isbn, title, summary, comment, rating, version) VALUES (?, ?, ?, ?, ?, ?) RETURNING id",
		"9780000099990", "main-test-書名", "main-test-まとめ", "main-test-感想", 5, 1,
	).Scan(&bookID).Error; err != nil {
		t.Fatalf("insert book: %v", err)
	}
	t.Cleanup(func() { db.Exec("DELETE FROM book WHERE id = ?", bookID) })

	updateBody := `{"summary":"main-test-まとめ","tagIds":[999999999],"comment":"main-test-感想","rating":4,"version":1}`
	updateReq := httptest.NewRequest(http.MethodPut, "/api/books/"+strconv.FormatInt(bookID, 10), strings.NewReader(updateBody))
	updateReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	updateReq.Header.Set(echo.HeaderAuthorization, "Bearer token")
	updateRec := httptest.NewRecorder()
	e.ServeHTTP(updateRec, updateReq)

	if updateRec.Code != http.StatusBadRequest {
		t.Fatalf("PUT status = %d, want 400 (body=%s)", updateRec.Code, updateRec.Body.String())
	}
}

func TestNewCatalog(t *testing.T) {
	// openBD と Google Books の両方の役を1つの TLS サーバーで受ける（Google Books は https でないと問い合わせないため）
	var googleCalls atomic.Int32
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/get":
			_, _ = w.Write([]byte(`[{"summary":{"title":"配線テストの本","cover":"https://cover.openbd.jp/1.jpg"}}]`))
		case "/books/v1/volumes":
			googleCalls.Add(1)
			_, _ = w.Write([]byte(`{"totalItems":1,"items":[{"volumeInfo":{"imageLinks":{"thumbnail":"http://books.google.com/books/content?id=w"},"infoLink":"http://books.google.co.jp/books?id=w"}}]}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()
	isbn, err := domainbook.NewISBN("9784873118703")
	if err != nil {
		t.Fatal(err)
	}

	t.Run("Google BooksのキーがあればGoogle Booksの書影を使う", func(t *testing.T) {
		googleCalls.Store(0)
		cfg := config.CatalogConfig{OpenBDBaseURL: srv.URL, GoogleBooksBaseURL: srv.URL, GoogleBooksAPIKey: "test-key"}
		entry, err := newCatalog(cfg, srv.Client()).Lookup(context.Background(), isbn)
		if err != nil {
			t.Fatal(err)
		}
		if entry.Cover == nil || entry.Cover.Source() != domainbook.CoverSourceGoogleBooks || googleCalls.Load() != 1 {
			t.Fatalf("cover = %+v (Google Books calls: %d), want the Google Books cover", entry.Cover, googleCalls.Load())
		}
	})

	t.Run("キーが無ければGoogle Booksに問い合わせずopenBDの書影を使う", func(t *testing.T) {
		googleCalls.Store(0)
		cfg := config.CatalogConfig{OpenBDBaseURL: srv.URL, GoogleBooksBaseURL: srv.URL}
		entry, err := newCatalog(cfg, srv.Client()).Lookup(context.Background(), isbn)
		if err != nil {
			t.Fatal(err)
		}
		if entry.Cover == nil || entry.Cover.Source() != domainbook.CoverSourceOpenBD || googleCalls.Load() != 0 {
			t.Fatalf("cover = %+v (Google Books calls: %d), want the openBD cover without asking Google Books", entry.Cover, googleCalls.Load())
		}
	})
}

// blockingFill は Execute が release まで返らない FillMissingCoversUsecase の Fake。
type blockingFill struct {
	started chan struct{}
	release chan struct{}
}

func (f *blockingFill) Execute(context.Context) error {
	close(f.started)
	<-f.release
	return nil
}

func TestStartFillMissingCovers(t *testing.T) {
	uc := &blockingFill{started: make(chan struct{}), release: make(chan struct{})}
	returned := make(chan (<-chan struct{}))
	go func() { returned <- startFillMissingCovers(context.Background(), uc) }()

	var done <-chan struct{}
	select {
	case done = <-returned:
	case <-time.After(time.Second):
		close(uc.release)
		t.Fatal("startFillMissingCovers blocked; the server must not wait for the external catalogs")
	}
	select {
	case <-uc.started:
	case <-time.After(time.Second):
		t.Fatal("filling missing covers was not started")
	}
	select {
	case <-done:
		t.Fatal("done closed while filling is still running; the DB could be closed under it")
	default:
	}
	close(uc.release)
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("done was not closed after filling finished")
	}
}

// migratedTempDB は共有の開発 DB を空にせずに「空の DB での起動」を試すため、一時的なデータベースを作って
// backend/migrations の up を順に流す。ローカルの PostgreSQL に繋がらなければ skip する。
func migratedTempDB(t *testing.T) *gorm.DB {
	t.Helper()
	base := pgcommon.Config{Host: "localhost", Port: "5432", User: "postgres", Password: "postgres", DBName: "book_management", SSLMode: "disable"}
	admin, err := pgcommon.Connect(base)
	if err != nil {
		t.Skipf("skipping: local Postgres not reachable (run `make db-up` first): %v", err)
	}
	t.Cleanup(func() { _ = pgcommon.Close(admin) })

	name := fmt.Sprintf("book_management_seedtest_%d", time.Now().UnixNano())
	if err := admin.Exec("CREATE DATABASE " + name).Error; err != nil {
		t.Skipf("skipping: cannot create a temporary database: %v", err)
	}
	t.Cleanup(func() { admin.Exec("DROP DATABASE IF EXISTS " + name + " WITH (FORCE)") })

	cfg := base
	cfg.DBName = name
	db, err := pgcommon.Connect(cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = pgcommon.Close(db) })

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
