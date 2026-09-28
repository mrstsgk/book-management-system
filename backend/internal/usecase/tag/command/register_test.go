package command_test

import (
	"context"
	"errors"
	"testing"

	"github.com/mrstsgk/book-management-system/backend/internal/domain/common"
	"github.com/mrstsgk/book-management-system/backend/internal/domain/tag"
	"github.com/mrstsgk/book-management-system/backend/internal/usecase/tag/command"
)

// fakeTags は tag.Repository の手書き Fake（docs/rules/testing.md）。
type fakeTags struct {
	created   *tag.Tag
	createErr error

	findByID    *tag.Tag
	findByIDErr error

	updated   *tag.Tag
	updateErr error

	deletedID tag.ID
	deleteErr error
}

func (f *fakeTags) FindByID(context.Context, tag.ID) (*tag.Tag, error) {
	return f.findByID, f.findByIDErr
}

func (f *fakeTags) Create(_ context.Context, t *tag.Tag) error {
	if f.createErr != nil {
		return f.createErr
	}
	t.ID, t.Version = 1, 1
	f.created = t
	return nil
}

func (f *fakeTags) Update(_ context.Context, t *tag.Tag) error {
	if f.updateErr != nil {
		return f.updateErr
	}
	t.Version++
	f.updated = t
	return nil
}

func (f *fakeTags) Delete(_ context.Context, id tag.ID) error {
	f.deletedID = id
	return f.deleteErr
}

func TestRegisterUsecase_Execute(t *testing.T) {
	t.Parallel()

	t.Run("タグ名を登録して返す", func(t *testing.T) {
		t.Parallel()
		tags := &fakeTags{}
		uc := &command.RegisterUsecaseImpl{Tags: tags}

		got, err := uc.Execute(context.Background(), command.RegisterCommand{Name: "データベース"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if tags.created == nil || tags.created.Name.String() != "データベース" {
			t.Fatalf("created %+v", tags.created)
		}
		if got.ID != 1 || got.Name != "データベース" || got.Version != 1 {
			t.Fatalf("got %+v", got)
		}
	})

	t.Run("空文字は保存を呼ばずにエラー", func(t *testing.T) {
		t.Parallel()
		tags := &fakeTags{}
		uc := &command.RegisterUsecaseImpl{Tags: tags}

		if _, err := uc.Execute(context.Background(), command.RegisterCommand{Name: " "}); !errors.Is(err, common.ErrInvalid) {
			t.Fatalf("err = %v, want ErrInvalid", err)
		}
		if tags.created != nil {
			t.Fatal("Repository.Create must not be called")
		}
	})

	t.Run("同名の登録済みはConflict", func(t *testing.T) {
		t.Parallel()
		tags := &fakeTags{createErr: common.ErrConflict}
		uc := &command.RegisterUsecaseImpl{Tags: tags}

		if _, err := uc.Execute(context.Background(), command.RegisterCommand{Name: "データベース"}); !errors.Is(err, common.ErrConflict) {
			t.Fatalf("err = %v, want ErrConflict", err)
		}
	})
}
