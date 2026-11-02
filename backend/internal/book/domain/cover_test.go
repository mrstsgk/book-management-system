package domain_test

import (
	"errors"
	"strings"
	"testing"

	bookdomain "github.com/mrstsgk/book-management-system/backend/internal/book/domain"
	"github.com/mrstsgk/book-management-system/backend/internal/domainerr"
)

func TestNewCover(t *testing.T) {
	t.Parallel()
	const prefix = "https://cover.openbd.jp/"
	tests := []struct {
		name    string
		url     string
		source  bookdomain.CoverSource
		wantErr bool
	}{
		{name: "openBDの書影は有効", url: "https://cover.openbd.jp/9784480434623.jpg", source: bookdomain.CoverSourceOpenBD},
		{name: "撤去した楽天は未知の提供元としてエラー", url: "https://thumbnail.image.rakuten.co.jp/0_mall/book/cabinet/1.jpg", source: bookdomain.CoverSource("rakuten"), wantErr: true},
		{name: "2048文字ちょうどは有効", url: prefix + strings.Repeat("a", 2048-len(prefix)), source: bookdomain.CoverSourceOpenBD},
		{name: "2049文字はエラー", url: prefix + strings.Repeat("a", 2049-len(prefix)), source: bookdomain.CoverSourceOpenBD, wantErr: true},
		{name: "httpはエラー", url: "http://cover.openbd.jp/1.jpg", source: bookdomain.CoverSourceOpenBD, wantErr: true},
		{name: "ホストが無いとエラー", url: "https:///1.jpg", source: bookdomain.CoverSourceOpenBD, wantErr: true},
		{name: "ポートだけでホスト名が無いとエラー", url: "https://:443/1.jpg", source: bookdomain.CoverSourceOpenBD, wantErr: true},
		{name: "ユーザー情報付きはエラー", url: "https://user@cover.openbd.jp/1.jpg", source: bookdomain.CoverSourceOpenBD, wantErr: true},
		{name: "未知の提供元はエラー", url: "https://cover.openbd.jp/1.jpg", source: bookdomain.CoverSource("amazon"), wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := bookdomain.NewCover(tt.url, tt.source)
			if tt.wantErr {
				if !errors.Is(err, domainerr.ErrInvalid) {
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
