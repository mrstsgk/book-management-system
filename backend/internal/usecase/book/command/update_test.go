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

func TestUpdateUsecase_ISBNAndCover(t *testing.T) {
	t.Parallel()
	withCatalogInfo := func(t *testing.T) *book.Book {
		t.Helper()
		b := existingBook(t, book.Unpublished)
		isbn, err := book.NewISBN("9784297146221")
		if err != nil {
			t.Fatal(err)
		}
		old, err := book.NewCover("https://cover.openbd.jp/old.jpg", book.CoverSourceOpenBD)
		if err != nil {
			t.Fatal(err)
		}
		b.ChangeCatalogInfo(&isbn, &old)
		return b
	}
	cmd := command.UpdateCommand{ID: 10, Title: "斜陽", Price: 800, AuthorIDs: []int64{2}, Status: 1, Version: 2}

	t.Run("保存のたびに書影を取り直す", func(t *testing.T) {
		t.Parallel()
		fresh := mustCover(t)
		books := &fakeBooks{findByID: withCatalogInfo(t)}
		catalog := &fakeCatalog{entry: &book.CatalogEntry{Cover: &fresh}}
		uc := &command.UpdateUsecaseImpl{Books: books, Authors: authorsExisting(2), Details: &fakeDetails{detail: &book.BookDetail{}}, Catalog: catalog}
		c := cmd
		c.ISBN = strPtr("9784297146221")

		if _, err := uc.Execute(context.Background(), c); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(catalog.called) != 1 || books.updated.Cover == nil || *books.updated.Cover != fresh {
			t.Fatalf("catalog called %v, cover %v; want one lookup and the fresh cover", catalog.called, books.updated.Cover)
		}
	})

	t.Run("ISBNを省略するとISBNも書影も解除する", func(t *testing.T) {
		t.Parallel()
		books := &fakeBooks{findByID: withCatalogInfo(t)}
		catalog := &fakeCatalog{}
		uc := &command.UpdateUsecaseImpl{Books: books, Authors: authorsExisting(2), Details: &fakeDetails{detail: &book.BookDetail{}}, Catalog: catalog}

		if _, err := uc.Execute(context.Background(), cmd); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if books.updated.ISBN != nil || books.updated.Cover != nil || len(catalog.called) != 0 {
			t.Fatalf("ISBN=%v Cover=%v lookups=%v, want both cleared without a lookup", books.updated.ISBN, books.updated.Cover, catalog.called)
		}
	})

	t.Run("不正なISBNはエラーで既存のISBNと書影も変わらない", func(t *testing.T) {
		t.Parallel()
		existing := withCatalogInfo(t)
		books := &fakeBooks{findByID: existing}
		uc := &command.UpdateUsecaseImpl{Books: books, Authors: authorsExisting(2), Details: &fakeDetails{}, Catalog: &fakeCatalog{}}
		c := cmd
		c.ISBN = strPtr("123")

		if _, err := uc.Execute(context.Background(), c); !errors.Is(err, common.ErrInvalid) {
			t.Fatalf("err = %v, want ErrInvalid", err)
		}
		if books.updated != nil || existing.ISBN.String() != "9784297146221" || existing.Cover.URL() != "https://cover.openbd.jp/old.jpg" {
			t.Fatalf("book changed on failure: %+v", existing)
		}
	})

	t.Run("同じISBNの別書籍があればConflictを返す", func(t *testing.T) {
		t.Parallel()
		books := &fakeBooks{findByID: withCatalogInfo(t), updateErr: common.ErrConflict}
		uc := &command.UpdateUsecaseImpl{Books: books, Authors: authorsExisting(2), Details: &fakeDetails{}, Catalog: &fakeCatalog{entry: &book.CatalogEntry{}}}
		c := cmd
		c.ISBN = strPtr("9784873118703")

		if _, err := uc.Execute(context.Background(), c); !errors.Is(err, common.ErrConflict) {
			t.Fatalf("err = %v, want ErrConflict", err)
		}
	})
}
