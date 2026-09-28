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

	domainbook "github.com/mrstsgk/book-management-system/backend/internal/domain/book"
	domaincommon "github.com/mrstsgk/book-management-system/backend/internal/domain/common"
	httpcatalog "github.com/mrstsgk/book-management-system/backend/internal/presentation/http/catalog"
	"github.com/mrstsgk/book-management-system/backend/internal/presentation/http/common"
)

type fakeLookup func(ctx context.Context, isbn string) (*domainbook.CatalogEntry, error)

func (f fakeLookup) Execute(ctx context.Context, isbn string) (*domainbook.CatalogEntry, error) {
	return f(ctx, isbn)
}

// serve wires the handler the same way cmd/api/main.go does (NewEcho + Register).
func serve(t *testing.T, h *httpcatalog.Handler, path string) *httptest.ResponseRecorder {
	t.Helper()
	orig := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, nil)))
	t.Cleanup(func() { slog.SetDefault(orig) })

	e := common.NewEcho()
	h.Register(e.Group("/api/catalog"))
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
	return rec
}

func entry(t *testing.T, isbn string, cover *domainbook.Cover) *domainbook.CatalogEntry {
	t.Helper()
	v, err := domainbook.NewISBN(isbn)
	if err != nil {
		t.Fatal(err)
	}
	price := int64(4600)
	return &domainbook.CatalogEntry{
		ISBN: v, Title: "データ指向アプリケーションデザイン", Authors: "Kleppmann,Martin 斉藤,太郎",
		Publisher: "オーム社", PublishedOn: "201907", Price: &price, Cover: cover,
	}
}

func strPtr(s string) *string { return &s }

func TestHandlerLookup(t *testing.T) {
	t.Run("書誌・書影・Amazonリンクを返す", func(t *testing.T) {
		cover, err := domainbook.NewCover("https://thumbnail.image.rakuten.co.jp/1.jpg", domainbook.CoverSourceRakuten)
		if err != nil {
			t.Fatal(err)
		}
		var gotISBN string
		h := &httpcatalog.Handler{LookupUC: fakeLookup(func(_ context.Context, isbn string) (*domainbook.CatalogEntry, error) {
			gotISBN = isbn
			return entry(t, "9784873118703", &cover), nil
		})}
		rec := serve(t, h, "/api/catalog/978-4-87311-870-3")

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200 (body=%s)", rec.Code, rec.Body.String())
		}
		if gotISBN != "978-4-87311-870-3" {
			t.Fatalf("usecase received %q; the raw path value must be passed through", gotISBN)
		}
		var got httpcatalog.Response
		if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
			t.Fatal(err)
		}
		price := int64(4600)
		want := httpcatalog.Response{
			ISBN: "9784873118703", Title: "データ指向アプリケーションデザイン", Authors: "Kleppmann,Martin 斉藤,太郎",
			Publisher: "オーム社", PublishedOn: "201907", Price: &price,
			AmazonURL: strPtr("https://www.amazon.co.jp/dp/4873118700"),
			CoverURL:  strPtr("https://thumbnail.image.rakuten.co.jp/1.jpg"), CoverSource: strPtr("rakuten"),
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("body = %+v\nwant %+v", got, want)
		}
	})

	t.Run("書影が無く979のISBNなら書影もAmazonリンクもnull", func(t *testing.T) {
		h := &httpcatalog.Handler{LookupUC: fakeLookup(func(context.Context, string) (*domainbook.CatalogEntry, error) {
			return entry(t, "9791032305690", nil), nil
		})}
		rec := serve(t, h, "/api/catalog/9791032305690")

		var got httpcatalog.Response
		if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
			t.Fatal(err)
		}
		if got.CoverURL != nil || got.CoverSource != nil || got.AmazonURL != nil {
			t.Fatalf("got %+v, want no cover and no Amazon link", got)
		}
	})

	statuses := []struct {
		name string
		err  error
		want int
	}{
		{name: "不正なISBNは400", err: domaincommon.ErrInvalid, want: http.StatusBadRequest},
		{name: "該当なしは404", err: domaincommon.ErrNotFound, want: http.StatusNotFound},
	}
	for _, tt := range statuses {
		t.Run(tt.name, func(t *testing.T) {
			h := &httpcatalog.Handler{LookupUC: fakeLookup(func(context.Context, string) (*domainbook.CatalogEntry, error) {
				return nil, tt.err
			})}
			if rec := serve(t, h, "/api/catalog/123"); rec.Code != tt.want {
				t.Fatalf("status = %d, want %d", rec.Code, tt.want)
			}
		})
	}
}

func TestHandlerLookup_TooLongISBNIs400WithoutCallingTheUsecase(t *testing.T) {
	h := &httpcatalog.Handler{LookupUC: fakeLookup(func(context.Context, string) (*domainbook.CatalogEntry, error) {
		t.Error("usecase must not be called")
		return nil, nil
	})}
	rec := serve(t, h, "/api/catalog/978-4-87311-870-3-0")

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
	var got common.ErrorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if want := []common.FieldError{{Field: "isbn", Rule: "max"}}; !reflect.DeepEqual(got.Errors, want) {
		t.Fatalf("errors = %+v, want %+v", got.Errors, want)
	}
}

func TestHandlerLookup_17CharISBNReachesTheUsecase(t *testing.T) {
	called := false
	h := &httpcatalog.Handler{LookupUC: fakeLookup(func(context.Context, string) (*domainbook.CatalogEntry, error) {
		called = true
		return nil, domaincommon.ErrNotFound
	})}
	serve(t, h, "/api/catalog/978-4-87311-870-3")

	if !called {
		t.Fatal("a 17-character ISBN with hyphens must reach the usecase")
	}
}
