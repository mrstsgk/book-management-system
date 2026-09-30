package common_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"

	"github.com/mrstsgk/book-management-system/backend/internal/presentation/http/common"
)

func TestRequireSameOrigin(t *testing.T) {
	silenceLog(t)
	tests := []struct {
		name   string
		method string
		origin string
		want   int
	}{
		{name: "Originが自サイトなら通す", method: http.MethodPost, origin: "http://localhost:3000", want: http.StatusOK},
		{name: "Originが別サイトは403", method: http.MethodPost, origin: "http://evil.test", want: http.StatusForbidden},
		{name: "Originがnullは403", method: http.MethodPost, origin: "null", want: http.StatusForbidden},
		{name: "Originが無ければ403", method: http.MethodPost, want: http.StatusForbidden},
		{name: "ポート違いは403", method: http.MethodPost, origin: "http://localhost:8080", want: http.StatusForbidden},
		{name: "GETには掛からない", method: http.MethodGet, origin: "http://evil.test", want: http.StatusOK},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := common.NewEcho()
			handler := func(c echo.Context) error { return c.NoContent(http.StatusOK) }
			e.POST("/w", handler, common.RequireSameOrigin())
			e.GET("/w", handler, common.RequireSameOrigin())
			req := httptest.NewRequest(tt.method, "http://localhost:3000/w", nil)
			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)
			if rec.Code != tt.want {
				t.Fatalf("status = %d, want %d", rec.Code, tt.want)
			}
		})
	}
}
