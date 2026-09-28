package main

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"

	"github.com/mrstsgk/book-management-system/backend/config"
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

func TestNewCatalog_WorksWithAndWithoutRakutenKeys(t *testing.T) {
	for _, cfg := range []config.CatalogConfig{
		{OpenBDBaseURL: "http://openbd.test"},
		{OpenBDBaseURL: "http://openbd.test", RakutenBaseURL: "http://rakuten.test", RakutenApplicationID: "app", RakutenAccessKey: "key"},
	} {
		if newCatalog(cfg) == nil {
			t.Fatalf("newCatalog(%+v) returned nil", cfg)
		}
	}
}
