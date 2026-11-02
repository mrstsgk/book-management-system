package admin_test

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	bookdomain "github.com/mrstsgk/book-management-system/backend/internal/book/domain"
	bookadmin "github.com/mrstsgk/book-management-system/backend/internal/book/presentation/http/admin"
	"github.com/mrstsgk/book-management-system/backend/internal/domainerr"
	"github.com/mrstsgk/book-management-system/backend/internal/platform/httpserver"
)

type fakeLookup func(context.Context, string) (*bookdomain.CatalogEntry, error)

func (f fakeLookup) Execute(ctx context.Context, isbn string) (*bookdomain.CatalogEntry, error) {
	return f(ctx, isbn)
}

// serveCatalog は cmd/api/main.go と同じ形（NewEcho + Register）でハンドラを組み立ててリクエストを流す。
func serveCatalog(t *testing.T, h *bookadmin.CatalogHandler, path string, withSession bool) *httptest.ResponseRecorder {
	t.Helper()
	orig := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, nil)))
	t.Cleanup(func() { slog.SetDefault(orig) })

	h.AdminOnly = fakeAdminOnly
	e := httpserver.NewEcho()
	h.Register(e.Group("/api/catalog"))
	req := httptest.NewRequest(http.MethodGet, path, nil)
	if withSession {
		req.AddCookie(&http.Cookie{Name: httpserver.SessionCookieName, Value: testSession})
	}
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}

func entry(t *testing.T, isbn string, cover *bookdomain.Cover) *bookdomain.CatalogEntry {
	t.Helper()
	v, err := bookdomain.NewISBN(isbn)
	if err != nil {
		t.Fatal(err)
	}
	bib, err := bookdomain.NewBibliography("データ指向アプリケーションデザイン", "Kleppmann,Martin", "オーム社", "201907")
	if err != nil {
		t.Fatal(err)
	}
	return &bookdomain.CatalogEntry{ISBN: v, Bibliography: bib, Cover: cover}
}

func TestHandlerLookup(t *testing.T) {
	t.Run("書誌・書影・Amazonリンクを返し、入力はそのままusecaseに渡す", func(t *testing.T) {
		cover, err := bookdomain.NewCover("https://cover.openbd.jp/9784873118703.jpg", bookdomain.CoverSourceOpenBD)
		if err != nil {
			t.Fatal(err)
		}
		var gotISBN string
		h := &bookadmin.CatalogHandler{LookupCatalogUsecase: fakeLookup(func(_ context.Context, isbn string) (*bookdomain.CatalogEntry, error) {
			gotISBN = isbn
			return entry(t, "9784873118703", &cover), nil
		})}
		rec := serveCatalog(t, h, "/api/catalog/978-4-87311-870-3", true)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200 (body=%s)", rec.Code, rec.Body.String())
		}
		if gotISBN != "978-4-87311-870-3" {
			t.Fatalf("usecase received %q", gotISBN)
		}
		var got bookadmin.CatalogResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
			t.Fatal(err)
		}
		want := bookadmin.CatalogResponse{
			ISBN: "9784873118703", Title: "データ指向アプリケーションデザイン", Authors: "Kleppmann,Martin",
			Publisher: "オーム社", PublishedOn: "201907", AmazonURL: new("https://www.amazon.co.jp/dp/4873118700"),
			CoverURL: new("https://cover.openbd.jp/9784873118703.jpg"), CoverSource: new("openbd"),
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("body = %+v\nwant %+v", got, want)
		}
	})

	t.Run("書影が無く979のISBNなら書影もAmazonリンクもnull", func(t *testing.T) {
		h := &bookadmin.CatalogHandler{LookupCatalogUsecase: fakeLookup(func(context.Context, string) (*bookdomain.CatalogEntry, error) {
			return entry(t, "9791032305690", nil), nil
		})}
		var got bookadmin.CatalogResponse
		if err := json.Unmarshal(serveCatalog(t, h, "/api/catalog/9791032305690", true).Body.Bytes(), &got); err != nil {
			t.Fatal(err)
		}
		if got.CoverURL != nil || got.CoverSource != nil || got.AmazonURL != nil {
			t.Fatalf("got %+v", got)
		}
	})

	for _, tt := range []struct {
		name        string
		path        string
		withSession bool
		err         error
		want        int
	}{
		{name: "ログインしていなければ401", path: "/api/catalog/9784873118703", withSession: false, want: http.StatusUnauthorized},
		{name: "18文字以上のISBNは400", path: "/api/catalog/978-4-87311-870-3-0", withSession: true, want: http.StatusBadRequest},
		{name: "不正なISBNは400", path: "/api/catalog/123", withSession: true, err: domainerr.ErrInvalid, want: http.StatusBadRequest},
		{name: "該当なしは404", path: "/api/catalog/9784873118703", withSession: true, err: domainerr.ErrNotFound, want: http.StatusNotFound},
	} {
		t.Run(tt.name, func(t *testing.T) {
			h := &bookadmin.CatalogHandler{LookupCatalogUsecase: fakeLookup(func(context.Context, string) (*bookdomain.CatalogEntry, error) {
				if tt.err == nil {
					t.Error("usecase must not be called")
				}
				return nil, tt.err
			})}
			if rec := serveCatalog(t, h, tt.path, tt.withSession); rec.Code != tt.want {
				t.Fatalf("status = %d, want %d", rec.Code, tt.want)
			}
		})
	}
}
