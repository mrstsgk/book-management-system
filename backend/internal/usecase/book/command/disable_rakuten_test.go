package command_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/mrstsgk/book-management-system/backend/internal/domain/book"
	"github.com/mrstsgk/book-management-system/backend/internal/domain/common"
	"github.com/mrstsgk/book-management-system/backend/internal/usecase/book/command"
)

// retryBooks は読むたびに新しい本を返し、保存の結果を順に返す Fake（競合からのやり直しを確かめるため）。
type retryBooks struct {
	fakeBooks
	newBook    func() *book.Book
	updateErrs []error
	finds      int
	updates    []*book.Book
}

func (f *retryBooks) FindByID(context.Context, book.ID) (*book.Book, error) {
	f.finds++
	return f.newBook(), nil
}

func (f *retryBooks) Update(_ context.Context, b *book.Book) error {
	f.updates = append(f.updates, b)
	if i := len(f.updates) - 1; i < len(f.updateErrs) {
		return f.updateErrs[i]
	}
	return nil
}

func rakutenCover(t *testing.T) *book.Cover {
	t.Helper()
	c, err := book.NewRakutenCover("https://thumbnail.image.rakuten.co.jp/1.jpg", "https://books.rakuten.co.jp/rb/1/", time.Now())
	if err != nil {
		t.Fatal(err)
	}
	return &c
}

func bookWithRakutenCover(t *testing.T) *book.Book {
	t.Helper()
	b := existingBook(t)
	b.RefreshCatalog(b.Bibliography, rakutenCover(t))
	return b
}

func TestDisableRakutenUsecase_Execute(t *testing.T) {
	t.Parallel()

	t.Run("楽天の書影を外して無効化し、読んだバージョンのまま保存する", func(t *testing.T) {
		t.Parallel()
		books := &retryBooks{newBook: func() *book.Book { return bookWithRakutenCover(t) }}

		if err := (&command.DisableRakutenUsecaseImpl{Books: books}).Execute(context.Background(), 10); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(books.updates) != 1 {
			t.Fatalf("updates = %d, want 1", len(books.updates))
		}
		u := books.updates[0]
		if !u.RakutenDisabled || u.Cover != nil || u.Version != 2 {
			t.Fatalf("saved disabled=%v cover=%v version=%d, want disabled, no cover, version 2", u.RakutenDisabled, u.Cover, u.Version)
		}
	})

	t.Run("保存が競合したら1回だけ読み直してやり直す", func(t *testing.T) {
		t.Parallel()
		books := &retryBooks{newBook: func() *book.Book { return bookWithRakutenCover(t) }, updateErrs: []error{common.ErrConflict}}

		if err := (&command.DisableRakutenUsecaseImpl{Books: books}).Execute(context.Background(), 10); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if books.finds != 2 || len(books.updates) != 2 {
			t.Fatalf("finds=%d updates=%d, want 2 and 2", books.finds, len(books.updates))
		}
	})

	t.Run("2回続けて競合したらConflictを返す", func(t *testing.T) {
		t.Parallel()
		books := &retryBooks{newBook: func() *book.Book { return bookWithRakutenCover(t) }, updateErrs: []error{common.ErrConflict, common.ErrConflict}}

		err := (&command.DisableRakutenUsecaseImpl{Books: books}).Execute(context.Background(), 10)
		if !errors.Is(err, common.ErrConflict) {
			t.Fatalf("err = %v, want ErrConflict", err)
		}
		if len(books.updates) != 2 {
			t.Fatalf("updates = %d, want 2 (no third attempt)", len(books.updates))
		}
	})

	t.Run("既に無効化済みで楽天の書影も無い本は保存せずに成功する", func(t *testing.T) {
		t.Parallel()
		books := &retryBooks{newBook: func() *book.Book {
			b := existingBook(t)
			b.DisableRakuten()
			return b
		}}

		if err := (&command.DisableRakutenUsecaseImpl{Books: books}).Execute(context.Background(), 10); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(books.updates) != 0 {
			t.Fatalf("updates = %d, want 0", len(books.updates))
		}
	})

	t.Run("存在しない本はNotFoundで保存しない", func(t *testing.T) {
		t.Parallel()
		books := &fakeBooks{findByIDErr: common.ErrNotFound}

		err := (&command.DisableRakutenUsecaseImpl{Books: books}).Execute(context.Background(), 10)
		if !errors.Is(err, common.ErrNotFound) {
			t.Fatalf("err = %v, want ErrNotFound", err)
		}
		if books.updated != nil {
			t.Fatal("Repository.Update must not be called")
		}
	})

	t.Run("保存のその他のエラーはそのまま返す", func(t *testing.T) {
		t.Parallel()
		wantErr := errors.New("db down")
		books := &retryBooks{newBook: func() *book.Book { return bookWithRakutenCover(t) }, updateErrs: []error{wantErr}}

		if err := (&command.DisableRakutenUsecaseImpl{Books: books}).Execute(context.Background(), 10); !errors.Is(err, wantErr) {
			t.Fatalf("err = %v, want %v", err, wantErr)
		}
		if len(books.updates) != 1 {
			t.Fatalf("updates = %d, want 1 (no retry on non-conflict errors)", len(books.updates))
		}
	})
}

func TestUpdateUsecase_KeepsRakutenDisabled(t *testing.T) {
	t.Parallel()
	existing := existingBook(t)
	existing.DisableRakuten()
	books := &fakeBooks{findByID: existing}
	uc := &command.UpdateUsecaseImpl{
		Books: books, Catalog: &fakeCatalog{entry: catalogEntry(t, "新しい書名", rakutenCover(t))},
		Details: &fakeDetails{detail: &book.BookDetail{}}, Tags: &fakeTagQuery{exists: true},
	}
	cmd := command.UpdateCommand{ID: 10, Summary: "まとめ", Comment: "感想", Rating: 4, Version: 2}

	if _, err := uc.Execute(context.Background(), cmd); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if books.updated == nil || books.updated.Cover != nil || !books.updated.RakutenDisabled {
		t.Fatalf("updated cover=%v disabled=%v, want no Rakuten cover even though the catalog returned one", books.updated.Cover, books.updated.RakutenDisabled)
	}
}
