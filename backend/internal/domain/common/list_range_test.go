package common_test

import (
	"errors"
	"testing"

	"github.com/mrstsgk/book-management-system/backend/internal/domain/common"
)

func TestNewListRange(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name          string
		limit, offset int
		wantErr       bool
	}{
		{name: "limit 1 は有効", limit: 1, offset: 0},
		{name: "limit 上限ちょうどは有効", limit: common.MaxListLimit, offset: 0},
		{name: "limit 上限+1はエラー", limit: common.MaxListLimit + 1, offset: 0, wantErr: true},
		{name: "limit 0 はエラー", limit: 0, offset: 0, wantErr: true},
		{name: "offset 0 は有効", limit: 20, offset: 0},
		{name: "offset 正の値は有効", limit: 20, offset: 40},
		{name: "offset 負の値はエラー", limit: 20, offset: -1, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := common.NewListRange(tt.limit, tt.offset)
			if tt.wantErr {
				if !errors.Is(err, common.ErrInvalid) {
					t.Fatalf("err = %v, want ErrInvalid", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Limit() != tt.limit || got.Offset() != tt.offset {
				t.Fatalf("got limit=%d offset=%d, want %d/%d", got.Limit(), got.Offset(), tt.limit, tt.offset)
			}
		})
	}
}

func TestDefaultListLimit_IsAValidLimit(t *testing.T) {
	t.Parallel()
	if _, err := common.NewListRange(common.DefaultListLimit, 0); err != nil {
		t.Fatalf("DefaultListLimit must be accepted by NewListRange: %v", err)
	}
}
