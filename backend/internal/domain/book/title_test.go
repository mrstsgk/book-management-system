package book_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/mrstsgk/book-management-system/backend/internal/domain/book"
	"github.com/mrstsgk/book-management-system/backend/internal/domain/common"
)

func TestNewTitle(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		in      string
		wantErr bool
	}{
		{name: "通常のタイトルは有効", in: "人間失格"},
		{name: "255文字ちょうどは有効", in: strings.Repeat("あ", 255)},
		{name: "256文字はエラー", in: strings.Repeat("あ", 256), wantErr: true},
		{name: "空文字はエラー", in: "", wantErr: true},
		{name: "半角空白を含むとエラー", in: "人間 失格", wantErr: true},
		{name: "全角空白を含むとエラー", in: "人間　失格", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := book.NewTitle(tt.in)
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
