package tag_test

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

	domaincommon "github.com/mrstsgk/book-management-system/backend/internal/domain/common"
	domaintag "github.com/mrstsgk/book-management-system/backend/internal/domain/tag"
	"github.com/mrstsgk/book-management-system/backend/internal/presentation/http/common"
	httptag "github.com/mrstsgk/book-management-system/backend/internal/presentation/http/tag"
	tagcmd "github.com/mrstsgk/book-management-system/backend/internal/usecase/tag/command"
)

const adminToken = "test-admin-token"

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

// serve は cmd/api/main.go と同じ形（NewEcho + Register）でハンドラを組み立ててリクエストを流す。
func serve(t *testing.T, h *httptag.Handler, method, path, body string, withToken bool) *httptest.ResponseRecorder {
	t.Helper()
	orig := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, nil)))
	t.Cleanup(func() { slog.SetDefault(orig) })

	h.AdminOnly = common.RequireAdminToken(adminToken)
	e := common.NewEcho()
	h.Register(e.Group("/api/tags"))
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	if withToken {
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+adminToken)
	}
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}

func mustNotCall(t *testing.T) func() {
	return func() { t.Error("usecase must not be called") }
}

func TestHandlerList(t *testing.T) {
	list := &domaintag.TagList{Items: []*domaintag.TagListItem{{ID: 1, Name: "データベース"}}}
	h := &httptag.Handler{ListUC: fakeList(func(context.Context) (*domaintag.TagList, error) { return list, nil })}

	rec := serve(t, h, http.MethodGet, "/api/tags", "", false)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (認証不要)", rec.Code)
	}
	var got httptag.ListResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	want := httptag.ListResponse{Items: []httptag.Response{{ID: 1, Name: "データベース"}}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("body = %+v, want %+v", got, want)
	}
}

func TestHandlerRegister(t *testing.T) {
	t.Run("トークンがあれば登録して200", func(t *testing.T) {
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

	t.Run("トークンが無ければ401でusecaseを呼ばない", func(t *testing.T) {
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

	t.Run("トークンがあればパスのIDとバージョンをusecaseに渡し200で返す", func(t *testing.T) {
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
		name      string
		path      string
		body      string
		withToken bool
		err       error
		want      int
	}{
		{name: "トークンが無ければ401", path: "/api/tags/5", body: body, withToken: false, want: http.StatusUnauthorized},
		{name: "不正なIDは400", path: "/api/tags/abc", body: body, withToken: true, want: http.StatusBadRequest},
		{name: "バージョンなしは400", path: "/api/tags/5", body: `{"name":"新名"}`, withToken: true, want: http.StatusBadRequest},
		{name: "存在しないタグは404", path: "/api/tags/5", body: body, withToken: true, err: domaincommon.ErrNotFound, want: http.StatusNotFound},
		{name: "楽観的ロックの競合は409", path: "/api/tags/5", body: body, withToken: true, err: domaincommon.ErrConflict, want: http.StatusConflict},
	} {
		t.Run(tt.name, func(t *testing.T) {
			h := &httptag.Handler{RenameUC: fakeRename(func(context.Context, tagcmd.RenameCommand) (*tagcmd.TagView, error) {
				if tt.err == nil {
					t.Error("usecase must not be called")
				}
				return nil, tt.err
			})}
			if rec := serve(t, h, http.MethodPut, tt.path, tt.body, tt.withToken); rec.Code != tt.want {
				t.Fatalf("status = %d, want %d", rec.Code, tt.want)
			}
		})
	}
}

func TestHandlerDelete(t *testing.T) {
	t.Run("トークンがあれば削除して204", func(t *testing.T) {
		var gotID int64
		h := &httptag.Handler{DeleteUC: fakeDelete(func(_ context.Context, id int64) error { gotID = id; return nil })}
		rec := serve(t, h, http.MethodDelete, "/api/tags/5", "", true)
		if rec.Code != http.StatusNoContent || gotID != 5 {
			t.Fatalf("status = %d id = %d, want 204 for id 5", rec.Code, gotID)
		}
	})

	for _, tt := range []struct {
		name      string
		withToken bool
		err       error
		want      int
	}{
		{name: "トークンが無ければ401", withToken: false, want: http.StatusUnauthorized},
		{name: "存在しないタグは404", withToken: true, err: domaincommon.ErrNotFound, want: http.StatusNotFound},
	} {
		t.Run(tt.name, func(t *testing.T) {
			h := &httptag.Handler{DeleteUC: fakeDelete(func(context.Context, int64) error {
				if tt.err == nil {
					t.Error("usecase must not be called")
				}
				return tt.err
			})}
			if rec := serve(t, h, http.MethodDelete, "/api/tags/5", "", tt.withToken); rec.Code != tt.want {
				t.Fatalf("status = %d, want %d", rec.Code, tt.want)
			}
		})
	}
}
