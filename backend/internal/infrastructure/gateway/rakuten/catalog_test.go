package rakuten_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	domainbook "github.com/mrstsgk/book-management-system/backend/internal/domain/book"
	"github.com/mrstsgk/book-management-system/backend/internal/domain/common"
	"github.com/mrstsgk/book-management-system/backend/internal/infrastructure/gateway/rakuten"
)

// formatVersion=2 の BooksBook/Search の応答と同じ形。
const found = `{"Items":[{"title":"データ指向アプリケーションデザイン","author":"Martin Kleppmann/斉藤 太郎","publisherName":"オライリー・ジャパン",
"salesDate":"2019年07月18日頃","itemPrice":5060,"largeImageUrl":"https://thumbnail.image.rakuten.co.jp/0_mall/book/cabinet/8703/9784873118703.jpg?_ex=200x200",
"itemUrl":"https://books.rakuten.co.jp/rb/15949390/"}],"count":1}`

var fetchedAt = time.Date(2026, 9, 28, 12, 0, 0, 0, time.UTC)

func fixedNow() time.Time { return fetchedAt }

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
		got, err := rakuten.NewCatalog(srv.URL, "app-id", "access-key", srv.Client(), fixedNow).Lookup(context.Background(), isbn(t))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if q.Get("isbn") != "9784873118703" || q.Get("applicationId") != "app-id" || q.Get("accessKey") != "access-key" ||
			q.Get("format") != "json" || q.Get("formatVersion") != "2" {
			t.Fatalf("query = %v", *q)
		}
		b := got.Bibliography
		if b.Title() != "データ指向アプリケーションデザイン" || b.Authors() != "Martin Kleppmann/斉藤 太郎" || b.Publisher() != "オライリー・ジャパン" || b.PublishedOn() != "2019年07月18日頃" {
			t.Fatalf("got %+v", got)
		}
		if got.Cover == nil || got.Cover.Source() != domainbook.CoverSourceRakuten || !strings.HasPrefix(got.Cover.URL(), "https://thumbnail.image.rakuten.co.jp/") ||
			got.Cover.ProductURL() != "https://books.rakuten.co.jp/rb/15949390/" || !got.Cover.FetchedAt().Equal(fetchedAt) {
			t.Fatalf("Cover = %+v, want the rakuten image with its product page, fetched now", got.Cover)
		}
	})

	t.Run("商品ページが無ければ画像があっても書影なし", func(t *testing.T) {
		srv, _ := serve(t, http.StatusOK, `{"Items":[{"title":"x","largeImageUrl":"https://thumbnail.image.rakuten.co.jp/1.jpg","itemUrl":""}]}`)
		got, err := rakuten.NewCatalog(srv.URL, "a", "k", srv.Client(), fixedNow).Lookup(context.Background(), isbn(t))
		if err != nil || got.Cover != nil {
			t.Fatalf("got (%+v, %v), want no cover", got, err)
		}
	})

	t.Run("0件はErrNotFound", func(t *testing.T) {
		srv, _ := serve(t, http.StatusOK, `{"Items":[],"count":0}`)
		if _, err := rakuten.NewCatalog(srv.URL, "a", "k", srv.Client(), fixedNow).Lookup(context.Background(), isbn(t)); !errors.Is(err, common.ErrNotFound) {
			t.Fatalf("err = %v, want ErrNotFound", err)
		}
	})

	t.Run("画像が無ければ書影なし", func(t *testing.T) {
		srv, _ := serve(t, http.StatusOK, `{"Items":[{"title":"x","largeImageUrl":""}]}`)
		got, err := rakuten.NewCatalog(srv.URL, "a", "k", srv.Client(), fixedNow).Lookup(context.Background(), isbn(t))
		if err != nil || got.Cover != nil {
			t.Fatalf("got (%+v, %v), want no cover", got, err)
		}
	})

	t.Run("書名が空ならエラー", func(t *testing.T) {
		srv, _ := serve(t, http.StatusOK, `{"Items":[{"title":""}]}`)
		_, err := rakuten.NewCatalog(srv.URL, "a", "k", srv.Client(), fixedNow).Lookup(context.Background(), isbn(t))
		if err == nil || errors.Is(err, common.ErrNotFound) {
			t.Fatalf("err = %v, want a non-NotFound error", err)
		}
	})

	t.Run("認証エラーなど200以外はエラー", func(t *testing.T) {
		srv, _ := serve(t, http.StatusUnauthorized, `{"error":"wrong_parameter"}`)
		_, err := rakuten.NewCatalog(srv.URL, "a", "k", srv.Client(), fixedNow).Lookup(context.Background(), isbn(t))
		if err == nil || errors.Is(err, common.ErrNotFound) {
			t.Fatalf("err = %v, want a non-NotFound error", err)
		}
	})

	t.Run("通信エラーのメッセージにキーを含めない", func(t *testing.T) {
		srv, _ := serve(t, http.StatusOK, found)
		srv.Close() // connection refused
		_, err := rakuten.NewCatalog(srv.URL, "secret-app-id", "secret-access-key", http.DefaultClient, fixedNow).Lookup(context.Background(), isbn(t))
		if err == nil {
			t.Fatal("expected a connection error")
		}
		if strings.Contains(err.Error(), "secret-app-id") || strings.Contains(err.Error(), "secret-access-key") {
			t.Fatalf("error leaks the keys: %v", err)
		}
	})
}
