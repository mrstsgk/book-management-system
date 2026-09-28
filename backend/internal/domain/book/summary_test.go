package book_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/mrstsgk/book-management-system/backend/internal/domain/book"
	"github.com/mrstsgk/book-management-system/backend/internal/domain/common"
)

func TestNewSummary(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		in      string
		want    string
		wantErr bool
	}{
		{name: "前後の空白を除いて保持する", in: "  分散データの設計を体系的に学べる  ", want: "分散データの設計を体系的に学べる"},
		{name: "1文字は有効", in: "良", want: "良"},
		{name: "100文字ちょうどは有効", in: strings.Repeat("あ", 100), want: strings.Repeat("あ", 100)},
		{name: "101文字はエラー", in: strings.Repeat("あ", 101), wantErr: true},
		{name: "空文字はエラー", in: "", wantErr: true},
		{name: "空白だけはエラー", in: "  ", wantErr: true},
		{name: "改行はエラー（感想と違い1行の文）", in: "設計\n入門", wantErr: true},
		{name: "前後の改行もエラー", in: "設計入門\n", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := book.NewSummary(tt.in)
			if tt.wantErr {
				if !errors.Is(err, common.ErrInvalid) {
					t.Fatalf("err = %v, want ErrInvalid", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.String() != tt.want {
				t.Fatalf("got %q, want %q", got.String(), tt.want)
			}
		})
	}
}
