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
		{name: "楽天の書影は有効", url: "https://thumbnail.image.rakuten.co.jp/0_mall/book/cabinet/1.jpg", source: book.CoverSourceRakuten},
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
