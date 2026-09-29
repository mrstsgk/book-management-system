package googlebooks_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/mrstsgk/book-management-system/backend/internal/infrastructure/gateway/googlebooks"
)

// Books API v1 の volumes.list の応答（https://developers.google.com/books/docs/v1/reference/volumes）から
// 必要な部分だけを残したもの。thumbnail と infoLink は実際の応答と同じく http で返る。
const withCover = `{"kind":"books#volumes","totalItems":1,"items":[{"id":"abc123","volumeInfo":{"title":"良いコード／悪いコードで学ぶ設計入門",
"imageLinks":{"smallThumbnail":"http://books.google.com/books/content?id=abc123&printsec=frontcover&img=1&zoom=5&source=gbs_api",
"thumbnail":"http://books.google.com/books/content?id=abc123&printsec=frontcover&img=1&zoom=1&source=gbs_api"},
"infoLink":"http://books.google.co.jp/books?id=abc123&dq=isbn:9784297146221&source=gbs_api"}}]}`

func serve(t *testing.T, status int, body string) (*httptest.Server, *url.Values) {
	t.Helper()
	var got url.Values
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/books/v1/volumes" {
			http.NotFound(w, r)
			return
		}
		got = r.URL.Query()
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	return srv, &got
}

func TestClient_FindCover(t *testing.T) {
	t.Run("書影とGoogle Booksのページをhttpsにして返し、ISBNとキーを付けて問い合わせる", func(t *testing.T) {
		srv, query := serve(t, http.StatusOK, withCover)
		got, err := googlebooks.NewClient(srv.URL, "test-key", srv.Client()).FindCover(context.Background(), "9784297146221")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if q := query.Get("q"); q != "isbn:9784297146221" {
			t.Fatalf("q = %q", q)
		}
		if k := query.Get("key"); k != "test-key" {
			t.Fatalf("key = %q", k)
		}
		want := googlebooks.Cover{
			ImageURL: "https://books.google.com/books/content?id=abc123&printsec=frontcover&img=1&zoom=1&source=gbs_api",
			PageURL:  "https://books.google.co.jp/books?id=abc123&dq=isbn:9784297146221&source=gbs_api",
		}
		if got == nil || *got != want {
			t.Fatalf("got %+v, want %+v", got, want)
		}
	})

	for _, tt := range []struct {
		name string
		body string
	}{
		{name: "該当なし（totalItems 0）", body: `{"kind":"books#volumes","totalItems":0}`},
		{name: "imageLinks が無い", body: `{"totalItems":1,"items":[{"volumeInfo":{"title":"x","infoLink":"http://books.google.com/books?id=x"}}]}`},
		{name: "thumbnail が空", body: `{"totalItems":1,"items":[{"volumeInfo":{"imageLinks":{"thumbnail":""},"infoLink":"http://books.google.com/books?id=x"}}]}`},
		// Google の規約では書影を出すとき本ごとの Google Books へのリンクが必須なので、リンクが無い書影は使えない
		{name: "infoLink が無い", body: `{"totalItems":1,"items":[{"volumeInfo":{"imageLinks":{"thumbnail":"http://books.google.com/books/content?id=x"}}}]}`},
	} {
		t.Run(tt.name+"なら書影なし（nil, nil）", func(t *testing.T) {
			srv, _ := serve(t, http.StatusOK, tt.body)
			got, err := googlebooks.NewClient(srv.URL, "k", srv.Client()).FindCover(context.Background(), "9784297146221")
			if err != nil || got != nil {
				t.Fatalf("got (%+v, %v), want (nil, nil)", got, err)
			}
		})
	}

	for _, tt := range []struct {
		name   string
		status int
		body   string
	}{
		{name: "日次の上限（429）", status: http.StatusTooManyRequests, body: `{"error":{"code":429}}`},
		{name: "サーバーエラー（500）", status: http.StatusInternalServerError, body: `{}`},
		{name: "壊れたJSON", status: http.StatusOK, body: `{"totalItems":`},
	} {
		t.Run(tt.name+"はエラー", func(t *testing.T) {
			srv, _ := serve(t, tt.status, tt.body)
			got, err := googlebooks.NewClient(srv.URL, "k", srv.Client()).FindCover(context.Background(), "9784297146221")
			if err == nil || got != nil {
				t.Fatalf("got (%+v, %v), want an error", got, err)
			}
		})
	}

	t.Run("通信に失敗してもエラーにAPIキーを含めない", func(t *testing.T) {
		srv, _ := serve(t, http.StatusOK, withCover)
		srv.Close()
		_, err := googlebooks.NewClient(srv.URL, "secret-key", srv.Client()).FindCover(context.Background(), "9784297146221")
		if err == nil || strings.Contains(err.Error(), "secret-key") {
			t.Fatalf("err = %v, want an error without the key", err)
		}
	})

	t.Run("キャンセルされたctxならエラー", func(t *testing.T) {
		srv, _ := serve(t, http.StatusOK, withCover)
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		_, err := googlebooks.NewClient(srv.URL, "k", srv.Client()).FindCover(ctx, "9784297146221")
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("err = %v, want context.Canceled", err)
		}
	})
}
