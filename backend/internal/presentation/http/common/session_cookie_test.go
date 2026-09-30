package common_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"

	"github.com/mrstsgk/book-management-system/backend/internal/presentation/http/common"
)

func setCookieOf(t *testing.T, handler echo.HandlerFunc) *http.Cookie {
	t.Helper()
	silenceLog(t)
	e := common.NewEcho()
	e.POST("/x", handler)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/x", nil))
	cookies := rec.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("cookies = %d, want 1", len(cookies))
	}
	return cookies[0]
}

func TestSetSessionCookie(t *testing.T) {
	ck := setCookieOf(t, func(c echo.Context) error {
		common.SetSessionCookie(c, "sid")
		return c.NoContent(http.StatusNoContent)
	})
	if ck.Name != common.SessionCookieName || ck.Value != "sid" {
		t.Fatalf("cookie = %s=%s", ck.Name, ck.Value)
	}
	if !ck.HttpOnly || ck.SameSite != http.SameSiteLaxMode || ck.Path != "/api" || ck.MaxAge != 86400 {
		t.Fatalf("attributes = %+v, want HttpOnly, Lax, Path=/api, Max-Age=86400", ck)
	}
}

func TestClearSessionCookie(t *testing.T) {
	ck := setCookieOf(t, func(c echo.Context) error {
		common.ClearSessionCookie(c)
		return c.NoContent(http.StatusNoContent)
	})
	if ck.Name != common.SessionCookieName || ck.Value != "" || ck.MaxAge != -1 {
		t.Fatalf("cookie = %+v, want an empty value with Max-Age=0", ck)
	}
	if !ck.HttpOnly || ck.Path != "/api" {
		t.Fatalf("clearing cookie must keep the same Path and HttpOnly: %+v", ck)
	}
}
