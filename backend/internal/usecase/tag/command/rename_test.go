package command_test

import (
	"context"
	"errors"
	"testing"

	"github.com/mrstsgk/book-management-system/backend/internal/domain/common"
	"github.com/mrstsgk/book-management-system/backend/internal/domain/tag"
	"github.com/mrstsgk/book-management-system/backend/internal/usecase/tag/command"
)

func TestRenameUsecase_Execute(t *testing.T) {
	t.Parallel()

	existing := func(t *testing.T) *tag.Tag {
		t.Helper()
		n, err := tag.NewName("旧名")
		if err != nil {
			t.Fatal(err)
		}
		tg := tag.New(n)
		tg.ID, tg.Version = 5, 2
		return tg
	}

	t.Run("名前を差し替えて返す", func(t *testing.T) {
		t.Parallel()
		tags := &fakeTags{findByID: existing(t)}
		uc := &command.RenameUsecaseImpl{Tags: tags}

		got, err := uc.Execute(context.Background(), command.RenameCommand{ID: 5, Name: "新名", Version: 2})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if tags.updated == nil || tags.updated.Name.String() != "新名" {
			t.Fatalf("updated %+v", tags.updated)
		}
		if got.Name != "新名" || got.Version != 3 {
			t.Fatalf("got %+v", got)
		}
	})

	t.Run("空文字はエラーで保存しない", func(t *testing.T) {
		t.Parallel()
		tags := &fakeTags{findByID: existing(t)}
		uc := &command.RenameUsecaseImpl{Tags: tags}

		if _, err := uc.Execute(context.Background(), command.RenameCommand{ID: 5, Name: " ", Version: 2}); !errors.Is(err, common.ErrInvalid) {
			t.Fatalf("err = %v, want ErrInvalid", err)
		}
		if tags.updated != nil {
			t.Fatal("Repository.Update must not be called")
		}
	})

	t.Run("存在しないタグはNotFound", func(t *testing.T) {
		t.Parallel()
		tags := &fakeTags{findByIDErr: common.ErrNotFound}
		uc := &command.RenameUsecaseImpl{Tags: tags}

		if _, err := uc.Execute(context.Background(), command.RenameCommand{ID: 5, Name: "新名", Version: 2}); !errors.Is(err, common.ErrNotFound) {
			t.Fatalf("err = %v, want ErrNotFound", err)
		}
	})

	t.Run("バージョン競合はConflict", func(t *testing.T) {
		t.Parallel()
		tags := &fakeTags{findByID: existing(t), updateErr: common.ErrConflict}
		uc := &command.RenameUsecaseImpl{Tags: tags}

		if _, err := uc.Execute(context.Background(), command.RenameCommand{ID: 5, Name: "新名", Version: 2}); !errors.Is(err, common.ErrConflict) {
			t.Fatalf("err = %v, want ErrConflict", err)
		}
	})
}
