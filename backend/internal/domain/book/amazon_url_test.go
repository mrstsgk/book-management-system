package book_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/mrstsgk/book-management-system/backend/internal/domain/book"
	"github.com/mrstsgk/book-management-system/backend/internal/domain/common"
)

func TestNewAmazonURL(t *testing.T) {
	t.Parallel()
	const prefix = "https://www.amazon.co.jp/dp/"
	tests := []struct {
		name    string
		in      string
		wantErr bool
	}{
		{name: "amazon.co.jpの商品URLは有効", in: "https://www.amazon.co.jp/dp/4101006059"},
		{name: "サブドメインなしのamazon.co.jpは有効", in: "https://amazon.co.jp/dp/4101006059"},
		{name: "amazon.comは有効", in: "https://www.amazon.com/dp/4101006059"},
		{name: "短縮URLamzn.asiaは有効", in: "https://amzn.asia/d/abc123"},
		{name: "短縮URLamzn.toは有効", in: "https://amzn.to/3abcDEF"},
		{name: "ホストの大文字は区別しない", in: "https://WWW.AMAZON.CO.JP/dp/4101006059"},
		{name: "2048文字ちょうどは有効", in: prefix + strings.Repeat("a", 2048-len(prefix))},
		{name: "2049文字はエラー", in: prefix + strings.Repeat("a", 2049-len(prefix)), wantErr: true},
		{name: "httpはエラー", in: "http://www.amazon.co.jp/dp/4101006059", wantErr: true},
		{name: "Amazon以外のホストはエラー", in: "https://example.com/dp/4101006059", wantErr: true},
		{name: "Amazonを含むだけの別ドメインはエラー", in: "https://amazon.co.jp.evil.example/dp/1", wantErr: true},
		{name: "末尾一致だけの別ドメインはエラー", in: "https://notamazon.com/dp/1", wantErr: true},
		{name: "ユーザー情報付きはエラー", in: "https://amazon.co.jp@evil.example/dp/1", wantErr: true},
		{name: "ポート指定はエラー", in: "https://www.amazon.co.jp:8443/dp/1", wantErr: true},
		{name: "空文字はエラー", in: "", wantErr: true},
		{name: "URLでない文字列はエラー", in: "amazon.co.jp/dp/1", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := book.NewAmazonURL(tt.in)
			if tt.wantErr {
				if !errors.Is(err, common.ErrInvalid) {
					t.Fatalf("err = %v, want ErrInvalid", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.String() != tt.in {
				t.Fatalf("String() = %q, want %q", got.String(), tt.in)
			}
		})
	}
}
