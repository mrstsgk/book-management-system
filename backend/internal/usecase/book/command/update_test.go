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

func existingBook(t *testing.T, status book.PublishStatus) *book.Book {
	t.Helper()
	title, err := book.NewTitle("人間失格")
	if err != nil {
		t.Fatal(err)
	}
	price, err := book.NewPrice(1500)
	if err != nil {
		t.Fatal(err)
	}
	b, err := book.New(title, price, []author.ID{1}, status)
	if err != nil {
		t.Fatal(err)
	}
	b.ID, b.Version = 10, 2
	return b
}

func TestUpdateUsecase_Execute(t *testing.T) {
	t.Parallel()
	cmd := command.UpdateCommand{ID: 10, Title: "斜陽", Price: 800, AuthorIDs: []int64{2}, Status: 2, Version: 2}

	t.Run("内容を差し替えて保存し詳細を返す", func(t *testing.T) {
		t.Parallel()
		books := &fakeBooks{findByID: existingBook(t, book.Unpublished)}
		want := &book.BookDetail{ID: 10}
		details := &fakeDetails{detail: want}
		uc := &command.UpdateUsecaseImpl{Books: books, Authors: authorsExisting(2), Details: details}

		got, err := uc.Execute(context.Background(), cmd)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		u := books.updated
		if u == nil {
			t.Fatal("Repository.Update was not called")
		}
		if u.Title.String() != "斜陽" || u.Price.Int64() != 800 || u.Status != book.Published ||
			!reflect.DeepEqual(u.AuthorIDs, []author.ID{2}) {
			t.Fatalf("updated %+v", u)
		}
		if details.gotID != 10 || got != want {
			t.Fatalf("detail looked up for id=%d, got %+v", details.gotID, got)
		}
	})

	t.Run("存在しない書籍はNotFoundを返す", func(t *testing.T) {
		t.Parallel()
		books := &fakeBooks{findByIDErr: common.ErrNotFound}
		uc := &command.UpdateUsecaseImpl{Books: books, Authors: authorsExisting(2), Details: &fakeDetails{}}

		if _, err := uc.Execute(context.Background(), cmd); !errors.Is(err, common.ErrNotFound) {
			t.Fatalf("err = %v, want ErrNotFound", err)
		}
		if books.updated != nil {
			t.Fatal("Repository.Update must not be called")
		}
	})

	t.Run("出版済みから未出版への変更はエラーで保存しない", func(t *testing.T) {
		t.Parallel()
		books := &fakeBooks{findByID: existingBook(t, book.Published)}
		uc := &command.UpdateUsecaseImpl{Books: books, Authors: authorsExisting(2), Details: &fakeDetails{}}
		c := cmd
		c.Status = 1

		if _, err := uc.Execute(context.Background(), c); !errors.Is(err, common.ErrInvalid) {
			t.Fatalf("err = %v, want ErrInvalid", err)
		}
		if books.updated != nil {
			t.Fatal("Repository.Update must not be called")
		}
	})

	t.Run("存在しない著者はエラーで保存しない", func(t *testing.T) {
		t.Parallel()
		books := &fakeBooks{findByID: existingBook(t, book.Unpublished)}
		uc := &command.UpdateUsecaseImpl{Books: books, Authors: authorsExisting(), Details: &fakeDetails{}}

		if _, err := uc.Execute(context.Background(), cmd); !errors.Is(err, common.ErrInvalid) {
			t.Fatalf("err = %v, want ErrInvalid", err)
		}
		if books.updated != nil {
			t.Fatal("Repository.Update must not be called")
		}
	})

	t.Run("楽観的ロックの競合はConflictを返し詳細は取得しない", func(t *testing.T) {
		t.Parallel()
		books := &fakeBooks{findByID: existingBook(t, book.Unpublished), updateErr: common.ErrConflict}
		details := &fakeDetails{}
		uc := &command.UpdateUsecaseImpl{Books: books, Authors: authorsExisting(2), Details: details}

		if _, err := uc.Execute(context.Background(), cmd); !errors.Is(err, common.ErrConflict) {
			t.Fatalf("err = %v, want ErrConflict", err)
		}
		if details.gotID != 0 {
			t.Fatal("the detail must not be queried when saving fails")
		}
	})
}

func TestUpdateUsecase_AmazonURL(t *testing.T) {
	t.Parallel()
	withURL := func(t *testing.T) *book.Book {
		t.Helper()
		b := existingBook(t, book.Unpublished)
		u, err := book.NewAmazonURL("https://www.amazon.co.jp/dp/old")
		if err != nil {
			t.Fatal(err)
		}
		b.ChangeAmazonURL(&u)
		return b
	}
	cmd := command.UpdateCommand{ID: 10, Title: "斜陽", Price: 800, AuthorIDs: []int64{2}, Status: 1, Version: 2}

	t.Run("新しいURLに差し替える", func(t *testing.T) {
		t.Parallel()
		books := &fakeBooks{findByID: withURL(t)}
		uc := &command.UpdateUsecaseImpl{Books: books, Authors: authorsExisting(2), Details: &fakeDetails{detail: &book.BookDetail{}}}
		c := cmd
		u := "https://amzn.asia/d/new"
		c.AmazonURL = &u

		if _, err := uc.Execute(context.Background(), c); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if books.updated.AmazonURL == nil || books.updated.AmazonURL.String() != u {
			t.Fatalf("AmazonURL = %v, want %s", books.updated.AmazonURL, u)
		}
	})

	t.Run("URLを省略すると解除する", func(t *testing.T) {
		t.Parallel()
		books := &fakeBooks{findByID: withURL(t)}
		uc := &command.UpdateUsecaseImpl{Books: books, Authors: authorsExisting(2), Details: &fakeDetails{detail: &book.BookDetail{}}}

		if _, err := uc.Execute(context.Background(), cmd); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if books.updated.AmazonURL != nil {
			t.Fatalf("AmazonURL = %v, want nil", books.updated.AmazonURL)
		}
	})

	t.Run("不正なURLはエラーで既存のURLも変わらない", func(t *testing.T) {
		t.Parallel()
		existing := withURL(t)
		books := &fakeBooks{findByID: existing}
		uc := &command.UpdateUsecaseImpl{Books: books, Authors: authorsExisting(2), Details: &fakeDetails{}}
		c := cmd
		u := "http://www.amazon.co.jp/dp/insecure"
		c.AmazonURL = &u

		if _, err := uc.Execute(context.Background(), c); !errors.Is(err, common.ErrInvalid) {
			t.Fatalf("err = %v, want ErrInvalid", err)
		}
		if books.updated != nil || existing.AmazonURL.String() != "https://www.amazon.co.jp/dp/old" {
			t.Fatalf("book changed on failure: %+v", existing)
		}
	})
}
