package author_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/mrstsgk/book-management-system/backend/internal/domain/author"
	"github.com/mrstsgk/book-management-system/backend/internal/domain/common"
)

func TestNewName(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		in      string
		want    string
		wantErr bool
	}{
		{name: "通常の著者名は有効", in: "太宰治", want: "太宰治"},
		{name: "途中の半角空白は残す", in: "Martin Kleppmann", want: "Martin Kleppmann"},
		{name: "途中の全角空白は残す", in: "太宰　治", want: "太宰　治"},
		{name: "前後の半角・全角空白は取り除く", in: " 　太宰治　 ", want: "太宰治"},
		{name: "100文字ちょうどは有効", in: strings.Repeat("あ", 100), want: strings.Repeat("あ", 100)},
		{name: "前後の空白を除いて100文字なら有効", in: " " + strings.Repeat("あ", 100) + " ", want: strings.Repeat("あ", 100)},
		{name: "101文字はエラー", in: strings.Repeat("あ", 101), wantErr: true},
		{name: "空文字はエラー", in: "", wantErr: true},
		{name: "空白だけはエラー", in: " 　 ", wantErr: true},
		{name: "途中のタブはエラー", in: "太宰\t治", wantErr: true},
		{name: "途中の改行はエラー", in: "太宰\n治", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := author.NewName(tt.in)
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
				t.Fatalf("String() = %q, want %q", got.String(), tt.want)
			}
		})
	}
}
