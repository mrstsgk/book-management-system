package query_test

import (
	"context"
	"errors"
	"testing"

	"github.com/mrstsgk/book-management-system/backend/internal/domain/author"
	"github.com/mrstsgk/book-management-system/backend/internal/domain/book"
	"github.com/mrstsgk/book-management-system/backend/internal/domain/common"
	"github.com/mrstsgk/book-management-system/backend/internal/usecase/book/query"
)

// fakeAuthorQuery is a hand-written Fake for author.Query (docs/rules/testing.md).
type fakeAuthorQuery struct {
	exists bool
	err    error
}

func (f *fakeAuthorQuery) Exists(context.Context, author.ID) (bool, error) {
	return f.exists, f.err
}

// fakeBookQuery is a hand-written Fake for book.Query.
type fakeBookQuery struct {
	summaries []*book.BookSummary
	err       error
	called    bool
}

func (f *fakeBookQuery) FindDetailByID(context.Context, book.ID) (*book.BookDetail, error) {
	return nil, nil
}

func (f *fakeBookQuery) FindSummariesByAuthorID(context.Context, author.ID) ([]*book.BookSummary, error) {
	f.called = true
	return f.summaries, f.err
}

func TestListByAuthorUsecase_Execute(t *testing.T) {
	t.Parallel()

	t.Run("著者の書籍一覧を返す", func(t *testing.T) {
		t.Parallel()
		want := []*book.BookSummary{{ID: 1, Title: "人間失格"}}
		uc := &query.ListByAuthorUsecaseImpl{Authors: &fakeAuthorQuery{exists: true}, Books: &fakeBookQuery{summaries: want}}

		got, err := uc.Execute(context.Background(), 1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(got) != 1 || got[0] != want[0] {
			t.Fatalf("got %+v, want %+v", got, want)
		}
	})

	t.Run("存在しない著者はNotFoundで書籍は検索しない", func(t *testing.T) {
		t.Parallel()
		books := &fakeBookQuery{}
		uc := &query.ListByAuthorUsecaseImpl{Authors: &fakeAuthorQuery{exists: false}, Books: books}

		if _, err := uc.Execute(context.Background(), 99); !errors.Is(err, common.ErrNotFound) {
			t.Fatalf("err = %v, want ErrNotFound", err)
		}
		if books.called {
			t.Fatal("books must not be queried for an unknown author")
		}
	})

	t.Run("著者の存在確認のエラーはそのまま返る", func(t *testing.T) {
		t.Parallel()
		wantErr := errors.New("db down")
		uc := &query.ListByAuthorUsecaseImpl{Authors: &fakeAuthorQuery{err: wantErr}, Books: &fakeBookQuery{}}

		if _, err := uc.Execute(context.Background(), 1); !errors.Is(err, wantErr) {
			t.Fatalf("err = %v, want %v", err, wantErr)
		}
	})
}
