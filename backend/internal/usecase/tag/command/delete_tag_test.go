package command_test

import (
	"context"
	"errors"
	"testing"

	"github.com/mrstsgk/book-management-system/backend/internal/domain/common"
	"github.com/mrstsgk/book-management-system/backend/internal/usecase/tag/command"
)

func TestDeleteTagUsecase_Execute(t *testing.T) {
	t.Parallel()

	t.Run("指定したIDのタグを削除する", func(t *testing.T) {
		t.Parallel()
		tags := &fakeTags{}
		uc := &command.DeleteTagUsecaseImpl{Tags: tags}

		if err := uc.Execute(context.Background(), 7); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if tags.deletedID != 7 {
			t.Fatalf("deletedID = %d, want 7", tags.deletedID)
		}
	})

	t.Run("存在しないタグはNotFoundのまま返す", func(t *testing.T) {
		t.Parallel()
		tags := &fakeTags{deleteErr: common.ErrNotFound}
		uc := &command.DeleteTagUsecaseImpl{Tags: tags}

		if err := uc.Execute(context.Background(), 7); !errors.Is(err, common.ErrNotFound) {
			t.Fatalf("err = %v, want ErrNotFound", err)
		}
	})
}
