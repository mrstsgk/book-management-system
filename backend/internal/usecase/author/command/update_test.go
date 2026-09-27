package command_test

import (
	"context"
	"errors"
	"testing"

	"github.com/mrstsgk/book-management-system/backend/internal/domain/author"
	"github.com/mrstsgk/book-management-system/backend/internal/domain/common"
	"github.com/mrstsgk/book-management-system/backend/internal/usecase/author/command"
)

func existingAuthor(t *testing.T) *author.Author {
	t.Helper()
	name, err := author.NewName("太宰治")
	if err != nil {
		t.Fatal(err)
	}
	return &author.Author{ID: 1, Name: name, Version: 3}
}

func TestUpdateUsecase_Execute(t *testing.T) {
	t.Parallel()

	t.Run("リクエストのバージョンで更新しバージョンが進む", func(t *testing.T) {
		t.Parallel()
		repo := &fakeRepository{findByID: existingAuthor(t)}
		uc := &command.UpdateUsecaseImpl{Authors: repo, Now: fixedNow}

		got, err := uc.Execute(context.Background(), command.UpdateCommand{ID: 1, Name: "津島修治", BirthDate: date(1909, 6, 19), Version: 2})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if repo.updated == nil {
			t.Fatal("Repository.Update was not called")
		}
		if got.Name.String() != "津島修治" || got.BirthDate == nil || got.Version != 3 {
			t.Fatalf("got %+v, want name 津島修治 / birth date set / version 3 (2 + 1)", got)
		}
	})

	t.Run("存在しない著者はNotFoundを返し更新しない", func(t *testing.T) {
		t.Parallel()
		repo := &fakeRepository{findByIDErr: common.ErrNotFound}
		uc := &command.UpdateUsecaseImpl{Authors: repo, Now: fixedNow}

		if _, err := uc.Execute(context.Background(), command.UpdateCommand{ID: 99, Name: "太宰治", Version: 1}); !errors.Is(err, common.ErrNotFound) {
			t.Fatalf("err = %v, want ErrNotFound", err)
		}
		if repo.updated != nil {
			t.Fatal("Repository.Update must not be called when the author is missing")
		}
	})

	t.Run("不正な名前は更新せずエラーで既存の著者も変わらない", func(t *testing.T) {
		t.Parallel()
		existing := existingAuthor(t)
		repo := &fakeRepository{findByID: existing}
		uc := &command.UpdateUsecaseImpl{Authors: repo, Now: fixedNow}

		if _, err := uc.Execute(context.Background(), command.UpdateCommand{ID: 1, Name: "", Version: 3}); !errors.Is(err, common.ErrInvalid) {
			t.Fatalf("err = %v, want ErrInvalid", err)
		}
		if repo.updated != nil || existing.Name.String() != "太宰治" || existing.Version != 3 {
			t.Fatalf("author changed on failure: %+v", existing)
		}
	})

	t.Run("楽観的ロックの競合はConflictを返す", func(t *testing.T) {
		t.Parallel()
		repo := &fakeRepository{findByID: existingAuthor(t), updateErr: common.ErrConflict}
		uc := &command.UpdateUsecaseImpl{Authors: repo, Now: fixedNow}

		if _, err := uc.Execute(context.Background(), command.UpdateCommand{ID: 1, Name: "太宰治", Version: 1}); !errors.Is(err, common.ErrConflict) {
			t.Fatalf("err = %v, want ErrConflict", err)
		}
	})
}
