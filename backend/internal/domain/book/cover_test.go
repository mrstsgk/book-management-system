package book_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/mrstsgk/book-management-system/backend/internal/domain/book"
	"github.com/mrstsgk/book-management-system/backend/internal/domain/common"
)

func TestNewCover(t *testing.T) {
	t.Parallel()
	const prefix = "https://cover.openbd.jp/"
	tests := []struct {
		name    string
		url     string
		source  book.CoverSource
		wantErr bool
	}{
		{name: "openBDの書影は有効", url: "https://cover.openbd.jp/9784480434623.jpg", source: book.CoverSourceOpenBD},
		{name: "楽天の書影は商品ページと取得日時が要るのでエラー", url: "https://thumbnail.image.rakuten.co.jp/0_mall/book/cabinet/1.jpg", source: book.CoverSourceRakuten, wantErr: true},
		{name: "2048文字ちょうどは有効", url: prefix + strings.Repeat("a", 2048-len(prefix)), source: book.CoverSourceOpenBD},
		{name: "2049文字はエラー", url: prefix + strings.Repeat("a", 2049-len(prefix)), source: book.CoverSourceOpenBD, wantErr: true},
		{name: "httpはエラー", url: "http://cover.openbd.jp/1.jpg", source: book.CoverSourceOpenBD, wantErr: true},
		{name: "ホストが無いとエラー", url: "https:///1.jpg", source: book.CoverSourceOpenBD, wantErr: true},
		{name: "ポートだけでホスト名が無いとエラー", url: "https://:443/1.jpg", source: book.CoverSourceOpenBD, wantErr: true},
		{name: "ユーザー情報付きはエラー", url: "https://user@cover.openbd.jp/1.jpg", source: book.CoverSourceOpenBD, wantErr: true},
		{name: "未知の提供元はエラー", url: "https://cover.openbd.jp/1.jpg", source: book.CoverSource("amazon"), wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := book.NewCover(tt.url, tt.source)
			if tt.wantErr {
				if !errors.Is(err, common.ErrInvalid) {
					t.Fatalf("err = %v, want ErrInvalid", err)
				}
				return
			}
			if err != nil || got.URL() != tt.url || got.Source() != tt.source {
				t.Fatalf("got (%s, %s, %v), want (%s, %s, nil)", got.URL(), got.Source(), err, tt.url, tt.source)
			}
		})
	}
}

func TestNewRakutenCover(t *testing.T) {
	t.Parallel()
	const image = "https://thumbnail.image.rakuten.co.jp/0_mall/book/cabinet/1.jpg"
	const product = "https://books.rakuten.co.jp/rb/15949390/"
	const productPrefix = "https://books.rakuten.co.jp/"
	fetched := time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)
	tests := []struct {
		name      string
		image     string
		product   string
		fetchedAt time.Time
		wantErr   bool
	}{
		{name: "画像・商品ページ・取得日時が揃えば有効", image: image, product: product, fetchedAt: fetched},
		{name: "商品ページが2048文字ちょうどは有効", image: image, product: productPrefix + strings.Repeat("a", 2048-len(productPrefix)), fetchedAt: fetched},
		{name: "商品ページが2049文字はエラー", image: image, product: productPrefix + strings.Repeat("a", 2049-len(productPrefix)), fetchedAt: fetched, wantErr: true},
		{name: "商品ページが空はエラー", image: image, product: "", fetchedAt: fetched, wantErr: true},
		{name: "商品ページがhttpはエラー", image: image, product: "http://books.rakuten.co.jp/rb/1/", fetchedAt: fetched, wantErr: true},
		{name: "画像がhttpはエラー", image: "http://thumbnail.image.rakuten.co.jp/1.jpg", product: product, fetchedAt: fetched, wantErr: true},
		{name: "取得日時がゼロ値はエラー", image: image, product: product, fetchedAt: time.Time{}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := book.NewRakutenCover(tt.image, tt.product, tt.fetchedAt)
			if tt.wantErr {
				if !errors.Is(err, common.ErrInvalid) {
					t.Fatalf("err = %v, want ErrInvalid", err)
				}
				return
			}
			if err != nil || got.URL() != tt.image || got.Source() != book.CoverSourceRakuten ||
				got.ProductURL() != tt.product || !got.FetchedAt().Equal(tt.fetchedAt) {
				t.Fatalf("got (%+v, %v)", got, err)
			}
		})
	}
}

func TestCover_ExpiryAndRefresh(t *testing.T) {
	t.Parallel()
	fetched := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	rakuten, err := book.NewRakutenCover("https://thumbnail.image.rakuten.co.jp/1.jpg", "https://books.rakuten.co.jp/rb/1/", fetched)
	if err != nil {
		t.Fatal(err)
	}
	openbd, err := book.NewCover("https://cover.openbd.jp/1.jpg", book.CoverSourceOpenBD)
	if err != nil {
		t.Fatal(err)
	}
	day := 24 * time.Hour
	tests := []struct {
		name             string
		cover            book.Cover
		now              time.Time
		wantExpired      bool
		wantNeedsRefresh bool
	}{
		{name: "楽天で取得から83日の直前は取り直さない", cover: rakuten, now: fetched.Add(83*day - time.Nanosecond)},
		{name: "楽天で取得から83日ちょうどは取り直す", cover: rakuten, now: fetched.Add(83 * day), wantNeedsRefresh: true},
		{name: "楽天で取得から90日の直前はまだ期限内", cover: rakuten, now: fetched.Add(90*day - time.Nanosecond), wantNeedsRefresh: true},
		{name: "楽天で取得から90日ちょうどは期限切れ", cover: rakuten, now: fetched.Add(90 * day), wantExpired: true, wantNeedsRefresh: true},
		{name: "openBDは10年たっても期限切れにも取り直しにもならない", cover: openbd, now: fetched.Add(3650 * day)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.cover.IsExpired(tt.now); got != tt.wantExpired {
				t.Fatalf("IsExpired = %v, want %v", got, tt.wantExpired)
			}
			if got := tt.cover.NeedsRefresh(tt.now); got != tt.wantNeedsRefresh {
				t.Fatalf("NeedsRefresh = %v, want %v", got, tt.wantNeedsRefresh)
			}
		})
	}
}
