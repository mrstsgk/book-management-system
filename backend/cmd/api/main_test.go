package main

import (
	"net/http"
	"net/http/httptest"
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
		"POST /api/tags",
		"PUT /api/tags/:id",
		"DELETE /api/tags/:id",
	} {
		if !got[want] {
			t.Errorf("route %q is not registered (got %v)", want, got)
		}
	}
}

// TestRegisterRoutes_WiresTagQueryIntoBookUsecases は main.go が RegisterUsecaseImpl/UpdateUsecaseImpl の
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
