package auth_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"

	domaincommon "github.com/mrstsgk/book-management-system/backend/internal/domain/common"
	httpauth "github.com/mrstsgk/book-management-system/backend/internal/presentation/http/auth"
	"github.com/mrstsgk/book-management-system/backend/internal/presentation/http/common"
	authcmd "github.com/mrstsgk/book-management-system/backend/internal/usecase/auth/command"
)

type fakeLogin func(context.Context, authcmd.LoginCommand) (string, error)

func (f fakeLogin) Execute(ctx context.Context, cmd authcmd.LoginCommand) (string, error) {
	return f(ctx, cmd)
}

type fakeLogout func(context.Context, string) error

func (f fakeLogout) Execute(ctx context.Context, id string) error { return f(ctx, id) }

type fakeCheck func(context.Context, string) error

func (f fakeCheck) Execute(ctx context.Context, id string) error { return f(ctx, id) }

// serve は cmd/api/main.go と同じ形（NewEcho + Register）でハンドラを組み立ててリクエストを流す。
// Origin は自サイトを付ける（RequireSameOrigin を通すため。Origin 自体の分岐は common のテストが担う）。
func serve(t *testing.T, h *httpauth.Handler, method, path, body, cookie string) *httptest.ResponseRecorder {
	t.Helper()
	orig := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, nil)))
	t.Cleanup(func() { slog.SetDefault(orig) })

	h.SameOrigin = common.RequireSameOrigin()
	e := common.NewEcho()
	h.Register(e.Group("/api/auth"))
	req := httptest.NewRequest(method, "http://localhost:3000"+path, strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set("Origin", "http://localhost:3000")
	if cookie != "" {
		req.AddCookie(&http.Cookie{Name: common.SessionCookieName, Value: cookie})
	}
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}

func sessionCookie(t *testing.T, rec *httptest.ResponseRecorder) *http.Cookie {
	t.Helper()
	for _, ck := range rec.Result().Cookies() {
		if ck.Name == common.SessionCookieName {
			return ck
		}
	}
	t.Fatalf("no %s cookie in %v", common.SessionCookieName, rec.Header())
	return nil
}

func TestHandlerLogin(t *testing.T) {
	t.Run("合っていれば204でhttpOnlyのCookieを返す", func(t *testing.T) {
		var got authcmd.LoginCommand
		h := &httpauth.Handler{LoginUC: fakeLogin(func(_ context.Context, cmd authcmd.LoginCommand) (string, error) {
			got = cmd
			return strings.Repeat("a", 43), nil
		})}
		rec := serve(t, h, http.MethodPost, "/api/auth/login", `{"id":"admin","password":"pw"}`, "")
		if rec.Code != http.StatusNoContent {
			t.Fatalf("status = %d, want 204 (body=%s)", rec.Code, rec.Body.String())
		}
		if got != (authcmd.LoginCommand{ID: "admin", Password: "pw"}) {
			t.Fatalf("usecase received %+v", got)
		}
		ck := sessionCookie(t, rec)
		if len(ck.Value) != 43 || !ck.HttpOnly || ck.SameSite != http.SameSiteLaxMode || ck.Path != "/api" {
			t.Fatalf("cookie = %+v", ck)
		}
	})

	t.Run("Originが別サイトなら403でUseCaseを呼ばない", func(t *testing.T) {
		h := &httpauth.Handler{LoginUC: fakeLogin(func(context.Context, authcmd.LoginCommand) (string, error) {
			t.Error("usecase must not be called")
			return "", nil
		})}
		h.SameOrigin = common.RequireSameOrigin()
		e := common.NewEcho()
		h.Register(e.Group("/api/auth"))
		req := httptest.NewRequest(http.MethodPost, "http://localhost:3000/api/auth/login", strings.NewReader(`{"id":"a","password":"b"}`))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set("Origin", "http://evil.test")
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		if rec.Code != http.StatusForbidden {
			t.Fatalf("status = %d, want 403", rec.Code)
		}
	})

	for _, tt := range []struct {
		name string
		body string
		err  error
		want int
	}{
		{name: "空欄は400でUseCaseを呼ばない", body: `{"id":"","password":""}`, want: http.StatusBadRequest},
		{name: "JSONでなければ400", body: `not json`, want: http.StatusBadRequest},
		{name: "不一致は401", body: `{"id":"admin","password":"bad"}`, err: domaincommon.ErrUnauthorized, want: http.StatusUnauthorized},
		{name: "ロック中は429", body: `{"id":"admin","password":"bad"}`, err: domaincommon.ErrTooManyAttempts, want: http.StatusTooManyRequests},
		{name: "UseCaseの障害は500", body: `{"id":"admin","password":"pw"}`, err: errors.New("db down"), want: http.StatusInternalServerError},
	} {
		t.Run(tt.name, func(t *testing.T) {
			h := &httpauth.Handler{LoginUC: fakeLogin(func(context.Context, authcmd.LoginCommand) (string, error) {
				if tt.err == nil {
					t.Error("usecase must not be called")
				}
				return "", tt.err
			})}
			rec := serve(t, h, http.MethodPost, "/api/auth/login", tt.body, "")
			if rec.Code != tt.want {
				t.Fatalf("status = %d, want %d (body=%s)", rec.Code, tt.want, rec.Body.String())
			}
			for _, ck := range rec.Result().Cookies() {
				if ck.Name == common.SessionCookieName {
					t.Fatal("must not set a session cookie on failure")
				}
			}
		})
	}
}

func TestHandlerLogout(t *testing.T) {
	t.Run("Cookieがあれば削除して204、Cookieを消す", func(t *testing.T) {
		var gotID string
		h := &httpauth.Handler{LogoutUC: fakeLogout(func(_ context.Context, id string) error { gotID = id; return nil })}
		rec := serve(t, h, http.MethodPost, "/api/auth/logout", "", "sid")
		if rec.Code != http.StatusNoContent || gotID != "sid" {
			t.Fatalf("status = %d id = %q", rec.Code, gotID)
		}
		if ck := sessionCookie(t, rec); ck.Value != "" || ck.MaxAge != -1 {
			t.Fatalf("cookie = %+v, want cleared", ck)
		}
	})

	t.Run("Cookieが無くても204でUseCaseを呼ばない", func(t *testing.T) {
		h := &httpauth.Handler{LogoutUC: fakeLogout(func(context.Context, string) error {
			t.Error("usecase must not be called")
			return nil
		})}
		if rec := serve(t, h, http.MethodPost, "/api/auth/logout", "", ""); rec.Code != http.StatusNoContent {
			t.Fatalf("status = %d, want 204", rec.Code)
		}
	})

	t.Run("UseCaseの障害は500", func(t *testing.T) {
		h := &httpauth.Handler{LogoutUC: fakeLogout(func(context.Context, string) error { return errors.New("db down") })}
		if rec := serve(t, h, http.MethodPost, "/api/auth/logout", "", "sid"); rec.Code != http.StatusInternalServerError {
			t.Fatalf("status = %d, want 500", rec.Code)
		}
	})
}

func TestHandlerSession(t *testing.T) {
	check := fakeCheck(func(_ context.Context, id string) error {
		if id == "good" {
			return nil
		}
		return domaincommon.ErrUnauthorized
	})
	for _, tt := range []struct {
		name   string
		cookie string
		want   int
	}{
		{name: "有効なら204", cookie: "good", want: http.StatusNoContent},
		{name: "無効なら401", cookie: "bad", want: http.StatusUnauthorized},
		{name: "Cookieが無ければ401", cookie: "", want: http.StatusUnauthorized},
	} {
		t.Run(tt.name, func(t *testing.T) {
			h := &httpauth.Handler{CheckUC: check}
			if rec := serve(t, h, http.MethodGet, "/api/auth/session", "", tt.cookie); rec.Code != tt.want {
				t.Fatalf("status = %d, want %d", rec.Code, tt.want)
			}
		})
	}
}
