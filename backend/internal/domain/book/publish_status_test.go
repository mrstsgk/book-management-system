package book_test

import (
	"errors"
	"testing"

	"github.com/mrstsgk/book-management-system/backend/internal/domain/book"
	"github.com/mrstsgk/book-management-system/backend/internal/domain/common"
)

func TestNewPublishStatus(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		in      int
		want    book.PublishStatus
		wantErr bool
	}{
		{name: "1は未出版", in: 1, want: book.Unpublished},
		{name: "2は出版済み", in: 2, want: book.Published},
		{name: "0はエラー", in: 0, wantErr: true},
		{name: "3はエラー", in: 3, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := book.NewPublishStatus(tt.in)
			if tt.wantErr {
				if !errors.Is(err, common.ErrInvalid) {
					t.Fatalf("err = %v, want ErrInvalid", err)
				}
				return
			}
			if err != nil || got != tt.want {
				t.Fatalf("got (%v, %v), want (%v, nil)", got, err, tt.want)
			}
		})
	}
}

func TestPublishStatus_CanChangeTo(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		from, to book.PublishStatus
		want     bool
	}{
		{name: "未出版から未出版は可", from: book.Unpublished, to: book.Unpublished, want: true},
		{name: "未出版から出版済みは可", from: book.Unpublished, to: book.Published, want: true},
		{name: "出版済みから出版済みは可", from: book.Published, to: book.Published, want: true},
		{name: "出版済みから未出版は不可", from: book.Published, to: book.Unpublished, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.from.CanChangeTo(tt.to); got != tt.want {
				t.Fatalf("CanChangeTo = %v, want %v", got, tt.want)
			}
		})
	}
}
