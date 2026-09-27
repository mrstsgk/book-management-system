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

func TestCreateUsecase_AmazonURL(t *testing.T) {
	t.Parallel()
	base := command.CreateCommand{Title: "人間失格", Price: 1500, AuthorIDs: []int64{1}, Status: 1}

	t.Run("AmazonのURLを保存する", func(t *testing.T) {
		t.Parallel()
		books := &fakeBooks{}
		uc := &command.CreateUsecaseImpl{Books: books, Authors: authorsExisting(1), Details: &fakeDetails{detail: &book.BookDetail{}}}
		cmd := base
		u := "https://www.amazon.co.jp/dp/4101006059"
		cmd.AmazonURL = &u

		if _, err := uc.Execute(context.Background(), cmd); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if books.created.AmazonURL == nil || books.created.AmazonURL.String() != u {
			t.Fatalf("AmazonURL = %v, want %s", books.created.AmazonURL, u)
		}
	})

	t.Run("Amazon以外のURLはエラーで保存しない", func(t *testing.T) {
		t.Parallel()
		books := &fakeBooks{}
		uc := &command.CreateUsecaseImpl{Books: books, Authors: authorsExisting(1), Details: &fakeDetails{}}
		cmd := base
		u := "https://example.com/dp/1"
		cmd.AmazonURL = &u

		if _, err := uc.Execute(context.Background(), cmd); !errors.Is(err, common.ErrInvalid) {
			t.Fatalf("err = %v, want ErrInvalid", err)
		}
		if books.created != nil {
			t.Fatal("Repository.Create must not be called")
		}
	})

	t.Run("保存後の詳細に画像があればURLを詰める", func(t *testing.T) {
		t.Parallel()
		key := "books/1/a.png"
		uc := &command.CreateUsecaseImpl{
			Books: &fakeBooks{}, Authors: authorsExisting(1),
			Details: &fakeDetails{detail: &book.BookDetail{ImageKey: &key}},
			Images:  &fakeImages{url: "https://storage.example/"},
		}

		got, err := uc.Execute(context.Background(), base)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.ImageURL == nil || *got.ImageURL != "https://storage.example/books/1/a.png" {
			t.Fatalf("ImageURL = %v", got.ImageURL)
		}
	})
}
