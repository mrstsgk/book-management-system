package command_test

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/mrstsgk/book-management-system/backend/internal/domain/author"
	"github.com/mrstsgk/book-management-system/backend/internal/domain/book"
	"github.com/mrstsgk/book-management-system/backend/internal/domain/common"
	"github.com/mrstsgk/book-management-system/backend/internal/usecase/book/command"
)

func TestCreateUsecase_Execute(t *testing.T) {
	t.Parallel()

	t.Run("保存した書籍の詳細を返す", func(t *testing.T) {
		t.Parallel()
		books := &fakeBooks{}
		want := &book.BookDetail{ID: 1, Title: "人間失格"}
		details := &fakeDetails{detail: want}
		uc := &command.CreateUsecaseImpl{Books: books, Authors: authorsExisting(1, 2), Details: details}

		got, err := uc.Execute(context.Background(), command.CreateCommand{Title: "人間失格", Price: 1500, AuthorIDs: []int64{1, 2}, Status: 2})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if books.created == nil {
			t.Fatal("Repository.Create was not called")
		}
		c := books.created
		if c.Title.String() != "人間失格" || c.Price.Int64() != 1500 || c.Status != book.Published ||
			!reflect.DeepEqual(c.AuthorIDs, []author.ID{1, 2}) {
			t.Fatalf("created %+v", c)
		}
		if details.gotID != 1 || got != want {
			t.Fatalf("detail looked up for id=%d, got %+v; want the created id and its detail", details.gotID, got)
		}
	})

	t.Run("Repositoryのエラーはそのまま返り詳細は取得しない", func(t *testing.T) {
		t.Parallel()
		wantErr := errors.New("db down")
		details := &fakeDetails{}
		uc := &command.CreateUsecaseImpl{Books: &fakeBooks{createErr: wantErr}, Authors: authorsExisting(1), Details: details}

		_, err := uc.Execute(context.Background(), command.CreateCommand{Title: "人間失格", Price: 1500, AuthorIDs: []int64{1}, Status: 1})
		if !errors.Is(err, wantErr) {
			t.Fatalf("err = %v, want %v", err, wantErr)
		}
		if details.gotID != 0 {
			t.Fatal("the detail must not be queried when saving fails")
		}
	})
}

func TestCreateUsecase_DuplicateISBNConflictPropagates(t *testing.T) {
	t.Parallel()
	details := &fakeDetails{}
	uc := &command.CreateUsecaseImpl{
		Books: &fakeBooks{createErr: common.ErrConflict}, Authors: authorsExisting(1), Details: details,
		Catalog: &fakeCatalog{entry: &book.CatalogEntry{}},
	}
	isbn := "9784873118703"

	_, err := uc.Execute(context.Background(), command.CreateCommand{Title: "人間失格", Price: 1500, AuthorIDs: []int64{1}, Status: 1, ISBN: &isbn})
	if !errors.Is(err, common.ErrConflict) {
		t.Fatalf("err = %v, want ErrConflict", err)
	}
	if details.gotID != 0 {
		t.Fatal("the detail must not be queried when saving fails")
	}
}
