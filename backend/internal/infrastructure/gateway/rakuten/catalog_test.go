package rakuten_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	domainbook "github.com/mrstsgk/book-management-system/backend/internal/domain/book"
	"github.com/mrstsgk/book-management-system/backend/internal/domain/common"
	"github.com/mrstsgk/book-management-system/backend/internal/infrastructure/gateway/rakuten"
)

// Shaped like a formatVersion=2 BooksBook/Search response.
const found = `{"Items":[{"title":"データ指向アプリケーションデザイン","author":"Martin Kleppmann/斉藤 太郎","publisherName":"オライリー・ジャパン",
"salesDate":"2019年07月18日頃","itemPrice":5060,"largeImageUrl":"https://thumbnail.image.rakuten.co.jp/0_mall/book/cabinet/8703/9784873118703.jpg?_ex=200x200"}],"count":1}`

func serve(t *testing.T, status int, body string) (*httptest.Server, *url.Values) {
	t.Helper()
	got := &url.Values{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/services/api/BooksBook/Search/20170404" {
			http.NotFound(w, r)
			return
		}
		*got = r.URL.Query()
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	return srv, got
}

func isbn(t *testing.T) domainbook.ISBN {
	t.Helper()
	v, err := domainbook.NewISBN("9784873118703")
	if err != nil {
		t.Fatal(err)
	}
	return v
}

func TestCatalog_Lookup(t *testing.T) {
	t.Run("ISBNとキーを付けて検索し、書誌と書影を返す", func(t *testing.T) {
		srv, q := serve(t, http.StatusOK, found)
		got, err := rakuten.NewCatalog(srv.URL, "app-id", "access-key", srv.Client()).Lookup(context.Background(), isbn(t))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if q.Get("isbn") != "9784873118703" || q.Get("applicationId") != "app-id" || q.Get("accessKey") != "access-key" ||
			q.Get("format") != "json" || q.Get("formatVersion") != "2" {
			t.Fatalf("query = %v", *q)
		}
		if got.Title != "データ指向アプリケーションデザイン" || got.Publisher != "オライリー・ジャパン" || got.PublishedOn != "2019年07月18日頃" {
			t.Fatalf("got %+v", got)
		}
		if got.Price == nil || *got.Price != 5060 {
			t.Fatalf("Price = %v, want 5060", got.Price)
		}
		if got.Cover == nil || got.Cover.Source() != domainbook.CoverSourceRakuten || !strings.HasPrefix(got.Cover.URL(), "https://thumbnail.image.rakuten.co.jp/") {
			t.Fatalf("Cover = %+v", got.Cover)
		}
	})

	t.Run("0件はErrNotFound", func(t *testing.T) {
		srv, _ := serve(t, http.StatusOK, `{"Items":[],"count":0}`)
		if _, err := rakuten.NewCatalog(srv.URL, "a", "k", srv.Client()).Lookup(context.Background(), isbn(t)); !errors.Is(err, common.ErrNotFound) {
			t.Fatalf("err = %v, want ErrNotFound", err)
		}
	})

	t.Run("画像も価格も無ければ書影なし・価格なし", func(t *testing.T) {
		srv, _ := serve(t, http.StatusOK, `{"Items":[{"title":"x","itemPrice":0,"largeImageUrl":""}]}`)
		got, err := rakuten.NewCatalog(srv.URL, "a", "k", srv.Client()).Lookup(context.Background(), isbn(t))
		if err != nil || got.Cover != nil || got.Price != nil {
			t.Fatalf("got (%+v, %v), want no cover and no price", got, err)
		}
	})

	t.Run("認証エラーなど200以外はエラー", func(t *testing.T) {
		srv, _ := serve(t, http.StatusUnauthorized, `{"error":"wrong_parameter"}`)
		_, err := rakuten.NewCatalog(srv.URL, "a", "k", srv.Client()).Lookup(context.Background(), isbn(t))
		if err == nil || errors.Is(err, common.ErrNotFound) {
			t.Fatalf("err = %v, want a non-NotFound error", err)
		}
	})

	t.Run("通信エラーのメッセージにキーを含めない", func(t *testing.T) {
		srv, _ := serve(t, http.StatusOK, found)
		srv.Close() // connection refused
		_, err := rakuten.NewCatalog(srv.URL, "secret-app-id", "secret-access-key", http.DefaultClient).Lookup(context.Background(), isbn(t))
		if err == nil {
			t.Fatal("expected a connection error")
		}
		if strings.Contains(err.Error(), "secret-app-id") || strings.Contains(err.Error(), "secret-access-key") {
			t.Fatalf("error leaks the keys: %v", err)
		}
	})
}
