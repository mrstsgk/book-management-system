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

	domainbook "github.com/mrstsgk/book-management-system/backend/internal/domain/book"
	domaincommon "github.com/mrstsgk/book-management-system/backend/internal/domain/common"
	httpbook "github.com/mrstsgk/book-management-system/backend/internal/presentation/http/book"
	"github.com/mrstsgk/book-management-system/backend/internal/presentation/http/common"
	bookcmd "github.com/mrstsgk/book-management-system/backend/internal/usecase/book/command"
)

type fakeCreate func(ctx context.Context, cmd bookcmd.CreateCommand) (*domainbook.BookDetail, error)

func (f fakeCreate) Execute(ctx context.Context, cmd bookcmd.CreateCommand) (*domainbook.BookDetail, error) {
	return f(ctx, cmd)
}

type fakeUpdate func(ctx context.Context, cmd bookcmd.UpdateCommand) (*domainbook.BookDetail, error)

func (f fakeUpdate) Execute(ctx context.Context, cmd bookcmd.UpdateCommand) (*domainbook.BookDetail, error) {
	return f(ctx, cmd)
}

// serve wires the handler the same way cmd/api/main.go does (NewEcho + Register).
func serve(t *testing.T, h *httpbook.Handler, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	orig := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, nil)))
	t.Cleanup(func() { slog.SetDefault(orig) })

	e := common.NewEcho()
	h.Register(e.Group("/api/books"))
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
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

var detail = &domainbook.BookDetail{
	ID: 1, Title: "人間失格", Price: 1500, Status: 2, Version: 1,
	Authors: []domainbook.AuthorSummary{{ID: 1, Name: "太宰治", Version: 1}},
}

var detailResponse = httpbook.Response{
	ID: 1, Title: "人間失格", Price: 1500, Status: 2, Version: 1,
	Authors: []httpbook.AuthorResponse{{ID: 1, Name: "太宰治", Version: 1}},
}

func TestHandlerCreate(t *testing.T) {
	t.Run("作成して200で書籍詳細を返す", func(t *testing.T) {
		var got bookcmd.CreateCommand
		h := &httpbook.Handler{CreateUC: fakeCreate(func(_ context.Context, cmd bookcmd.CreateCommand) (*domainbook.BookDetail, error) {
			got = cmd
			return detail, nil
		})}
		rec := serve(t, h, http.MethodPost, "/api/books", `{"title":"人間失格","price":0,"authorIds":[1,2],"status":2}`)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200 (body=%s)", rec.Code, rec.Body.String())
		}
		want := bookcmd.CreateCommand{Title: "人間失格", Price: 0, AuthorIDs: []int64{1, 2}, Status: 2}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("usecase received %+v, want %+v", got, want)
		}
		if res := decode[httpbook.Response](t, rec); !reflect.DeepEqual(res, detailResponse) {
			t.Fatalf("body = %+v, want %+v", res, detailResponse)
		}
	})

	invalid := []struct {
		name string
		body string
		want []common.FieldError
	}{
		{name: "必須項目なしは400", body: `{}`, want: []common.FieldError{
			{Field: "title", Rule: "required"}, {Field: "price", Rule: "required"},
			{Field: "authorIds", Rule: "required"}, {Field: "status", Rule: "required"},
		}},
		{name: "タイトル256文字は400", body: `{"title":"` + strings.Repeat("あ", 256) + `","price":1,"authorIds":[1],"status":1}`, want: []common.FieldError{{Field: "title", Rule: "max"}}},
		{name: "価格が負は400", body: `{"title":"a","price":-1,"authorIds":[1],"status":1}`, want: []common.FieldError{{Field: "price", Rule: "min"}}},
		{name: "価格が上限超過は400", body: `{"title":"a","price":100000000,"authorIds":[1],"status":1}`, want: []common.FieldError{{Field: "price", Rule: "max"}}},
		{name: "著者0人は400", body: `{"title":"a","price":1,"authorIds":[],"status":1}`, want: []common.FieldError{{Field: "authorIds", Rule: "min"}}},
		{name: "著者IDが0以下は400", body: `{"title":"a","price":1,"authorIds":[0],"status":1}`, want: []common.FieldError{{Field: "authorIds[0]", Rule: "gt"}}},
		{name: "不正な出版状況は400", body: `{"title":"a","price":1,"authorIds":[1],"status":3}`, want: []common.FieldError{{Field: "status", Rule: "oneof"}}},
	}
	for _, tt := range invalid {
		t.Run(tt.name, func(t *testing.T) {
			h := &httpbook.Handler{CreateUC: fakeCreate(func(context.Context, bookcmd.CreateCommand) (*domainbook.BookDetail, error) {
				t.Error("usecase must not be called")
				return nil, nil
			})}
			rec := serve(t, h, http.MethodPost, "/api/books", tt.body)

			if rec.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want 400", rec.Code)
			}
			if got := decode[common.ErrorResponse](t, rec).Errors; !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("errors = %+v, want %+v", got, tt.want)
			}
		})
	}

	t.Run("存在しない著者などドメインルール違反は400", func(t *testing.T) {
		h := &httpbook.Handler{CreateUC: fakeCreate(func(context.Context, bookcmd.CreateCommand) (*domainbook.BookDetail, error) {
			return nil, domaincommon.ErrInvalid
		})}
		if rec := serve(t, h, http.MethodPost, "/api/books", `{"title":"a","price":1,"authorIds":[999],"status":1}`); rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want 400", rec.Code)
		}
	})
}

func TestHandlerUpdate(t *testing.T) {
	body := `{"title":"人間失格","price":1500,"authorIds":[1],"status":2,"version":1}`

	t.Run("パスのIDとバージョンをusecaseに渡し200で返す", func(t *testing.T) {
		var got bookcmd.UpdateCommand
		h := &httpbook.Handler{UpdateUC: fakeUpdate(func(_ context.Context, cmd bookcmd.UpdateCommand) (*domainbook.BookDetail, error) {
			got = cmd
			return detail, nil
		})}
		rec := serve(t, h, http.MethodPut, "/api/books/5", body)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200 (body=%s)", rec.Code, rec.Body.String())
		}
		want := bookcmd.UpdateCommand{ID: 5, Title: "人間失格", Price: 1500, AuthorIDs: []int64{1}, Status: 2, Version: 1}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("usecase received %+v, want %+v", got, want)
		}
	})

	statuses := []struct {
		name string
		path string
		body string
		err  error
		want int
	}{
		{name: "不正なIDは400", path: "/api/books/abc", body: body, want: http.StatusBadRequest},
		{name: "バージョンなしは400", path: "/api/books/5", body: `{"title":"a","price":1,"authorIds":[1],"status":1}`, want: http.StatusBadRequest},
		{name: "出版済みから未出版への変更は400", path: "/api/books/5", body: body, err: domaincommon.ErrInvalid, want: http.StatusBadRequest},
		{name: "存在しない書籍は404", path: "/api/books/5", body: body, err: domaincommon.ErrNotFound, want: http.StatusNotFound},
		{name: "楽観的ロックの競合は409", path: "/api/books/5", body: body, err: domaincommon.ErrConflict, want: http.StatusConflict},
	}
	for _, tt := range statuses {
		t.Run(tt.name, func(t *testing.T) {
			h := &httpbook.Handler{UpdateUC: fakeUpdate(func(context.Context, bookcmd.UpdateCommand) (*domainbook.BookDetail, error) {
				if tt.err == nil {
					t.Error("usecase must not be called")
				}
				return nil, tt.err
			})}
			if rec := serve(t, h, http.MethodPut, tt.path, tt.body); rec.Code != tt.want {
				t.Fatalf("status = %d, want %d (body=%s)", rec.Code, tt.want, rec.Body.String())
			}
		})
	}
}

type fakeGet func(ctx context.Context, id int64) (*domainbook.BookDetail, error)

func (f fakeGet) Execute(ctx context.Context, id int64) (*domainbook.BookDetail, error) {
	return f(ctx, id)
}

func strPtr(s string) *string { return &s }

func TestHandlerGet(t *testing.T) {
	t.Run("ISBN・Amazonリンク・書影を含む書籍詳細を返す", func(t *testing.T) {
		var gotID int64
		d := *detail
		d.ISBN = strPtr("9784873118703")
		d.AmazonURL = strPtr("https://www.amazon.co.jp/dp/4873118700")
		d.CoverURL = strPtr("https://cover.openbd.jp/9784873118703.jpg")
		d.CoverSource = strPtr("openbd")
		h := &httpbook.Handler{GetUC: fakeGet(func(_ context.Context, id int64) (*domainbook.BookDetail, error) {
			gotID = id
			return &d, nil
		})}
		rec := serve(t, h, http.MethodGet, "/api/books/1", "")

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200 (body=%s)", rec.Code, rec.Body.String())
		}
		want := detailResponse
		want.ISBN = d.ISBN
		want.AmazonURL = d.AmazonURL
		want.CoverURL = d.CoverURL
		want.CoverSource = d.CoverSource
		if got := decode[httpbook.Response](t, rec); gotID != 1 || !reflect.DeepEqual(got, want) {
			t.Fatalf("id=%d body = %+v, want %+v", gotID, got, want)
		}
	})

	t.Run("ISBNも書影もなければnullで返す", func(t *testing.T) {
		h := &httpbook.Handler{GetUC: fakeGet(func(context.Context, int64) (*domainbook.BookDetail, error) { return detail, nil })}
		rec := serve(t, h, http.MethodGet, "/api/books/1", "")

		for _, f := range []string{`"isbn":null`, `"amazonUrl":null`, `"coverUrl":null`, `"coverSource":null`} {
			if !strings.Contains(rec.Body.String(), f) {
				t.Fatalf("body = %s, want %s", rec.Body.String(), f)
			}
		}
	})

	statuses := []struct {
		name string
		path string
		err  error
		want int
	}{
		{name: "不正なIDは400", path: "/api/books/abc", want: http.StatusBadRequest},
		{name: "存在しない書籍は404", path: "/api/books/1", err: domaincommon.ErrNotFound, want: http.StatusNotFound},
	}
	for _, tt := range statuses {
		t.Run(tt.name, func(t *testing.T) {
			h := &httpbook.Handler{GetUC: fakeGet(func(context.Context, int64) (*domainbook.BookDetail, error) {
				if tt.err == nil {
					t.Error("usecase must not be called")
				}
				return nil, tt.err
			})}
			if rec := serve(t, h, http.MethodGet, tt.path, ""); rec.Code != tt.want {
				t.Fatalf("status = %d, want %d", rec.Code, tt.want)
			}
		})
	}
}

func TestHandlerCreateAndUpdate_PassISBN(t *testing.T) {
	const isbn = "978-4-87311-870-3"
	var created, updated *string
	h := &httpbook.Handler{
		CreateUC: fakeCreate(func(_ context.Context, cmd bookcmd.CreateCommand) (*domainbook.BookDetail, error) {
			created = cmd.ISBN
			return detail, nil
		}),
		UpdateUC: fakeUpdate(func(_ context.Context, cmd bookcmd.UpdateCommand) (*domainbook.BookDetail, error) {
			updated = cmd.ISBN
			return detail, nil
		}),
	}

	serve(t, h, http.MethodPost, "/api/books", `{"title":"a","price":1,"authorIds":[1],"status":1,"isbn":"`+isbn+`"}`)
	serve(t, h, http.MethodPut, "/api/books/1", `{"title":"a","price":1,"authorIds":[1],"status":1,"isbn":"`+isbn+`","version":1}`)

	// Normalizing and validating the ISBN is the domain's job, so the raw input is passed through.
	if created == nil || *created != isbn || updated == nil || *updated != isbn {
		t.Fatalf("create got %v, update got %v, want %s for both", created, updated, isbn)
	}

	rec := serve(t, h, http.MethodPost, "/api/books", `{"title":"a","price":1,"authorIds":[1],"status":1,"isbn":"`+strings.Repeat("9", 18)+`"}`)
	if rec.Code != http.StatusBadRequest || !reflect.DeepEqual(decode[common.ErrorResponse](t, rec).Errors, []common.FieldError{{Field: "isbn", Rule: "max"}}) {
		t.Fatalf("status = %d body = %s, want 400 isbn/max", rec.Code, rec.Body.String())
	}
}

type fakeList func(ctx context.Context, limit, offset int) (*domainbook.BookList, error)

func (f fakeList) Execute(ctx context.Context, limit, offset int) (*domainbook.BookList, error) {
	return f(ctx, limit, offset)
}
func TestHandlerList(t *testing.T) {
	list := &domainbook.BookList{
		Items: []*domainbook.BookListItem{{
			ID: 1, Title: "人間失格", Price: 1500, Status: 2,
			Authors:     []domainbook.BookListAuthor{{ID: 1, Name: "太宰治"}, {ID: 2, Name: "芥川龍之介"}},
			ISBN:        strPtr("9784873118703"),
			AmazonURL:   strPtr("https://www.amazon.co.jp/dp/4873118700"),
			CoverURL:    strPtr("https://thumbnail.image.rakuten.co.jp/1.jpg"),
			CoverSource: strPtr("rakuten"),
		}},
		Total: 21,
	}

	pages := []struct {
		name                  string
		query                 string
		wantLimit, wantOffset int
	}{
		{name: "クエリなしは既定の取得範囲で取得する", query: "", wantLimit: domaincommon.DefaultListLimit, wantOffset: 0},
		{name: "limitとoffsetを指定できる", query: "?limit=5&offset=10", wantLimit: 5, wantOffset: 10},
		{name: "limit上限ちょうどは有効", query: "?limit=100", wantLimit: 100, wantOffset: 0},
	}
	for _, tt := range pages {
		t.Run(tt.name, func(t *testing.T) {
			var gotLimit, gotOffset int
			h := &httpbook.Handler{ListUC: fakeList(func(_ context.Context, limit, offset int) (*domainbook.BookList, error) {
				gotLimit, gotOffset = limit, offset
				return list, nil
			})}
			rec := serve(t, h, http.MethodGet, "/api/books"+tt.query, "")

			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d, want 200 (body=%s)", rec.Code, rec.Body.String())
			}
			if gotLimit != tt.wantLimit || gotOffset != tt.wantOffset {
				t.Fatalf("usecase received limit=%d offset=%d, want %d/%d", gotLimit, gotOffset, tt.wantLimit, tt.wantOffset)
			}
			want := httpbook.ListResponse{
				Items: []httpbook.ListItemResponse{{
					ID: 1, Title: "人間失格", Price: 1500, Status: 2,
					Authors:     []httpbook.ListAuthorResponse{{ID: 1, Name: "太宰治"}, {ID: 2, Name: "芥川龍之介"}},
					ISBN:        strPtr("9784873118703"),
					AmazonURL:   strPtr("https://www.amazon.co.jp/dp/4873118700"),
					CoverURL:    strPtr("https://thumbnail.image.rakuten.co.jp/1.jpg"),
					CoverSource: strPtr("rakuten"),
				}},
				Total: 21, Limit: tt.wantLimit, Offset: tt.wantOffset,
			}
			if got := decode[httpbook.ListResponse](t, rec); !reflect.DeepEqual(got, want) {
				t.Fatalf("body = %+v, want %+v", got, want)
			}
		})
	}

	t.Run("0件は空配列で返す", func(t *testing.T) {
		h := &httpbook.Handler{ListUC: fakeList(func(context.Context, int, int) (*domainbook.BookList, error) {
			return &domainbook.BookList{Total: 0}, nil
		})}
		rec := serve(t, h, http.MethodGet, "/api/books", "")

		if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"items":[]`) {
			t.Fatalf("status = %d body = %s, want 200 with items []", rec.Code, rec.Body.String())
		}
	})

	invalid := []struct {
		name  string
		query string
		want  []common.FieldError
	}{
		{name: "limit 0 は400", query: "?limit=0", want: []common.FieldError{{Field: "limit", Rule: "min"}}},
		{name: "limit上限+1は400", query: "?limit=101", want: []common.FieldError{{Field: "limit", Rule: "max"}}},
		{name: "offset負は400", query: "?offset=-1", want: []common.FieldError{{Field: "offset", Rule: "min"}}},
		{name: "数値でないlimitは400", query: "?limit=abc", want: nil},
	}
	for _, tt := range invalid {
		t.Run(tt.name, func(t *testing.T) {
			h := &httpbook.Handler{ListUC: fakeList(func(context.Context, int, int) (*domainbook.BookList, error) {
				t.Error("usecase must not be called")
				return nil, nil
			})}
			rec := serve(t, h, http.MethodGet, "/api/books"+tt.query, "")

			if rec.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want 400", rec.Code)
			}
			if got := decode[common.ErrorResponse](t, rec).Errors; !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("errors = %+v, want %+v", got, tt.want)
			}
		})
	}
}
