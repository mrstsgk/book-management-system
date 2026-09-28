package query_test

import (
	"context"
	"errors"
	"testing"

	"github.com/mrstsgk/book-management-system/backend/internal/domain/book"
	"github.com/mrstsgk/book-management-system/backend/internal/domain/common"
	"github.com/mrstsgk/book-management-system/backend/internal/usecase/book/query"
)

// fakeListQuery is a hand-written Fake for book.ListQuery (docs/rules/testing.md).
type fakeListQuery struct {
	list     *book.BookList
	err      error
	gotRange *common.ListRange
}

func (f *fakeListQuery) FindList(_ context.Context, r common.ListRange) (*book.BookList, error) {
	f.gotRange = &r
	return f.list, f.err
}

func TestListUsecase_Execute(t *testing.T) {
	t.Parallel()

	t.Run("指定した取得範囲で一覧を返す", func(t *testing.T) {
		t.Parallel()
		want := &book.BookList{Items: []*book.BookListItem{{ID: 1, Title: "人間失格"}}, Total: 30}
		books := &fakeListQuery{list: want}
		uc := &query.ListUsecaseImpl{Books: books}

		got, err := uc.Execute(context.Background(), 10, 20)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != want {
			t.Fatalf("got %+v, want %+v", got, want)
		}
		if books.gotRange == nil || books.gotRange.Limit() != 10 || books.gotRange.Offset() != 20 {
			t.Fatalf("query received range %+v, want limit=10 offset=20", books.gotRange)
		}
	})

	invalid := []struct {
		name          string
		limit, offset int
	}{
		{name: "limit上限超過はエラーで検索しない", limit: common.MaxListLimit + 1, offset: 0},
		{name: "limit 0 はエラーで検索しない", limit: 0, offset: 0},
		{name: "offset負はエラーで検索しない", limit: 10, offset: -1},
	}
	for _, tt := range invalid {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			books := &fakeListQuery{}
			uc := &query.ListUsecaseImpl{Books: books}

			if _, err := uc.Execute(context.Background(), tt.limit, tt.offset); !errors.Is(err, common.ErrInvalid) {
				t.Fatalf("err = %v, want ErrInvalid", err)
			}
			if books.gotRange != nil {
				t.Fatal("the query must not run for an invalid range")
			}
		})
	}

	t.Run("Queryのエラーはそのまま返る", func(t *testing.T) {
		t.Parallel()
		wantErr := errors.New("db down")
		uc := &query.ListUsecaseImpl{Books: &fakeListQuery{err: wantErr}}

		if _, err := uc.Execute(context.Background(), 10, 0); !errors.Is(err, wantErr) {
			t.Fatalf("err = %v, want %v", err, wantErr)
		}
	})
}
