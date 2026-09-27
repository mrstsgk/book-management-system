package query_test

import (
	"context"
	"errors"
	"io"
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

// fakeImages is a hand-written Fake for book.ImageStorage.
type fakeImages struct{ err error }

func (f *fakeImages) Put(context.Context, book.ImageKey, book.Image, io.Reader) error { return nil }
func (f *fakeImages) Delete(context.Context, book.ImageKey) error                     { return nil }
func (f *fakeImages) URL(_ context.Context, key book.ImageKey) (string, error) {
	return "https://storage.example/" + key.String(), f.err
}

func TestGetUsecase_Execute(t *testing.T) {
	t.Parallel()
	key := "books/1/a.png"

	t.Run("書籍の詳細を画像URL付きで返す", func(t *testing.T) {
		t.Parallel()
		uc := &query.GetUsecaseImpl{Books: &fakeDetailQuery{detail: &book.BookDetail{ID: 1, ImageKey: &key}}, Images: &fakeImages{}}

		got, err := uc.Execute(context.Background(), 1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.ID != 1 || got.ImageURL == nil || *got.ImageURL != "https://storage.example/"+key {
			t.Fatalf("got %+v", got)
		}
	})

	t.Run("存在しない書籍はNotFound", func(t *testing.T) {
		t.Parallel()
		uc := &query.GetUsecaseImpl{Books: &fakeDetailQuery{err: common.ErrNotFound}, Images: &fakeImages{}}

		if _, err := uc.Execute(context.Background(), 1); !errors.Is(err, common.ErrNotFound) {
			t.Fatalf("err = %v, want ErrNotFound", err)
		}
	})

	t.Run("URLの発行に失敗したらエラーを返す", func(t *testing.T) {
		t.Parallel()
		wantErr := errors.New("storage down")
		uc := &query.GetUsecaseImpl{Books: &fakeDetailQuery{detail: &book.BookDetail{ImageKey: &key}}, Images: &fakeImages{err: wantErr}}

		if _, err := uc.Execute(context.Background(), 1); !errors.Is(err, wantErr) {
			t.Fatalf("err = %v, want %v", err, wantErr)
		}
	})
}
