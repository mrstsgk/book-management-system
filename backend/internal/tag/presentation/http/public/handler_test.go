package public_test

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

	"github.com/mrstsgk/book-management-system/backend/internal/platform/httpserver"
	tagdomain "github.com/mrstsgk/book-management-system/backend/internal/tag/domain"
	tagpublic "github.com/mrstsgk/book-management-system/backend/internal/tag/presentation/http/public"
)

const testSession = "test-session"

type fakeList func(context.Context) (*tagdomain.TagList, error)

func (f fakeList) Execute(ctx context.Context) (*tagdomain.TagList, error) { return f(ctx) }

type fakeCountBooks func(context.Context) (*tagdomain.TagBookCounts, error)

func (f fakeCountBooks) Execute(ctx context.Context) (*tagdomain.TagBookCounts, error) { return f(ctx) }

func TestHandlerCountBooks(t *testing.T) {
	t.Run("ログイン無しでも分野タグごとの冊数を返す", func(t *testing.T) {
		counts := &tagdomain.TagBookCounts{Items: []*tagdomain.TagBookCount{{ID: 1, Name: "設計", BookCount: 3}}}
		h := &tagpublic.Handler{CountBooksUsecase: fakeCountBooks(func(context.Context) (*tagdomain.TagBookCounts, error) { return counts, nil })}

		rec := serve(t, h, http.MethodGet, "/api/tags/counts", "", false)
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200 (body=%s)", rec.Code, rec.Body.String())
		}
		var got tagpublic.BookCountListResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
			t.Fatal(err)
		}
		want := tagpublic.BookCountListResponse{Items: []tagpublic.BookCountResponse{{ID: 1, Name: "設計", BookCount: 3}}}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("body = %+v, want %+v", got, want)
		}
	})

	t.Run("0件は空配列", func(t *testing.T) {
		h := &tagpublic.Handler{CountBooksUsecase: fakeCountBooks(func(context.Context) (*tagdomain.TagBookCounts, error) { return &tagdomain.TagBookCounts{}, nil })}
		rec := serve(t, h, http.MethodGet, "/api/tags/counts", "", false)
		if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"items":[]`) {
			t.Fatalf("status = %d body = %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("ユースケースの障害は500", func(t *testing.T) {
		h := &tagpublic.Handler{CountBooksUsecase: fakeCountBooks(func(context.Context) (*tagdomain.TagBookCounts, error) { return nil, errors.New("db: timeout") })}
		if rec := serve(t, h, http.MethodGet, "/api/tags/counts", "", false); rec.Code != http.StatusInternalServerError {
			t.Fatalf("status = %d, want 500", rec.Code)
		}
	})

}

// serve は cmd/api/main.go と同じ形（NewEcho + Register）でハンドラを組み立ててリクエストを流す。
func serve(t *testing.T, h *tagpublic.Handler, method, path, body string, withSession bool) *httptest.ResponseRecorder {
	t.Helper()
	orig := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, nil)))
	t.Cleanup(func() { slog.SetDefault(orig) })

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

func TestHandlerList(t *testing.T) {
	list := &tagdomain.TagList{Items: []*tagdomain.TagListItem{{ID: 1, Name: "データベース", Version: 1}}}
	h := &tagpublic.Handler{SearchTagsUsecase: fakeList(func(context.Context) (*tagdomain.TagList, error) { return list, nil })}

	rec := serve(t, h, http.MethodGet, "/api/tags", "", false)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (認証不要)", rec.Code)
	}
	var got tagpublic.ListResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	want := tagpublic.ListResponse{Items: []tagpublic.Response{{ID: 1, Name: "データベース", Version: 1}}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("body = %+v, want %+v", got, want)
	}
}
