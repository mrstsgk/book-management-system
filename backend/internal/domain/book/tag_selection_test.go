package book_test

import (
	"errors"
	"testing"

	"github.com/mrstsgk/book-management-system/backend/internal/domain/book"
	"github.com/mrstsgk/book-management-system/backend/internal/domain/common"
	"github.com/mrstsgk/book-management-system/backend/internal/domain/tag"
)

func TestNewTagSelection(t *testing.T) {
	t.Parallel()
	ids := func(n int) []tag.ID {
		s := make([]tag.ID, n)
		for i := range s {
			s[i] = tag.ID(i + 1)
		}
		return s
	}

	tests := []struct {
		name    string
		ids     []tag.ID
		wantLen int
		wantErr bool
	}{
		{name: "0個は有効（タグ無し）", ids: nil, wantLen: 0},
		{name: "10個ちょうどは有効", ids: ids(10), wantLen: 10},
		{name: "11個はエラー", ids: ids(11), wantErr: true},
		{name: "同じIDの重複はエラー", ids: []tag.ID{1, 1}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := book.NewTagSelection(tt.ids)
			if tt.wantErr {
				if !errors.Is(err, common.ErrInvalid) {
					t.Fatalf("err = %v, want ErrInvalid", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(got.IDs()) != tt.wantLen {
				t.Fatalf("len(IDs()) = %d, want %d", len(got.IDs()), tt.wantLen)
			}
		})
	}
}
