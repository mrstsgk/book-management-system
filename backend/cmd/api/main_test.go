package main

import (
	"testing"

	"github.com/labstack/echo/v4"

	"github.com/mrstsgk/book-management-system/backend/config"
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
	} {
		if !got[want] {
			t.Errorf("route %q is not registered (got %v)", want, got)
		}
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
