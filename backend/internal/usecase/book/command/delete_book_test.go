package command_test

import (
	"context"
	"errors"
	"testing"

	"github.com/mrstsgk/book-management-system/backend/internal/domain/common"
	"github.com/mrstsgk/book-management-system/backend/internal/usecase/book/command"
)

func TestDeleteBookUsecase_Execute(t *testing.T) {
	t.Parallel()

	t.Run("指定したIDの本を削除する", func(t *testing.T) {
		t.Parallel()
		books := &fakeBooks{}
		if err := (&command.DeleteBookUsecaseImpl{Books: books}).Execute(context.Background(), 7); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if books.deletedID != 7 {
			t.Fatalf("deleted %d, want 7", books.deletedID)
		}
	})

	t.Run("存在しない本はNotFoundのまま返す", func(t *testing.T) {
		t.Parallel()
		err := (&command.DeleteBookUsecaseImpl{Books: &fakeBooks{deleteErr: common.ErrNotFound}}).Execute(context.Background(), 7)
		if !errors.Is(err, common.ErrNotFound) {
			t.Fatalf("err = %v, want ErrNotFound", err)
		}
	})
}
