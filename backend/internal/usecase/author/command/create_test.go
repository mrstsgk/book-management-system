package command_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/mrstsgk/book-management-system/backend/internal/domain/author"
	"github.com/mrstsgk/book-management-system/backend/internal/domain/common"
	"github.com/mrstsgk/book-management-system/backend/internal/usecase/author/command"
)

// fakeRepository is a hand-written Fake for author.Repository (docs/rules/testing.md).
type fakeRepository struct {
	created   *author.Author
	createErr error

	findByID    *author.Author
	findByIDErr error

	updated   *author.Author
	updateErr error
}

func (f *fakeRepository) FindByID(context.Context, author.ID) (*author.Author, error) {
	return f.findByID, f.findByIDErr
}

func (f *fakeRepository) CountByIDs(context.Context, []author.ID) (int, error) {
	return 0, nil
}

func (f *fakeRepository) Create(_ context.Context, a *author.Author) error {
	if f.createErr != nil {
		return f.createErr
	}
	a.ID, a.Version = 1, 1
	f.created = a
	return nil
}

func (f *fakeRepository) Update(_ context.Context, a *author.Author) error {
	if f.updateErr != nil {
		return f.updateErr
	}
	a.Version++
	f.updated = a
	return nil
}

var fixedNow = func() time.Time { return time.Date(2026, 9, 28, 12, 0, 0, 0, time.UTC) }

func date(y int, m time.Month, d int) *time.Time {
	t := time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
	return &t
}

func TestCreateUsecase_Execute(t *testing.T) {
	t.Parallel()

	t.Run("名前と生年月日が保存され採番結果が返る", func(t *testing.T) {
		t.Parallel()
		repo := &fakeRepository{}
		uc := &command.CreateUsecaseImpl{Authors: repo, Now: fixedNow}

		got, err := uc.Execute(context.Background(), command.CreateCommand{Name: "太宰治", BirthDate: date(1909, 6, 19)})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if repo.created == nil {
			t.Fatal("Repository.Create was not called")
		}
		if got.ID != 1 || got.Version != 1 || got.Name.String() != "太宰治" || !got.BirthDate.Time().Equal(*date(1909, 6, 19)) {
			t.Fatalf("got %+v", got)
		}
	})

	t.Run("生年月日は省略できる", func(t *testing.T) {
		t.Parallel()
		repo := &fakeRepository{}
		uc := &command.CreateUsecaseImpl{Authors: repo, Now: fixedNow}

		got, err := uc.Execute(context.Background(), command.CreateCommand{Name: "太宰治"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.BirthDate != nil {
			t.Fatalf("BirthDate = %v, want nil", got.BirthDate)
		}
	})

	invalid := []struct {
		name string
		cmd  command.CreateCommand
	}{
		{name: "空白を含む名前は保存せずエラー", cmd: command.CreateCommand{Name: "太宰 治"}},
		{name: "今日の生年月日は保存せずエラー", cmd: command.CreateCommand{Name: "太宰治", BirthDate: date(2026, 9, 28)}},
	}
	for _, tt := range invalid {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			repo := &fakeRepository{}
			uc := &command.CreateUsecaseImpl{Authors: repo, Now: fixedNow}

			if _, err := uc.Execute(context.Background(), tt.cmd); !errors.Is(err, common.ErrInvalid) {
				t.Fatalf("err = %v, want ErrInvalid", err)
			}
			if repo.created != nil {
				t.Fatal("Repository.Create must not be called when validation fails")
			}
		})
	}

	t.Run("Repositoryのエラーはそのまま返る", func(t *testing.T) {
		t.Parallel()
		wantErr := errors.New("db down")
		uc := &command.CreateUsecaseImpl{Authors: &fakeRepository{createErr: wantErr}, Now: fixedNow}

		if _, err := uc.Execute(context.Background(), command.CreateCommand{Name: "太宰治"}); !errors.Is(err, wantErr) {
			t.Fatalf("err = %v, want %v", err, wantErr)
		}
	})
}
