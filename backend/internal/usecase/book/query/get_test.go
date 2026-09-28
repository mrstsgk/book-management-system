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

// fakeDetailQuery is a hand-written Fake for book.Query.
type fakeDetailQuery struct {
	detail *book.BookDetail
	err    error
}

func (f *fakeDetailQuery) FindDetailByID(context.Context, book.ID) (*book.BookDetail, error) {
	return f.detail, f.err
}

func (f *fakeDetailQuery) FindSummariesByAuthorID(context.Context, author.ID) ([]*book.BookSummary, error) {
	return nil, nil
}

func TestGetUsecase_Execute(t *testing.T) {
	t.Parallel()

	t.Run("書籍の詳細を返す", func(t *testing.T) {
		t.Parallel()
		want := &book.BookDetail{ID: 1}
		uc := &query.GetUsecaseImpl{Books: &fakeDetailQuery{detail: want}}

		got, err := uc.Execute(context.Background(), 1)
		if err != nil || got != want {
			t.Fatalf("got (%+v, %v), want (%+v, nil)", got, err, want)
		}
	})

	t.Run("存在しない書籍はNotFound", func(t *testing.T) {
		t.Parallel()
		uc := &query.GetUsecaseImpl{Books: &fakeDetailQuery{err: common.ErrNotFound}}

		if _, err := uc.Execute(context.Background(), 1); !errors.Is(err, common.ErrNotFound) {
			t.Fatalf("err = %v, want ErrNotFound", err)
		}
	})
}
