package public_test

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

	bookdomain "github.com/mrstsgk/book-management-system/backend/internal/book/domain"
	bookpublic "github.com/mrstsgk/book-management-system/backend/internal/book/presentation/http/public"
	bookqry "github.com/mrstsgk/book-management-system/backend/internal/book/usecase/query"
	"github.com/mrstsgk/book-management-system/backend/internal/domainerr"
	"github.com/mrstsgk/book-management-system/backend/internal/platform/httpserver"
)

const testSession = "test-session"

type fakeGet func(context.Context, int64) (*bookdomain.BookDetail, error)

func (f fakeGet) Execute(ctx context.Context, id int64) (*bookdomain.BookDetail, error) {
	return f(ctx, id)
}

type fakeList func(context.Context, bookqry.SearchBooksInput) (*bookdomain.BookList, error)

func (f fakeList) Execute(ctx context.Context, in bookqry.SearchBooksInput) (*bookdomain.BookList, error) {
	return f(ctx, in)
}

// serve は cmd/api/main.go と同じ形（NewEcho + Register）でハンドラを組み立ててリクエストを流す。
func serve(t *testing.T, h *bookpublic.Handler, method, path, body string, withSession bool) *httptest.ResponseRecorder {
	t.Helper()
	orig := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, nil)))
	t.Cleanup(func() { slog.SetDefault(orig) })

	e := httpserver.NewEcho()
	h.Register(e.Group("/api/books"))
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	if withSession {
		req.AddCookie(&http.Cookie{Name: httpserver.SessionCookieName, Value: testSession})
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

var detail = &bookdomain.BookDetail{
	ID: 1, ISBN: "9784873118703", Title: "データ指向アプリケーションデザイン", Authors: "Kleppmann,Martin",
	Publisher: "オーム社", PublishedOn: "201907", AmazonURL: strPtr("https://www.amazon.co.jp/dp/4873118700"),
	CoverURL: strPtr("https://cover.openbd.jp/9784873118703.jpg"), CoverSource: strPtr("openbd"),
	Summary: "分散データの設計を学べる", Tags: []string{"データベース"}, Comment: "良書\n2行目", Rating: 5, Version: 1,
}

var detailResponse = bookpublic.Response{
	ID: 1, ISBN: "9784873118703", Title: "データ指向アプリケーションデザイン", Authors: "Kleppmann,Martin",
	Publisher: "オーム社", PublishedOn: "201907", AmazonURL: strPtr("https://www.amazon.co.jp/dp/4873118700"),
	CoverURL: strPtr("https://cover.openbd.jp/9784873118703.jpg"), CoverSource: strPtr("openbd"),
	Summary: "分散データの設計を学べる", Tags: []string{"データベース"}, Comment: "良書\n2行目", Rating: 5, Version: 1,
}

func mustNotCall(t *testing.T) func() {
	return func() { t.Error("usecase must not be called") }
}

func TestHandlerList(t *testing.T) {
	list := &bookdomain.BookList{Total: 21, Items: []*bookdomain.BookListItem{{
		ID: 1, ISBN: "9784873118703", Title: "データ指向アプリケーションデザイン", Summary: "分散データの設計を学べる", Tags: []string{"データベース"}, Authors: "Kleppmann,Martin",
		AmazonURL: strPtr("https://www.amazon.co.jp/dp/4873118700"), CoverURL: strPtr("https://cover.openbd.jp/9784873118703.jpg"), CoverSource: strPtr("openbd"), Rating: 5,
	}}}

	for _, tt := range []struct {
		name                  string
		query                 string
		wantLimit, wantOffset int
	}{
		{name: "クエリなしは既定の取得範囲", query: "", wantLimit: bookdomain.DefaultListLimit},
		{name: "limitとoffsetを指定できる", query: "?limit=5&offset=10", wantLimit: 5, wantOffset: 10},
	} {
		t.Run(tt.name+"で誰でも取得できる", func(t *testing.T) {
			var gotLimit, gotOffset int
			h := &bookpublic.Handler{SearchBooksUsecase: fakeList(func(_ context.Context, in bookqry.SearchBooksInput) (*bookdomain.BookList, error) {
				gotLimit, gotOffset = in.Limit, in.Offset
				return list, nil
			})}
			rec := serve(t, h, http.MethodGet, "/api/books"+tt.query, "", false)

			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d, want 200 (body=%s)", rec.Code, rec.Body.String())
			}
			if gotLimit != tt.wantLimit || gotOffset != tt.wantOffset {
				t.Fatalf("usecase received limit=%d offset=%d", gotLimit, gotOffset)
			}
			want := bookpublic.ListResponse{Total: 21, Limit: tt.wantLimit, Offset: tt.wantOffset, Items: []bookpublic.ListItemResponse{{
				ID: 1, ISBN: "9784873118703", Title: "データ指向アプリケーションデザイン", Summary: "分散データの設計を学べる", Tags: []string{"データベース"}, Authors: "Kleppmann,Martin",
				AmazonURL: strPtr("https://www.amazon.co.jp/dp/4873118700"), CoverURL: strPtr("https://cover.openbd.jp/9784873118703.jpg"), CoverSource: strPtr("openbd"), Rating: 5,
			}}}
			if got := decode[bookpublic.ListResponse](t, rec); !reflect.DeepEqual(got, want) {
				t.Fatalf("body = %+v, want %+v", got, want)
			}
		})
	}

	t.Run("0件は空配列", func(t *testing.T) {
		h := &bookpublic.Handler{SearchBooksUsecase: fakeList(func(context.Context, bookqry.SearchBooksInput) (*bookdomain.BookList, error) {
			return &bookdomain.BookList{}, nil
		})}
		rec := serve(t, h, http.MethodGet, "/api/books", "", false)
		if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"items":[]`) {
			t.Fatalf("status = %d body = %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("キーワードと分野タグをusecaseに渡す", func(t *testing.T) {
		var got bookqry.SearchBooksInput
		h := &bookpublic.Handler{SearchBooksUsecase: fakeList(func(_ context.Context, in bookqry.SearchBooksInput) (*bookdomain.BookList, error) {
			got = in
			return &bookdomain.BookList{}, nil
		})}
		rec := serve(t, h, http.MethodGet, "/api/books?q=%E8%A8%AD%E8%A8%88&tagId=3", "", false)
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200 (body=%s)", rec.Code, rec.Body.String())
		}
		if got.Keyword != "設計" || got.TagID == nil || *got.TagID != 3 {
			t.Fatalf("usecase received %+v, want Keyword=設計 TagID=3", got)
		}
	})

	t.Run("前後の空白を除いて100文字のキーワードは200", func(t *testing.T) {
		var got bookqry.SearchBooksInput
		h := &bookpublic.Handler{SearchBooksUsecase: fakeList(func(_ context.Context, in bookqry.SearchBooksInput) (*bookdomain.BookList, error) {
			got = in
			return &bookdomain.BookList{}, nil
		})}
		kw := strings.Repeat("a", 100)
		rec := serve(t, h, http.MethodGet, "/api/books?q=%20%20"+kw+"%20%20", "", false)
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200 (body=%s)", rec.Code, rec.Body.String())
		}
		if got.Keyword != "  "+kw+"  " {
			t.Fatalf("usecase received %q, want the raw keyword (trimming is the domain's job)", got.Keyword)
		}
	})

	t.Run("前後の空白を除いて101文字のキーワードは400", func(t *testing.T) {
		h := &bookpublic.Handler{SearchBooksUsecase: fakeList(func(context.Context, bookqry.SearchBooksInput) (*bookdomain.BookList, error) {
			return nil, domainerr.ErrInvalid
		})}
		rec := serve(t, h, http.MethodGet, "/api/books?q=%20"+strings.Repeat("a", 101)+"%20", "", false)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want 400", rec.Code)
		}
	})

	t.Run("キーワードと分野タグを省略すると絞り込まない", func(t *testing.T) {
		var got bookqry.SearchBooksInput
		h := &bookpublic.Handler{SearchBooksUsecase: fakeList(func(_ context.Context, in bookqry.SearchBooksInput) (*bookdomain.BookList, error) {
			got = in
			return &bookdomain.BookList{}, nil
		})}
		if rec := serve(t, h, http.MethodGet, "/api/books", "", false); rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200", rec.Code)
		}
		if got.Keyword != "" || got.TagID != nil {
			t.Fatalf("usecase received %+v, want no keyword and no tag", got)
		}
	})

	for _, tt := range []struct {
		name  string
		query string
	}{
		{name: "limit上限+1は400", query: "?limit=101"},
		{name: "分野タグID 0は400", query: "?tagId=0"},
		{name: "分野タグIDが数値でなければ400", query: "?tagId=abc"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			called := mustNotCall(t)
			h := &bookpublic.Handler{SearchBooksUsecase: fakeList(func(context.Context, bookqry.SearchBooksInput) (*bookdomain.BookList, error) {
				called()
				return nil, nil
			})}
			rec := serve(t, h, http.MethodGet, "/api/books"+tt.query, "", false)
			if rec.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want 400", rec.Code)
			}
		})
	}
}

func TestHandlerGet(t *testing.T) {
	t.Run("誰でも詳細を取得できる", func(t *testing.T) {
		h := &bookpublic.Handler{ReadBookUsecase: fakeGet(func(_ context.Context, id int64) (*bookdomain.BookDetail, error) {
			if id != 1 {
				t.Errorf("id = %d, want 1", id)
			}
			return detail, nil
		})}
		rec := serve(t, h, http.MethodGet, "/api/books/1", "", false)
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200", rec.Code)
		}
		if got := decode[bookpublic.Response](t, rec); !reflect.DeepEqual(got, detailResponse) {
			t.Fatalf("body = %+v, want %+v", got, detailResponse)
		}
	})

	t.Run("存在しない本は404", func(t *testing.T) {
		h := &bookpublic.Handler{ReadBookUsecase: fakeGet(func(context.Context, int64) (*bookdomain.BookDetail, error) { return nil, domainerr.ErrNotFound })}
		if rec := serve(t, h, http.MethodGet, "/api/books/1", "", false); rec.Code != http.StatusNotFound {
			t.Fatalf("status = %d, want 404", rec.Code)
		}
	})
}
