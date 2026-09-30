package common_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"

	domaincommon "github.com/mrstsgk/book-management-system/backend/internal/domain/common"
	"github.com/mrstsgk/book-management-system/backend/internal/presentation/http/common"
)

type fakeCheck func(context.Context, string) error

func (f fakeCheck) Execute(ctx context.Context, id string) error { return f(ctx, id) }

func TestRequireAdminSession(t *testing.T) {
	silenceLog(t)
	valid := fakeCheck(func(_ context.Context, id string) error {
		if id == "good" {
			return nil
		}
		return domaincommon.ErrUnauthorized
	})
	tests := []struct {
		name    string
		cookies []string
		check   fakeCheck
		want    int
	}{
		{name: "有効なセッションなら通す", cookies: []string{"good"}, check: valid, want: http.StatusOK},
		{name: "無効なセッションは401", cookies: []string{"bad"}, check: valid, want: http.StatusUnauthorized},
		{name: "Cookieが無ければ401でUseCaseを呼ばない", cookies: nil, check: fakeCheck(func(context.Context, string) error {
			t.Error("must not be called")
			return nil
		}), want: http.StatusUnauthorized},
		{name: "値が空なら401", cookies: []string{""}, check: valid, want: http.StatusUnauthorized},
		{name: "UseCaseの他のエラーは500", cookies: []string{"good"}, check: fakeCheck(func(context.Context, string) error {
			return errors.New("db down")
		}), want: http.StatusInternalServerError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := common.NewEcho()
			called := false
			e.POST("/w", func(c echo.Context) error {
				called = true
				return c.NoContent(http.StatusOK)
			}, common.RequireAdminSession(tt.check))
			req := httptest.NewRequest(http.MethodPost, "/w", nil)
			for _, v := range tt.cookies {
				req.AddCookie(&http.Cookie{Name: common.SessionCookieName, Value: v})
			}
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			if rec.Code != tt.want {
				t.Fatalf("status = %d, want %d", rec.Code, tt.want)
			}
			if called != (tt.want == http.StatusOK) {
				t.Fatalf("handler called = %v", called)
			}
		})
	}
}
