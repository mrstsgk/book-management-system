package command_test

import (
	"context"
	"errors"
	"testing"

	bookcmd "github.com/mrstsgk/book-management-system/backend/internal/book/usecase/command"
	"github.com/mrstsgk/book-management-system/backend/internal/domainerr"
)

func TestDeleteBookUsecase_Execute(t *testing.T) {
	t.Parallel()

	t.Run("指定したIDの本を削除する", func(t *testing.T) {
		t.Parallel()
		books := &fakeBooks{}
		if err := (&bookcmd.DeleteBookUsecaseImpl{Books: books}).Execute(context.Background(), 7); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if books.deletedID != 7 {
			t.Fatalf("deleted %d, want 7", books.deletedID)
		}
	})

	t.Run("存在しない本はNotFoundのまま返す", func(t *testing.T) {
		t.Parallel()
		err := (&bookcmd.DeleteBookUsecaseImpl{Books: &fakeBooks{deleteErr: domainerr.ErrNotFound}}).Execute(context.Background(), 7)
		if !errors.Is(err, domainerr.ErrNotFound) {
			t.Fatalf("err = %v, want ErrNotFound", err)
		}
	})
}
