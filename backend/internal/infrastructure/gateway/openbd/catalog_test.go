package openbd_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	domainbook "github.com/mrstsgk/book-management-system/backend/internal/domain/book"
	"github.com/mrstsgk/book-management-system/backend/internal/domain/common"
	"github.com/mrstsgk/book-management-system/backend/internal/infrastructure/gateway/openbd"
)

// Trimmed from a real /v1/get response (ISBN 9784480434623, which has a cover).
const withCover = `[{"onix":{"ProductSupply":{"SupplyDetail":{"Price":[{"PriceType":"03","PriceAmount":"950","CurrencyCode":"JPY"}]}}},
"summary":{"isbn":"9784480434623","title":"高峰秀子の捨てられない荷物","publisher":"筑摩書房","pubdate":"20170807",
"cover":"https://cover.openbd.jp/9784480434623.jpg","author":"斎藤明美／著"}}]`

// Trimmed from a real response without a cover (ISBN 9784873118703).
const withoutCover = `[{"onix":{"ProductSupply":{"SupplyDetail":{"Price":[{"PriceAmount":"4600"}]}}},
"summary":{"isbn":"9784873118703","title":"データ指向アプリケーションデザイン","publisher":"オーム社","pubdate":"201907",
"cover":"","author":"Kleppmann,Martin 斉藤,太郎,pub.2019 玉川,竜司,1964-"}}]`

func serve(t *testing.T, status int, body string) (*httptest.Server, *string) {
	t.Helper()
	var gotQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/get" {
			http.NotFound(w, r)
			return
		}
		gotQuery = r.URL.Query().Get("isbn")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	return srv, &gotQuery
}

func isbn(t *testing.T, s string) domainbook.ISBN {
	t.Helper()
	v, err := domainbook.NewISBN(s)
	if err != nil {
		t.Fatal(err)
	}
	return v
}

func TestCatalog_Lookup(t *testing.T) {
	t.Run("書誌と書影と価格を返す", func(t *testing.T) {
		srv, gotISBN := serve(t, http.StatusOK, withCover)
		got, err := openbd.NewCatalog(srv.URL, srv.Client()).Lookup(context.Background(), isbn(t, "9784480434623"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if *gotISBN != "9784480434623" {
			t.Fatalf("requested isbn=%q", *gotISBN)
		}
		if got.Title != "高峰秀子の捨てられない荷物" || got.Authors != "斎藤明美／著" || got.Publisher != "筑摩書房" || got.PublishedOn != "20170807" {
			t.Fatalf("got %+v", got)
		}
		if got.Price == nil || *got.Price != 950 {
			t.Fatalf("Price = %v, want 950", got.Price)
		}
		if got.Cover == nil || got.Cover.URL() != "https://cover.openbd.jp/9784480434623.jpg" || got.Cover.Source() != domainbook.CoverSourceOpenBD {
			t.Fatalf("Cover = %+v", got.Cover)
		}
	})

	t.Run("書影が空なら書影なしで返す", func(t *testing.T) {
		srv, _ := serve(t, http.StatusOK, withoutCover)
		got, err := openbd.NewCatalog(srv.URL, srv.Client()).Lookup(context.Background(), isbn(t, "9784873118703"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Cover != nil || got.Authors != "Kleppmann,Martin 斉藤,太郎,pub.2019 玉川,竜司,1964-" {
			t.Fatalf("got %+v", got)
		}
	})

	t.Run("該当なし（配列のnull）はErrNotFound", func(t *testing.T) {
		srv, _ := serve(t, http.StatusOK, `[null]`)
		if _, err := openbd.NewCatalog(srv.URL, srv.Client()).Lookup(context.Background(), isbn(t, "9784873118703")); !errors.Is(err, common.ErrNotFound) {
			t.Fatalf("err = %v, want ErrNotFound", err)
		}
	})

	failures := []struct {
		name   string
		status int
		body   string
	}{
		{name: "200以外はエラー", status: http.StatusServiceUnavailable, body: ""},
		{name: "JSONでなければエラー", status: http.StatusOK, body: "<html>"},
		{name: "書影URLが不正ならエラー", status: http.StatusOK, body: `[{"summary":{"title":"x","cover":"http://insecure.example/1.jpg"}}]`},
	}
	for _, tt := range failures {
		t.Run(tt.name, func(t *testing.T) {
			srv, _ := serve(t, tt.status, tt.body)
			_, err := openbd.NewCatalog(srv.URL, srv.Client()).Lookup(context.Background(), isbn(t, "9784873118703"))
			if err == nil || errors.Is(err, common.ErrNotFound) {
				t.Fatalf("err = %v, want a non-NotFound error", err)
			}
		})
	}
}
