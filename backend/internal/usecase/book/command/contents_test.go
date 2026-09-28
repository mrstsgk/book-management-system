package command_test

import (
	"context"
	"errors"
	"testing"

	"github.com/mrstsgk/book-management-system/backend/internal/domain/author"
	"github.com/mrstsgk/book-management-system/backend/internal/domain/book"
	"github.com/mrstsgk/book-management-system/backend/internal/domain/common"
	"github.com/mrstsgk/book-management-system/backend/internal/usecase/book/command"
)

// fakeBooks is a hand-written Fake for book.Repository (docs/rules/testing.md).
type fakeBooks struct {
	created   *book.Book
	createErr error

	findByID    *book.Book
	findByIDErr error

	updated   *book.Book
	updateErr error
}

func (f *fakeBooks) FindByID(context.Context, book.ID) (*book.Book, error) {
	return f.findByID, f.findByIDErr
}

func (f *fakeBooks) Create(_ context.Context, b *book.Book) error {
	if f.createErr != nil {
		return f.createErr
	}
	b.ID, b.Version = 1, 1
	f.created = b
	return nil
}

func (f *fakeBooks) Update(_ context.Context, b *book.Book) error {
	if f.updateErr != nil {
		return f.updateErr
	}
	b.Version++
	f.updated = b
	return nil
}

// fakeAuthors is a hand-written Fake for author.Repository; only CountByIDs is used here.
type fakeAuthors struct {
	existing map[author.ID]bool
	countErr error
}

func (f *fakeAuthors) FindByID(context.Context, author.ID) (*author.Author, error) {
	return nil, common.ErrNotFound
}

func (f *fakeAuthors) CountByIDs(_ context.Context, ids []author.ID) (int, error) {
	if f.countErr != nil {
		return 0, f.countErr
	}
	n := 0
	for _, id := range ids {
		if f.existing[id] {
			n++
		}
	}
	return n, nil
}

func (f *fakeAuthors) Create(context.Context, *author.Author) error { return nil }
func (f *fakeAuthors) Update(context.Context, *author.Author) error { return nil }

// fakeDetails is a hand-written Fake for book.Query.
type fakeDetails struct {
	detail *book.BookDetail
	err    error
	gotID  book.ID
}

func (f *fakeDetails) FindDetailByID(_ context.Context, id book.ID) (*book.BookDetail, error) {
	f.gotID = id
	return f.detail, f.err
}

func (f *fakeDetails) FindSummariesByAuthorID(context.Context, author.ID) ([]*book.BookSummary, error) {
	return nil, nil
}

func authorsExisting(ids ...author.ID) *fakeAuthors {
	m := map[author.ID]bool{}
	for _, id := range ids {
		m[id] = true
	}
	return &fakeAuthors{existing: m}
}

// The shared input checks are exercised through CreateUsecase, the public entry point.
func TestCreateUsecase_RejectsInvalidContentsBeforeSaving(t *testing.T) {
	t.Parallel()
	valid := command.CreateCommand{Title: "人間失格", Price: 1500, AuthorIDs: []int64{1}, Status: 1}
	tests := []struct {
		name   string
		mutate func(*command.CreateCommand)
	}{
		{name: "空白を含むタイトル", mutate: func(c *command.CreateCommand) { c.Title = "人間 失格" }},
		{name: "負の価格", mutate: func(c *command.CreateCommand) { c.Price = -1 }},
		{name: "上限超過の価格", mutate: func(c *command.CreateCommand) { c.Price = 100_000_000 }},
		{name: "不正な出版状況", mutate: func(c *command.CreateCommand) { c.Status = 3 }},
		{name: "著者0人", mutate: func(c *command.CreateCommand) { c.AuthorIDs = nil }},
		{name: "著者IDの重複", mutate: func(c *command.CreateCommand) { c.AuthorIDs = []int64{1, 1} }},
		{name: "存在しない著者", mutate: func(c *command.CreateCommand) { c.AuthorIDs = []int64{1, 999} }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cmd := valid
			cmd.AuthorIDs = append([]int64(nil), valid.AuthorIDs...)
			tt.mutate(&cmd)
			books := &fakeBooks{}
			uc := &command.CreateUsecaseImpl{Books: books, Authors: authorsExisting(1), Details: &fakeDetails{}}

			if _, err := uc.Execute(context.Background(), cmd); !errors.Is(err, common.ErrInvalid) {
				t.Fatalf("err = %v, want ErrInvalid", err)
			}
			if books.created != nil {
				t.Fatal("Repository.Create must not be called when validation fails")
			}
		})
	}
}

func TestCreateUsecase_AuthorLookupErrorPropagates(t *testing.T) {
	t.Parallel()
	wantErr := errors.New("db down")
	books := &fakeBooks{}
	uc := &command.CreateUsecaseImpl{Books: books, Authors: &fakeAuthors{countErr: wantErr}, Details: &fakeDetails{}}

	_, err := uc.Execute(context.Background(), command.CreateCommand{Title: "人間失格", Price: 1500, AuthorIDs: []int64{1}, Status: 1})
	if !errors.Is(err, wantErr) {
		t.Fatalf("err = %v, want %v", err, wantErr)
	}
	if books.created != nil {
		t.Fatal("Repository.Create must not be called when the author lookup fails")
	}
}

// fakeCatalog is a hand-written Fake for book.BookCatalog.
type fakeCatalog struct {
	entry  *book.CatalogEntry
	err    error
	called []book.ISBN
}

func (f *fakeCatalog) Lookup(_ context.Context, isbn book.ISBN) (*book.CatalogEntry, error) {
	f.called = append(f.called, isbn)
	return f.entry, f.err
}

func mustCover(t *testing.T) book.Cover {
	t.Helper()
	c, err := book.NewCover("https://cover.openbd.jp/9784873118703.jpg", book.CoverSourceOpenBD)
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func strPtr(s string) *string { return &s }

// The ISBN handling and the cover lookup are shared by create and update; exercised through CreateUsecase.
func TestCreateUsecase_ISBNAndCover(t *testing.T) {
	t.Parallel()
	base := command.CreateCommand{Title: "データ指向アプリケーションデザイン", Price: 4600, AuthorIDs: []int64{1}, Status: 2}

	t.Run("ISBNを正規化して保存し、カタログの書影を付ける", func(t *testing.T) {
		t.Parallel()
		cover := mustCover(t)
		books := &fakeBooks{}
		catalog := &fakeCatalog{entry: &book.CatalogEntry{Cover: &cover}}
		uc := &command.CreateUsecaseImpl{Books: books, Authors: authorsExisting(1), Details: &fakeDetails{detail: &book.BookDetail{}}, Catalog: catalog}
		cmd := base
		cmd.ISBN = strPtr("4873118700")

		if _, err := uc.Execute(context.Background(), cmd); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if books.created.ISBN == nil || books.created.ISBN.String() != "9784873118703" {
			t.Fatalf("ISBN = %v, want the ISBN-13 form", books.created.ISBN)
		}
		if books.created.Cover == nil || *books.created.Cover != cover {
			t.Fatalf("Cover = %v, want %v", books.created.Cover, cover)
		}
		if len(catalog.called) != 1 || catalog.called[0].String() != "9784873118703" {
			t.Fatalf("catalog called with %v", catalog.called)
		}
	})

	covers := []struct {
		name    string
		catalog *fakeCatalog
	}{
		{name: "カタログに書影が無ければ書影なしで保存する", catalog: &fakeCatalog{entry: &book.CatalogEntry{}}},
		{name: "カタログに該当が無ければ書影なしで保存する", catalog: &fakeCatalog{err: common.ErrNotFound}},
		{name: "カタログの障害でも登録は止めず書影なしで保存する", catalog: &fakeCatalog{err: errors.New("openbd: timeout")}},
	}
	for _, tt := range covers {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			books := &fakeBooks{}
			uc := &command.CreateUsecaseImpl{Books: books, Authors: authorsExisting(1), Details: &fakeDetails{detail: &book.BookDetail{}}, Catalog: tt.catalog}
			cmd := base
			cmd.ISBN = strPtr("9784873118703")

			if _, err := uc.Execute(context.Background(), cmd); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if books.created == nil || books.created.ISBN == nil || books.created.Cover != nil {
				t.Fatalf("created %+v, want the ISBN kept and no cover", books.created)
			}
		})
	}

	t.Run("ISBNが無ければカタログを引かない", func(t *testing.T) {
		t.Parallel()
		books := &fakeBooks{}
		catalog := &fakeCatalog{}
		uc := &command.CreateUsecaseImpl{Books: books, Authors: authorsExisting(1), Details: &fakeDetails{detail: &book.BookDetail{}}, Catalog: catalog}

		if _, err := uc.Execute(context.Background(), base); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(catalog.called) != 0 || books.created.ISBN != nil || books.created.Cover != nil {
			t.Fatalf("catalog called %v, created %+v; want no lookup and no ISBN/cover", catalog.called, books.created)
		}
	})

	t.Run("不正なISBNは保存せずカタログも引かずにエラー", func(t *testing.T) {
		t.Parallel()
		books := &fakeBooks{}
		catalog := &fakeCatalog{}
		uc := &command.CreateUsecaseImpl{Books: books, Authors: authorsExisting(1), Details: &fakeDetails{}, Catalog: catalog}
		cmd := base
		cmd.ISBN = strPtr("9784873118704")

		if _, err := uc.Execute(context.Background(), cmd); !errors.Is(err, common.ErrInvalid) {
			t.Fatalf("err = %v, want ErrInvalid", err)
		}
		if books.created != nil || len(catalog.called) != 0 {
			t.Fatal("neither the repository nor the catalog may be called for an invalid ISBN")
		}
	})
}
