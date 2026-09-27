package book_test

import (
	"errors"
	"testing"

	"github.com/mrstsgk/book-management-system/backend/internal/domain/book"
	"github.com/mrstsgk/book-management-system/backend/internal/domain/common"
)

func TestNewPrice(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		in      int64
		wantErr bool
	}{
		{name: "0円は有効", in: 0},
		{name: "上限ちょうどは有効", in: 99_999_999},
		{name: "上限+1はエラー", in: 100_000_000, wantErr: true},
		{name: "負の値はエラー", in: -1, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := book.NewPrice(tt.in)
			if tt.wantErr {
				if !errors.Is(err, common.ErrInvalid) {
					t.Fatalf("err = %v, want ErrInvalid", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Int64() != tt.in {
				t.Fatalf("Int64() = %d, want %d", got.Int64(), tt.in)
			}
		})
	}
}
