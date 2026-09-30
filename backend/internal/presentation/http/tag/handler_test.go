package tag_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"

	domaincommon "github.com/mrstsgk/book-management-system/backend/internal/domain/common"
	domaintag "github.com/mrstsgk/book-management-system/backend/internal/domain/tag"
	"github.com/mrstsgk/book-management-system/backend/internal/presentation/http/common"
	httptag "github.com/mrstsgk/book-management-system/backend/internal/presentation/http/tag"
	tagcmd "github.com/mrstsgk/book-management-system/backend/internal/usecase/tag/command"
)

const testSession = "test-session"

// fakeAdminOnly は本物の RequireAdminSession の代わり（Cookie の値が testSession なら通す）。
// セッションの分岐は common のテストが担うので、ここでは「認証が要る経路に付いているか」だけを見る
func fakeAdminOnly(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if ck, err := c.Cookie(common.SessionCookieName); err != nil || ck.Value != testSession {
			return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
		}
		return next(c)
	}
}

type fakeRegister func(context.Context, tagcmd.RegisterCommand) (*tagcmd.TagView, error)

func (f fakeRegister) Execute(ctx context.Context, cmd tagcmd.RegisterCommand) (*tagcmd.TagView, error) {
	return f(ctx, cmd)
}

type fakeRename func(context.Context, tagcmd.RenameCommand) (*tagcmd.TagView, error)

func (f fakeRename) Execute(ctx context.Context, cmd tagcmd.RenameCommand) (*tagcmd.TagView, error) {
	return f(ctx, cmd)
}

type fakeDelete func(context.Context, int64) error

func (f fakeDelete) Execute(ctx context.Context, id int64) error { return f(ctx, id) }

type fakeList func(context.Context) (*domaintag.TagList, error)

func (f fakeList) Execute(ctx context.Context) (*domaintag.TagList, error) { return f(ctx) }

type fakeCountBooks func(context.Context) (*domaintag.TagBookCounts, error)

func (f fakeCountBooks) Execute(ctx context.Context) (*domaintag.TagBookCounts, error) { return f(ctx) }

func TestHandlerCountBooks(t *testing.T) {
	t.Run("ログイン無しでも分野タグごとの冊数を返す", func(t *testing.T) {
		counts := &domaintag.TagBookCounts{Items: []*domaintag.TagBookCount{{ID: 1, Name: "設計", BookCount: 3}}}
		h := &httptag.Handler{CountBooksUC: fakeCountBooks(func(context.Context) (*domaintag.TagBookCounts, error) { return counts, nil })}

		rec := serve(t, h, http.MethodGet, "/api/tags/counts", "", false)
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200 (body=%s)", rec.Code, rec.Body.String())
		}
		var got httptag.BookCountListResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
			t.Fatal(err)
		}
		want := httptag.BookCountListResponse{Items: []httptag.BookCountResponse{{ID: 1, Name: "設計", BookCount: 3}}}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("body = %+v, want %+v", got, want)
		}
	})

	t.Run("0件は空配列", func(t *testing.T) {
		h := &httptag.Handler{CountBooksUC: fakeCountBooks(func(context.Context) (*domaintag.TagBookCounts, error) { return &domaintag.TagBookCounts{}, nil })}
		rec := serve(t, h, http.MethodGet, "/api/tags/counts", "", false)
		if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"items":[]`) {
			t.Fatalf("status = %d body = %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("ユースケースの障害は500", func(t *testing.T) {
		h := &httptag.Handler{CountBooksUC: fakeCountBooks(func(context.Context) (*domaintag.TagBookCounts, error) { return nil, errors.New("db: timeout") })}
		if rec := serve(t, h, http.MethodGet, "/api/tags/counts", "", false); rec.Code != http.StatusInternalServerError {
			t.Fatalf("status = %d, want 500", rec.Code)
		}
	})

	t.Run("countsは/:idのIDとして扱われない（PUTはIDのパースで400）", func(t *testing.T) {
		h := &httptag.Handler{RenameUC: fakeRename(func(context.Context, tagcmd.RenameCommand) (*tagcmd.TagView, error) {
			t.Error("usecase must not be called")
			return nil, nil
		})}
		if rec := serve(t, h, http.MethodPut, "/api/tags/counts", `{"name":"x","version":1}`, true); rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want 400", rec.Code)
		}
	})
}

// serve は cmd/api/main.go と同じ形（NewEcho + Register）でハンドラを組み立ててリクエストを流す。
func serve(t *testing.T, h *httptag.Handler, method, path, body string, withSession bool) *httptest.ResponseRecorder {
	t.Helper()
	orig := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, nil)))
	t.Cleanup(func() { slog.SetDefault(orig) })

	h.AdminOnly = fakeAdminOnly
	e := common.NewEcho()
	h.Register(e.Group("/api/tags"))
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	if withSession {
		req.AddCookie(&http.Cookie{Name: common.SessionCookieName, Value: testSession})
	}
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}

func mustNotCall(t *testing.T) func() {
	return func() { t.Error("usecase must not be called") }
}

func TestHandlerList(t *testing.T) {
	list := &domaintag.TagList{Items: []*domaintag.TagListItem{{ID: 1, Name: "データベース", Version: 1}}}
	h := &httptag.Handler{ListUC: fakeList(func(context.Context) (*domaintag.TagList, error) { return list, nil })}

	rec := serve(t, h, http.MethodGet, "/api/tags", "", false)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (認証不要)", rec.Code)
	}
	var got httptag.ListResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	want := httptag.ListResponse{Items: []httptag.Response{{ID: 1, Name: "データベース", Version: 1}}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("body = %+v, want %+v", got, want)
	}
}

func TestHandlerRegister(t *testing.T) {
	t.Run("ログイン済みなら登録して200", func(t *testing.T) {
		var got tagcmd.RegisterCommand
		h := &httptag.Handler{RegisterUC: fakeRegister(func(_ context.Context, cmd tagcmd.RegisterCommand) (*tagcmd.TagView, error) {
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
		h := &httptag.Handler{RegisterUC: fakeRegister(func(context.Context, tagcmd.RegisterCommand) (*tagcmd.TagView, error) {
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
		want []common.FieldError
	}{
		{name: "必須項目なしは400", body: `{}`, want: []common.FieldError{{Field: "name", Rule: "required"}}},
		{name: "31文字は400", body: `{"name":"` + strings.Repeat("あ", 31) + `"}`, want: []common.FieldError{{Field: "name", Rule: "max"}}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			called := mustNotCall(t)
			h := &httptag.Handler{RegisterUC: fakeRegister(func(context.Context, tagcmd.RegisterCommand) (*tagcmd.TagView, error) {
				called()
				return nil, nil
			})}
			rec := serve(t, h, http.MethodPost, "/api/tags", tt.body, true)
			if rec.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want 400", rec.Code)
			}
			if got := (func() []common.FieldError {
				var e common.ErrorResponse
				_ = json.Unmarshal(rec.Body.Bytes(), &e)
				return e.Errors
			})(); !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("errors = %+v, want %+v", got, tt.want)
			}
		})
	}

	t.Run("同名の登録済みは409", func(t *testing.T) {
		h := &httptag.Handler{RegisterUC: fakeRegister(func(context.Context, tagcmd.RegisterCommand) (*tagcmd.TagView, error) {
			return nil, domaincommon.ErrConflict
		})}
		if rec := serve(t, h, http.MethodPost, "/api/tags", `{"name":"データベース"}`, true); rec.Code != http.StatusConflict {
			t.Fatalf("status = %d, want 409", rec.Code)
		}
	})
}

func TestHandlerRename(t *testing.T) {
	body := `{"name":"新名","version":2}`

	t.Run("ログイン済みならパスのIDとバージョンをusecaseに渡し200で返す", func(t *testing.T) {
		var got tagcmd.RenameCommand
		h := &httptag.Handler{RenameUC: fakeRename(func(_ context.Context, cmd tagcmd.RenameCommand) (*tagcmd.TagView, error) {
			got = cmd
			return &tagcmd.TagView{ID: 5, Name: "新名", Version: 3}, nil
		})}
		rec := serve(t, h, http.MethodPut, "/api/tags/5", body, true)
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200 (body=%s)", rec.Code, rec.Body.String())
		}
		if want := (tagcmd.RenameCommand{ID: 5, Name: "新名", Version: 2}); got != want {
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
		{name: "存在しないタグは404", path: "/api/tags/5", body: body, withSession: true, err: domaincommon.ErrNotFound, want: http.StatusNotFound},
		{name: "楽観的ロックの競合は409", path: "/api/tags/5", body: body, withSession: true, err: domaincommon.ErrConflict, want: http.StatusConflict},
	} {
		t.Run(tt.name, func(t *testing.T) {
			h := &httptag.Handler{RenameUC: fakeRename(func(context.Context, tagcmd.RenameCommand) (*tagcmd.TagView, error) {
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
}

func TestHandlerDelete(t *testing.T) {
	t.Run("ログイン済みなら削除して204", func(t *testing.T) {
		var gotID int64
		h := &httptag.Handler{DeleteUC: fakeDelete(func(_ context.Context, id int64) error { gotID = id; return nil })}
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
		{name: "存在しないタグは404", withSession: true, err: domaincommon.ErrNotFound, want: http.StatusNotFound},
	} {
		t.Run(tt.name, func(t *testing.T) {
			h := &httptag.Handler{DeleteUC: fakeDelete(func(context.Context, int64) error {
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
