package book_test

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

	domainbook "github.com/mrstsgk/book-management-system/backend/internal/domain/book"
	domaincommon "github.com/mrstsgk/book-management-system/backend/internal/domain/common"
	httpbook "github.com/mrstsgk/book-management-system/backend/internal/presentation/http/book"
	"github.com/mrstsgk/book-management-system/backend/internal/presentation/http/common"
	bookcmd "github.com/mrstsgk/book-management-system/backend/internal/usecase/book/command"
)

const adminToken = "test-admin-token"

type fakeRegister func(context.Context, bookcmd.RegisterCommand) (*domainbook.BookDetail, error)

func (f fakeRegister) Execute(ctx context.Context, cmd bookcmd.RegisterCommand) (*domainbook.BookDetail, error) {
	return f(ctx, cmd)
}

type fakeUpdate func(context.Context, bookcmd.UpdateCommand) (*domainbook.BookDetail, error)

func (f fakeUpdate) Execute(ctx context.Context, cmd bookcmd.UpdateCommand) (*domainbook.BookDetail, error) {
	return f(ctx, cmd)
}

type fakeDelete func(context.Context, int64) error

func (f fakeDelete) Execute(ctx context.Context, id int64) error { return f(ctx, id) }

type fakeGet func(context.Context, int64) (*domainbook.BookDetail, error)

func (f fakeGet) Execute(ctx context.Context, id int64) (*domainbook.BookDetail, error) {
	return f(ctx, id)
}

type fakeList func(context.Context, int, int) (*domainbook.BookList, error)

func (f fakeList) Execute(ctx context.Context, limit, offset int) (*domainbook.BookList, error) {
	return f(ctx, limit, offset)
}

// serve は cmd/api/main.go と同じ形（NewEcho + Register）でハンドラを組み立ててリクエストを流す。
func serve(t *testing.T, h *httpbook.Handler, method, path, body string, withToken bool) *httptest.ResponseRecorder {
	t.Helper()
	orig := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, nil)))
	t.Cleanup(func() { slog.SetDefault(orig) })

	h.AdminOnly = common.RequireAdminToken(adminToken)
	e := common.NewEcho()
	h.Register(e.Group("/api/books"))
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	if withToken {
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+adminToken)
	}
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}

func decode[T any](t *testing.T, rec *httptest.ResponseRecorder) T {
	t.Helper()
	var v T
	if err := json.Unmarshal(rec.Body.Bytes(), &v); err != nil {
		t.Fatalf("invalid JSON body %q: %v", rec.Body.String(), err)
	}
	return v
}

func strPtr(s string) *string { return &s }

var detail = &domainbook.BookDetail{
	ID: 1, ISBN: "9784873118703", Title: "データ指向アプリケーションデザイン", Authors: "Kleppmann,Martin",
	Publisher: "オーム社", PublishedOn: "201907", AmazonURL: strPtr("https://www.amazon.co.jp/dp/4873118700"),
	CoverURL: strPtr("https://thumbnail.image.rakuten.co.jp/1.jpg"), CoverSource: strPtr("rakuten"),
	Summary: "分散データの設計を学べる", Comment: "良書\n2行目", Rating: 5, Version: 1,
}

var detailResponse = httpbook.Response{
	ID: 1, ISBN: "9784873118703", Title: "データ指向アプリケーションデザイン", Authors: "Kleppmann,Martin",
	Publisher: "オーム社", PublishedOn: "201907", AmazonURL: strPtr("https://www.amazon.co.jp/dp/4873118700"),
	CoverURL: strPtr("https://thumbnail.image.rakuten.co.jp/1.jpg"), CoverSource: strPtr("rakuten"),
	Summary: "分散データの設計を学べる", Comment: "良書\n2行目", Rating: 5, Version: 1,
}

func mustNotCall(t *testing.T) func() {
	return func() { t.Error("usecase must not be called") }
}

func TestHandlerList(t *testing.T) {
	list := &domainbook.BookList{Total: 21, Items: []*domainbook.BookListItem{{
		ID: 1, ISBN: "9784873118703", Title: "データ指向アプリケーションデザイン", Summary: "分散データの設計を学べる", Authors: "Kleppmann,Martin",
		AmazonURL: strPtr("https://www.amazon.co.jp/dp/4873118700"), Rating: 5,
	}}}

	for _, tt := range []struct {
		name                  string
		query                 string
		wantLimit, wantOffset int
	}{
		{name: "クエリなしは既定の取得範囲", query: "", wantLimit: domaincommon.DefaultListLimit},
		{name: "limitとoffsetを指定できる", query: "?limit=5&offset=10", wantLimit: 5, wantOffset: 10},
	} {
		t.Run(tt.name+"で誰でも取得できる", func(t *testing.T) {
			var gotLimit, gotOffset int
			h := &httpbook.Handler{ListUC: fakeList(func(_ context.Context, limit, offset int) (*domainbook.BookList, error) {
				gotLimit, gotOffset = limit, offset
				return list, nil
			})}
			rec := serve(t, h, http.MethodGet, "/api/books"+tt.query, "", false)

			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d, want 200 (body=%s)", rec.Code, rec.Body.String())
			}
			if gotLimit != tt.wantLimit || gotOffset != tt.wantOffset {
				t.Fatalf("usecase received limit=%d offset=%d", gotLimit, gotOffset)
			}
			want := httpbook.ListResponse{Total: 21, Limit: tt.wantLimit, Offset: tt.wantOffset, Items: []httpbook.ListItemResponse{{
				ID: 1, ISBN: "9784873118703", Title: "データ指向アプリケーションデザイン", Summary: "分散データの設計を学べる", Authors: "Kleppmann,Martin",
				AmazonURL: strPtr("https://www.amazon.co.jp/dp/4873118700"), Rating: 5,
			}}}
			if got := decode[httpbook.ListResponse](t, rec); !reflect.DeepEqual(got, want) {
				t.Fatalf("body = %+v, want %+v", got, want)
			}
		})
	}

	t.Run("0件は空配列", func(t *testing.T) {
		h := &httpbook.Handler{ListUC: fakeList(func(context.Context, int, int) (*domainbook.BookList, error) {
			return &domainbook.BookList{}, nil
		})}
		rec := serve(t, h, http.MethodGet, "/api/books", "", false)
		if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"items":[]`) {
			t.Fatalf("status = %d body = %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("limit上限+1は400", func(t *testing.T) {
		called := mustNotCall(t)
		h := &httpbook.Handler{ListUC: fakeList(func(context.Context, int, int) (*domainbook.BookList, error) { called(); return nil, nil })}
		rec := serve(t, h, http.MethodGet, "/api/books?limit=101", "", false)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want 400", rec.Code)
		}
	})
}

func TestHandlerGet(t *testing.T) {
	t.Run("誰でも詳細を取得できる", func(t *testing.T) {
		h := &httpbook.Handler{GetUC: fakeGet(func(_ context.Context, id int64) (*domainbook.BookDetail, error) {
			if id != 1 {
				t.Errorf("id = %d, want 1", id)
			}
			return detail, nil
		})}
		rec := serve(t, h, http.MethodGet, "/api/books/1", "", false)
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200", rec.Code)
		}
		if got := decode[httpbook.Response](t, rec); !reflect.DeepEqual(got, detailResponse) {
			t.Fatalf("body = %+v, want %+v", got, detailResponse)
		}
	})

	t.Run("存在しない本は404", func(t *testing.T) {
		h := &httpbook.Handler{GetUC: fakeGet(func(context.Context, int64) (*domainbook.BookDetail, error) { return nil, domaincommon.ErrNotFound })}
		if rec := serve(t, h, http.MethodGet, "/api/books/1", "", false); rec.Code != http.StatusNotFound {
			t.Fatalf("status = %d, want 404", rec.Code)
		}
	})
}

func TestHandlerRegister(t *testing.T) {
	body := `{"isbn":"978-4-87311-870-3","summary":"分散データの設計を学べる","comment":"良書","rating":5}`

	t.Run("トークンがあれば入力をそのままusecaseに渡し200で返す", func(t *testing.T) {
		var got bookcmd.RegisterCommand
		h := &httpbook.Handler{RegisterUC: fakeRegister(func(_ context.Context, cmd bookcmd.RegisterCommand) (*domainbook.BookDetail, error) {
			got = cmd
			return detail, nil
		})}
		rec := serve(t, h, http.MethodPost, "/api/books", body, true)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200 (body=%s)", rec.Code, rec.Body.String())
		}
		if want := (bookcmd.RegisterCommand{ISBN: "978-4-87311-870-3", Summary: "分散データの設計を学べる", Comment: "良書", Rating: 5}); !reflect.DeepEqual(got, want) {
			t.Fatalf("usecase received %+v, want %+v", got, want)
		}
	})

	t.Run("書名の上書きを指定するとusecaseに渡る", func(t *testing.T) {
		var got bookcmd.RegisterCommand
		h := &httpbook.Handler{RegisterUC: fakeRegister(func(_ context.Context, cmd bookcmd.RegisterCommand) (*domainbook.BookDetail, error) {
			got = cmd
			return detail, nil
		})}
		withOverride := `{"isbn":"978-4-87311-870-3","summary":"分散データの設計を学べる","titleOverride":"正しい書名","comment":"良書","rating":5}`
		rec := serve(t, h, http.MethodPost, "/api/books", withOverride, true)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200 (body=%s)", rec.Code, rec.Body.String())
		}
		if got.TitleOverride != "正しい書名" {
			t.Fatalf("usecase received TitleOverride=%q, want 正しい書名", got.TitleOverride)
		}
	})

	t.Run("書名の上書きを省略すると空文字でusecaseに渡る", func(t *testing.T) {
		var got bookcmd.RegisterCommand
		h := &httpbook.Handler{RegisterUC: fakeRegister(func(_ context.Context, cmd bookcmd.RegisterCommand) (*domainbook.BookDetail, error) {
			got = cmd
			return detail, nil
		})}
		rec := serve(t, h, http.MethodPost, "/api/books", body, true)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200 (body=%s)", rec.Code, rec.Body.String())
		}
		if got.TitleOverride != "" {
			t.Fatalf("usecase received TitleOverride=%q, want empty", got.TitleOverride)
		}
	})

	t.Run("トークンが無ければ401でusecaseを呼ばない", func(t *testing.T) {
		called := mustNotCall(t)
		h := &httpbook.Handler{RegisterUC: fakeRegister(func(context.Context, bookcmd.RegisterCommand) (*domainbook.BookDetail, error) {
			called()
			return nil, nil
		})}
		if rec := serve(t, h, http.MethodPost, "/api/books", body, false); rec.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d, want 401", rec.Code)
		}
	})

	for _, tt := range []struct {
		name string
		body string
		want []common.FieldError
	}{
		{name: "必須項目なしは400", body: `{}`, want: []common.FieldError{{Field: "isbn", Rule: "required"}, {Field: "summary", Rule: "required"}, {Field: "comment", Rule: "required"}, {Field: "rating", Rule: "required"}}},
		{name: "評価6は400", body: `{"isbn":"4873118700","summary":"要約","comment":"良書","rating":6}`, want: []common.FieldError{{Field: "rating", Rule: "max"}}},
		{name: "評価0は400", body: `{"isbn":"4873118700","summary":"要約","comment":"良書","rating":0}`, want: []common.FieldError{{Field: "rating", Rule: "min"}}},
		{name: "ISBN18文字は400", body: `{"isbn":"` + strings.Repeat("9", 18) + `","summary":"要約","comment":"良書","rating":5}`, want: []common.FieldError{{Field: "isbn", Rule: "max"}}},
		{name: "感想5001文字は400", body: `{"isbn":"4873118700","summary":"要約","comment":"` + strings.Repeat("あ", 5001) + `","rating":5}`, want: []common.FieldError{{Field: "comment", Rule: "max"}}},
		{name: "一言まとめが無いは400", body: `{"isbn":"4873118700","comment":"良書","rating":5}`, want: []common.FieldError{{Field: "summary", Rule: "required"}}},
		{name: "一言まとめが101文字は400", body: `{"isbn":"4873118700","summary":"` + strings.Repeat("あ", 101) + `","comment":"良書","rating":5}`, want: []common.FieldError{{Field: "summary", Rule: "max"}}},
		{name: "書名の上書きが256文字は400", body: `{"isbn":"4873118700","summary":"要約","titleOverride":"` + strings.Repeat("あ", 256) + `","comment":"良書","rating":5}`, want: []common.FieldError{{Field: "titleOverride", Rule: "max"}}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			called := mustNotCall(t)
			h := &httpbook.Handler{RegisterUC: fakeRegister(func(context.Context, bookcmd.RegisterCommand) (*domainbook.BookDetail, error) {
				called()
				return nil, nil
			})}
			rec := serve(t, h, http.MethodPost, "/api/books", tt.body, true)
			if rec.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want 400", rec.Code)
			}
			if got := decode[common.ErrorResponse](t, rec).Errors; !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("errors = %+v, want %+v", got, tt.want)
			}
		})
	}

	for _, tt := range []struct {
		name string
		err  error
		want int
	}{
		{name: "カタログに無いISBNなどドメインの不正は400", err: domaincommon.ErrInvalid, want: http.StatusBadRequest},
		{name: "同じISBNの登録済みは409", err: domaincommon.ErrConflict, want: http.StatusConflict},
	} {
		t.Run(tt.name, func(t *testing.T) {
			h := &httpbook.Handler{RegisterUC: fakeRegister(func(context.Context, bookcmd.RegisterCommand) (*domainbook.BookDetail, error) { return nil, tt.err })}
			if rec := serve(t, h, http.MethodPost, "/api/books", body, true); rec.Code != tt.want {
				t.Fatalf("status = %d, want %d", rec.Code, tt.want)
			}
		})
	}
}

func TestHandlerUpdate(t *testing.T) {
	body := `{"summary":"読み返してのまとめ","comment":"読み返した","rating":4,"version":2}`

	t.Run("トークンがあればパスのIDとバージョンをusecaseに渡し200で返す", func(t *testing.T) {
		var got bookcmd.UpdateCommand
		h := &httpbook.Handler{UpdateUC: fakeUpdate(func(_ context.Context, cmd bookcmd.UpdateCommand) (*domainbook.BookDetail, error) {
			got = cmd
			return detail, nil
		})}
		rec := serve(t, h, http.MethodPut, "/api/books/5", body, true)
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200 (body=%s)", rec.Code, rec.Body.String())
		}
		if want := (bookcmd.UpdateCommand{ID: 5, Summary: "読み返してのまとめ", Comment: "読み返した", Rating: 4, Version: 2}); !reflect.DeepEqual(got, want) {
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
		{name: "トークンが無ければ401", path: "/api/books/5", body: body, withToken: false, want: http.StatusUnauthorized},
		{name: "不正なIDは400", path: "/api/books/abc", body: body, withToken: true, want: http.StatusBadRequest},
		{name: "バージョンなしは400", path: "/api/books/5", body: `{"comment":"x","rating":4}`, withToken: true, want: http.StatusBadRequest},
		{name: "存在しない本は404", path: "/api/books/5", body: body, withToken: true, err: domaincommon.ErrNotFound, want: http.StatusNotFound},
		{name: "楽観的ロックの競合は409", path: "/api/books/5", body: body, withToken: true, err: domaincommon.ErrConflict, want: http.StatusConflict},
	} {
		t.Run(tt.name, func(t *testing.T) {
			h := &httpbook.Handler{UpdateUC: fakeUpdate(func(context.Context, bookcmd.UpdateCommand) (*domainbook.BookDetail, error) {
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
		h := &httpbook.Handler{DeleteUC: fakeDelete(func(_ context.Context, id int64) error { gotID = id; return nil })}
		rec := serve(t, h, http.MethodDelete, "/api/books/5", "", true)
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
		{name: "存在しない本は404", withToken: true, err: domaincommon.ErrNotFound, want: http.StatusNotFound},
	} {
		t.Run(tt.name, func(t *testing.T) {
			h := &httpbook.Handler{DeleteUC: fakeDelete(func(context.Context, int64) error {
				if tt.err == nil {
					t.Error("usecase must not be called")
				}
				return tt.err
			})}
			if rec := serve(t, h, http.MethodDelete, "/api/books/5", "", tt.withToken); rec.Code != tt.want {
				t.Fatalf("status = %d, want %d", rec.Code, tt.want)
			}
		})
	}
}
