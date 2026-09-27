package common_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"

	"github.com/mrstsgk/book-management-system/backend/internal/presentation/http/common"
)

func TestNewEcho_HealthReturnsOK(t *testing.T) {
	silenceLog(t)
	e := common.NewEcho()
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/health", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid JSON body %q: %v", rec.Body.String(), err)
	}
	if body["status"] != "ok" {
		t.Errorf("status field = %q, want ok", body["status"])
	}
	if rec.Header().Get(echo.HeaderXRequestID) == "" {
		t.Error("X-Request-Id header is not set")
	}
}

func TestNewEcho_PanicIsRecoveredAs500(t *testing.T) {
	silenceLog(t)
	e := common.NewEcho()
	e.GET("/panic", func(echo.Context) error { panic("secret-internal-state") })
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/panic", nil))

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", rec.Code)
	}
	var got common.ErrorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("invalid JSON body %q: %v", rec.Body.String(), err)
	}
	if got.Message != "internal" {
		t.Errorf("message = %q, want internal", got.Message)
	}
}
