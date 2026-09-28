package common_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"

	"github.com/mrstsgk/book-management-system/backend/internal/presentation/http/common"
)

func TestRequireAdminToken(t *testing.T) {
	silenceLog(t)
	tests := []struct {
		name       string
		configured string
		header     string
		want       int
	}{
		{name: "一致するトークンなら通す", configured: "s3cret", header: "Bearer s3cret", want: http.StatusOK},
		{name: "違うトークンは401", configured: "s3cret", header: "Bearer wrong", want: http.StatusUnauthorized},
		{name: "前方一致だけのトークンは401", configured: "s3cret", header: "Bearer s3cre", want: http.StatusUnauthorized},
		{name: "ヘッダが無ければ401", configured: "s3cret", header: "", want: http.StatusUnauthorized},
		{name: "Bearer以外の形式は401", configured: "s3cret", header: "Basic s3cret", want: http.StatusUnauthorized},
		{name: "設定が空なら空のトークンでも401", configured: "", header: "Bearer ", want: http.StatusUnauthorized},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := common.NewEcho()
			called := false
			e.POST("/w", func(c echo.Context) error {
				called = true
				return c.NoContent(http.StatusOK)
			}, common.RequireAdminToken(tt.configured))
			req := httptest.NewRequest(http.MethodPost, "/w", nil)
			if tt.header != "" {
				req.Header.Set(echo.HeaderAuthorization, tt.header)
			}
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			if rec.Code != tt.want {
				t.Fatalf("status = %d, want %d", rec.Code, tt.want)
			}
			if called != (tt.want == http.StatusOK) {
				t.Fatalf("handler called = %v, want %v", called, tt.want == http.StatusOK)
			}
		})
	}
}
