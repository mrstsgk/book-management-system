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
