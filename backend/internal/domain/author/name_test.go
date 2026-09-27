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
		wantErr bool
	}{
		{name: "通常の著者名は有効", in: "太宰治"},
		{name: "100文字ちょうどは有効", in: strings.Repeat("あ", 100)},
		{name: "101文字はエラー", in: strings.Repeat("あ", 101), wantErr: true},
		{name: "空文字はエラー", in: "", wantErr: true},
		{name: "半角空白を含むとエラー", in: "太宰 治", wantErr: true},
		{name: "全角空白を含むとエラー", in: "太宰　治", wantErr: true},
		{name: "タブを含むとエラー", in: "太宰\t治", wantErr: true},
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
			if got.String() != tt.in {
				t.Fatalf("String() = %q, want %q", got.String(), tt.in)
			}
		})
	}
}
