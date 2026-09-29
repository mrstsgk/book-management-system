package book_test

import (
	"errors"
	"strings"
	"testing"

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
		{name: "撤去した楽天は未知の提供元としてエラー", url: "https://thumbnail.image.rakuten.co.jp/0_mall/book/cabinet/1.jpg", source: book.CoverSource("rakuten"), wantErr: true},
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
			if got.PageURL() != "" {
				t.Fatalf("PageURL() = %q, want empty (openBD の書影は本のページを持たない)", got.PageURL())
			}
		})
	}
}

func TestNewCover_GoogleBooksの書影はページが要るので作れない(t *testing.T) {
	t.Parallel()
	// ページ無しで作れてしまうと、規約上必要な Google Books へのリンクを出せない書影が保存されるため
	_, err := book.NewCover("https://books.google.com/books/content?id=abc", book.CoverSourceGoogleBooks)
	if !errors.Is(err, common.ErrInvalid) {
		t.Fatalf("err = %v, want ErrInvalid", err)
	}
}

func TestNewGoogleBooksCover(t *testing.T) {
	t.Parallel()
	const (
		image = "https://books.google.com/books/content?id=abc&printsec=frontcover&img=1&zoom=1"
		page  = "https://books.google.co.jp/books?id=abc"
	)
	tests := []struct {
		name    string
		image   string
		page    string
		wantErr bool
	}{
		{name: "画像とGoogle Booksのページが揃えば有効", image: image, page: page},
		{name: "ページが空ならエラー", image: image, page: "", wantErr: true},
		{name: "ページがhttpならエラー", image: image, page: "http://books.google.co.jp/books?id=abc", wantErr: true},
		{name: "画像が空ならエラー", image: "", page: page, wantErr: true},
		{name: "画像がhttpならエラー", image: "http://books.google.com/books/content?id=abc", page: page, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := book.NewGoogleBooksCover(tt.image, tt.page)
			if tt.wantErr {
				if !errors.Is(err, common.ErrInvalid) {
					t.Fatalf("err = %v, want ErrInvalid", err)
				}
				return
			}
			if err != nil || got.URL() != tt.image || got.Source() != book.CoverSourceGoogleBooks || got.PageURL() != tt.page {
				t.Fatalf("got (%s, %s, %s, %v), want (%s, googlebooks, %s, nil)", got.URL(), got.Source(), got.PageURL(), err, tt.image, tt.page)
			}
		})
	}
}
