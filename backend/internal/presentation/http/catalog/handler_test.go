package catalog_test

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/labstack/echo/v4"

	domainbook "github.com/mrstsgk/book-management-system/backend/internal/domain/book"
	domaincommon "github.com/mrstsgk/book-management-system/backend/internal/domain/common"
	httpcatalog "github.com/mrstsgk/book-management-system/backend/internal/presentation/http/catalog"
	"github.com/mrstsgk/book-management-system/backend/internal/presentation/http/common"
)

const adminToken = "test-admin-token"

type fakeLookup func(context.Context, string) (*domainbook.CatalogEntry, error)

func (f fakeLookup) Execute(ctx context.Context, isbn string) (*domainbook.CatalogEntry, error) {
	return f(ctx, isbn)
}

// serve は cmd/api/main.go と同じ形（NewEcho + Register）でハンドラを組み立ててリクエストを流す。
func serve(t *testing.T, h *httpcatalog.Handler, path string, withToken bool) *httptest.ResponseRecorder {
	t.Helper()
	orig := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, nil)))
	t.Cleanup(func() { slog.SetDefault(orig) })

	h.AdminOnly = common.RequireAdminToken(adminToken)
	e := common.NewEcho()
	h.Register(e.Group("/api/catalog"))
	req := httptest.NewRequest(http.MethodGet, path, nil)
	if withToken {
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+adminToken)
	}
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}

func entry(t *testing.T, isbn string, cover *domainbook.Cover) *domainbook.CatalogEntry {
	t.Helper()
	v, err := domainbook.NewISBN(isbn)
	if err != nil {
		t.Fatal(err)
	}
	bib, err := domainbook.NewBibliography("データ指向アプリケーションデザイン", "Kleppmann,Martin", "オーム社", "201907")
	if err != nil {
		t.Fatal(err)
	}
	return &domainbook.CatalogEntry{ISBN: v, Bibliography: bib, Cover: cover}
}

func strPtr(s string) *string { return &s }

func TestHandlerLookup(t *testing.T) {
	t.Run("書誌・書影・Amazonリンクを返し、入力はそのままusecaseに渡す", func(t *testing.T) {
		cover, err := domainbook.NewCover("https://cover.openbd.jp/9784873118703.jpg", domainbook.CoverSourceOpenBD)
		if err != nil {
			t.Fatal(err)
		}
		var gotISBN string
		h := &httpcatalog.Handler{LookupUC: fakeLookup(func(_ context.Context, isbn string) (*domainbook.CatalogEntry, error) {
			gotISBN = isbn
			return entry(t, "9784873118703", &cover), nil
		})}
		rec := serve(t, h, "/api/catalog/978-4-87311-870-3", true)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200 (body=%s)", rec.Code, rec.Body.String())
		}
		if gotISBN != "978-4-87311-870-3" {
			t.Fatalf("usecase received %q", gotISBN)
		}
		var got httpcatalog.Response
		if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
			t.Fatal(err)
		}
		want := httpcatalog.Response{
			ISBN: "9784873118703", Title: "データ指向アプリケーションデザイン", Authors: "Kleppmann,Martin",
			Publisher: "オーム社", PublishedOn: "201907", AmazonURL: strPtr("https://www.amazon.co.jp/dp/4873118700"),
			CoverURL: strPtr("https://cover.openbd.jp/9784873118703.jpg"), CoverSource: strPtr("openbd"),
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("body = %+v\nwant %+v", got, want)
		}
	})

	t.Run("Google Booksの書影なら本のページも返す", func(t *testing.T) {
		cover, err := domainbook.NewGoogleBooksCover("https://books.google.com/books/content?id=a", "https://books.google.co.jp/books?id=a")
		if err != nil {
			t.Fatal(err)
		}
		h := &httpcatalog.Handler{LookupUC: fakeLookup(func(context.Context, string) (*domainbook.CatalogEntry, error) {
			return entry(t, "9784873118703", &cover), nil
		})}
		var got httpcatalog.Response
		if err := json.Unmarshal(serve(t, h, "/api/catalog/9784873118703", true).Body.Bytes(), &got); err != nil {
			t.Fatal(err)
		}
		if got.CoverSource == nil || *got.CoverSource != "googlebooks" || got.CoverPageURL == nil || *got.CoverPageURL != "https://books.google.co.jp/books?id=a" {
			t.Fatalf("got %+v, want the Google Books cover with its page", got)
		}
	})

	t.Run("書影が無く979のISBNなら書影もAmazonリンクもnull", func(t *testing.T) {
		h := &httpcatalog.Handler{LookupUC: fakeLookup(func(context.Context, string) (*domainbook.CatalogEntry, error) {
			return entry(t, "9791032305690", nil), nil
		})}
		var got httpcatalog.Response
		if err := json.Unmarshal(serve(t, h, "/api/catalog/9791032305690", true).Body.Bytes(), &got); err != nil {
			t.Fatal(err)
		}
		if got.CoverURL != nil || got.CoverSource != nil || got.CoverPageURL != nil || got.AmazonURL != nil {
			t.Fatalf("got %+v", got)
		}
	})

	for _, tt := range []struct {
		name      string
		path      string
		withToken bool
		err       error
		want      int
	}{
		{name: "トークンが無ければ401", path: "/api/catalog/9784873118703", withToken: false, want: http.StatusUnauthorized},
		{name: "18文字以上のISBNは400", path: "/api/catalog/978-4-87311-870-3-0", withToken: true, want: http.StatusBadRequest},
		{name: "不正なISBNは400", path: "/api/catalog/123", withToken: true, err: domaincommon.ErrInvalid, want: http.StatusBadRequest},
		{name: "該当なしは404", path: "/api/catalog/9784873118703", withToken: true, err: domaincommon.ErrNotFound, want: http.StatusNotFound},
	} {
		t.Run(tt.name, func(t *testing.T) {
			h := &httpcatalog.Handler{LookupUC: fakeLookup(func(context.Context, string) (*domainbook.CatalogEntry, error) {
				if tt.err == nil {
					t.Error("usecase must not be called")
				}
				return nil, tt.err
			})}
			if rec := serve(t, h, tt.path, tt.withToken); rec.Code != tt.want {
				t.Fatalf("status = %d, want %d", rec.Code, tt.want)
			}
		})
	}
}
