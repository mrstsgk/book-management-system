package admin_test

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"

	"github.com/mrstsgk/book-management-system/backend/internal/domainerr"
	"github.com/mrstsgk/book-management-system/backend/internal/platform/httpserver"
	tagadmin "github.com/mrstsgk/book-management-system/backend/internal/tag/presentation/http/admin"
	tagcmd "github.com/mrstsgk/book-management-system/backend/internal/tag/usecase/command"
)

const testSession = "test-session"

// fakeAdminOnly は本物の RequireAdminSession の代わり（Cookie の値が testSession なら通す）。
// セッションの分岐は common のテストが担うので、ここでは「認証が要る経路に付いているか」だけを見る
func fakeAdminOnly(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if ck, err := c.Cookie(httpserver.SessionCookieName); err != nil || ck.Value != testSession {
			return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
		}
		return next(c)
	}
}

type fakeRegister func(context.Context, tagcmd.RegisterTagCommand) (*tagcmd.TagView, error)

func (f fakeRegister) Execute(ctx context.Context, cmd tagcmd.RegisterTagCommand) (*tagcmd.TagView, error) {
	return f(ctx, cmd)
}

type fakeRename func(context.Context, tagcmd.RenameTagCommand) (*tagcmd.TagView, error)

func (f fakeRename) Execute(ctx context.Context, cmd tagcmd.RenameTagCommand) (*tagcmd.TagView, error) {
	return f(ctx, cmd)
}

type fakeDelete func(context.Context, int64) error

func (f fakeDelete) Execute(ctx context.Context, id int64) error { return f(ctx, id) }

// serve は cmd/api/main.go と同じ形（NewEcho + Register）でハンドラを組み立ててリクエストを流す。
func serve(t *testing.T, h *tagadmin.Handler, method, path, body string, withSession bool) *httptest.ResponseRecorder {
	t.Helper()
	orig := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, nil)))
	t.Cleanup(func() { slog.SetDefault(orig) })

	h.AdminOnly = fakeAdminOnly
	e := httpserver.NewEcho()
	h.Register(e.Group("/api/tags"))
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	if withSession {
		req.AddCookie(&http.Cookie{Name: httpserver.SessionCookieName, Value: testSession})
	}
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}

func mustNotCall(t *testing.T) func() {
	return func() { t.Error("usecase must not be called") }
}

func TestHandlerRegister(t *testing.T) {
	t.Run("ログイン済みなら登録して200", func(t *testing.T) {
		var got tagcmd.RegisterTagCommand
		h := &tagadmin.Handler{RegisterTagUsecase: fakeRegister(func(_ context.Context, cmd tagcmd.RegisterTagCommand) (*tagcmd.TagView, error) {
			got = cmd
			return &tagcmd.TagView{ID: 1, Name: "データベース", Version: 1}, nil
		})}
		rec := serve(t, h, http.MethodPost, "/api/tags", `{"name":"データベース"}`, true)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200 (body=%s)", rec.Code, rec.Body.String())
		}
		if got.Name != "データベース" {
			t.Fatalf("usecase received %+v", got)
		}
	})

	t.Run("ログインしていなければ401でusecaseを呼ばない", func(t *testing.T) {
		called := mustNotCall(t)
		h := &tagadmin.Handler{RegisterTagUsecase: fakeRegister(func(context.Context, tagcmd.RegisterTagCommand) (*tagcmd.TagView, error) {
			called()
			return nil, nil
		})}
		if rec := serve(t, h, http.MethodPost, "/api/tags", `{"name":"データベース"}`, false); rec.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d, want 401", rec.Code)
		}
	})

	for _, tt := range []struct {
		name string
		body string
		want []httpserver.FieldError
	}{
		{name: "必須項目なしは400", body: `{}`, want: []httpserver.FieldError{{Field: "name", Rule: "required"}}},
		{name: "31文字は400", body: `{"name":"` + strings.Repeat("あ", 31) + `"}`, want: []httpserver.FieldError{{Field: "name", Rule: "max"}}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			called := mustNotCall(t)
			h := &tagadmin.Handler{RegisterTagUsecase: fakeRegister(func(context.Context, tagcmd.RegisterTagCommand) (*tagcmd.TagView, error) {
				called()
				return nil, nil
			})}
			rec := serve(t, h, http.MethodPost, "/api/tags", tt.body, true)
			if rec.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want 400", rec.Code)
			}
			if got := (func() []httpserver.FieldError {
				var e httpserver.ErrorResponse
				_ = json.Unmarshal(rec.Body.Bytes(), &e)
				return e.Errors
			})(); !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("errors = %+v, want %+v", got, tt.want)
			}
		})
	}

	t.Run("同時追加で一意制約に当たったときは409", func(t *testing.T) {
		h := &tagadmin.Handler{RegisterTagUsecase: fakeRegister(func(context.Context, tagcmd.RegisterTagCommand) (*tagcmd.TagView, error) {
			return nil, domainerr.ErrConflict
		})}
		if rec := serve(t, h, http.MethodPost, "/api/tags", `{"name":"データベース"}`, true); rec.Code != http.StatusConflict {
			t.Fatalf("status = %d, want 409", rec.Code)
		}
	})

	t.Run("同名のタグはusecaseの違反として400でerrorsに載る", func(t *testing.T) {
		h := &tagadmin.Handler{RegisterTagUsecase: fakeRegister(func(context.Context, tagcmd.RegisterTagCommand) (*tagcmd.TagView, error) {
			return nil, domainerr.Violations{{Field: "name", Rule: "tagNameDuplicated"}}
		})}
		rec := serve(t, h, http.MethodPost, "/api/tags", `{"name":"データベース"}`, true)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want 400 (body=%s)", rec.Code, rec.Body.String())
		}
		var e httpserver.ErrorResponse
		_ = json.Unmarshal(rec.Body.Bytes(), &e)
		if want := []httpserver.FieldError{{Field: "name", Rule: "tagNameDuplicated"}}; !reflect.DeepEqual(e.Errors, want) {
			t.Fatalf("errors = %+v, want %+v", e.Errors, want)
		}
	})
}

func TestHandlerRename(t *testing.T) {
	body := `{"name":"新名","version":2}`

	t.Run("ログイン済みならパスのIDとバージョンをusecaseに渡し200で返す", func(t *testing.T) {
		var got tagcmd.RenameTagCommand
		h := &tagadmin.Handler{RenameTagUsecase: fakeRename(func(_ context.Context, cmd tagcmd.RenameTagCommand) (*tagcmd.TagView, error) {
			got = cmd
			return &tagcmd.TagView{ID: 5, Name: "新名", Version: 3}, nil
		})}
		rec := serve(t, h, http.MethodPut, "/api/tags/5", body, true)
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200 (body=%s)", rec.Code, rec.Body.String())
		}
		if want := (tagcmd.RenameTagCommand{ID: 5, Name: "新名", Version: 2}); got != want {
			t.Fatalf("usecase received %+v, want %+v", got, want)
		}
	})

	for _, tt := range []struct {
		name        string
		path        string
		body        string
		withSession bool
		err         error
		want        int
	}{
		{name: "ログインしていなければ401", path: "/api/tags/5", body: body, withSession: false, want: http.StatusUnauthorized},
		{name: "不正なIDは400", path: "/api/tags/abc", body: body, withSession: true, want: http.StatusBadRequest},
		{name: "バージョンなしは400", path: "/api/tags/5", body: `{"name":"新名"}`, withSession: true, want: http.StatusBadRequest},
		{name: "存在しないタグは404", path: "/api/tags/5", body: body, withSession: true, err: domainerr.ErrNotFound, want: http.StatusNotFound},
		{name: "楽観的ロックの競合は409", path: "/api/tags/5", body: body, withSession: true, err: domainerr.ErrConflict, want: http.StatusConflict},
		{name: "他のタグと同名はusecaseの違反として400", path: "/api/tags/5", body: body, withSession: true, err: domainerr.Violations{{Field: "name", Rule: "tagNameDuplicated"}}, want: http.StatusBadRequest},
	} {
		t.Run(tt.name, func(t *testing.T) {
			h := &tagadmin.Handler{RenameTagUsecase: fakeRename(func(context.Context, tagcmd.RenameTagCommand) (*tagcmd.TagView, error) {
				if tt.err == nil {
					t.Error("usecase must not be called")
				}
				return nil, tt.err
			})}
			if rec := serve(t, h, http.MethodPut, tt.path, tt.body, tt.withSession); rec.Code != tt.want {
				t.Fatalf("status = %d, want %d", rec.Code, tt.want)
			}
		})
	}
	t.Run("countsは/:idのIDとして扱われない（PUTはIDのパースで400）", func(t *testing.T) {
		h := &tagadmin.Handler{RenameTagUsecase: fakeRename(func(context.Context, tagcmd.RenameTagCommand) (*tagcmd.TagView, error) {
			t.Error("usecase must not be called")
			return nil, nil
		})}
		if rec := serve(t, h, http.MethodPut, "/api/tags/counts", `{"name":"x","version":1}`, true); rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want 400", rec.Code)
		}
	})
}

func TestHandlerDelete(t *testing.T) {
	t.Run("ログイン済みなら削除して204", func(t *testing.T) {
		var gotID int64
		h := &tagadmin.Handler{DeleteTagUsecase: fakeDelete(func(_ context.Context, id int64) error { gotID = id; return nil })}
		rec := serve(t, h, http.MethodDelete, "/api/tags/5", "", true)
		if rec.Code != http.StatusNoContent || gotID != 5 {
			t.Fatalf("status = %d id = %d, want 204 for id 5", rec.Code, gotID)
		}
	})

	for _, tt := range []struct {
		name        string
		withSession bool
		err         error
		want        int
	}{
		{name: "ログインしていなければ401", withSession: false, want: http.StatusUnauthorized},
		{name: "存在しないタグは404", withSession: true, err: domainerr.ErrNotFound, want: http.StatusNotFound},
	} {
		t.Run(tt.name, func(t *testing.T) {
			h := &tagadmin.Handler{DeleteTagUsecase: fakeDelete(func(context.Context, int64) error {
				if tt.err == nil {
					t.Error("usecase must not be called")
				}
				return tt.err
			})}
			if rec := serve(t, h, http.MethodDelete, "/api/tags/5", "", tt.withSession); rec.Code != tt.want {
				t.Fatalf("status = %d, want %d", rec.Code, tt.want)
			}
		})
	}
}
