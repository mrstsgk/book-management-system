package query_test

import (
	"context"
	"errors"
	"testing"

	"github.com/mrstsgk/book-management-system/backend/internal/domain/book"
	"github.com/mrstsgk/book-management-system/backend/internal/domain/common"
	"github.com/mrstsgk/book-management-system/backend/internal/usecase/book/query"
)

// fakeQuery は book.Query の手書き Fake（docs/rules/testing.md）。
type fakeQuery struct {
	detail   *book.BookDetail
	list     *book.BookList
	err      error
	gotID    book.ID
	gotRange *common.ListRange
	gotCond  *book.ListCondition
}

func (f *fakeQuery) FindDetailByID(_ context.Context, id book.ID) (*book.BookDetail, error) {
	f.gotID = id
	return f.detail, f.err
}

func (f *fakeQuery) FindList(_ context.Context, c book.ListCondition, r common.ListRange) (*book.BookList, error) {
	f.gotCond = &c
	f.gotRange = &r
	return f.list, f.err
}

func TestGetUsecase_Execute(t *testing.T) {
	t.Parallel()

	t.Run("指定したIDの詳細を返す", func(t *testing.T) {
		t.Parallel()
		want := &book.BookDetail{ID: 3}
		q := &fakeQuery{detail: want}
		got, err := (&query.GetUsecaseImpl{Books: q}).Execute(context.Background(), 3)
		if err != nil || got != want || q.gotID != 3 {
			t.Fatalf("got (%+v, %v) for id %d", got, err, q.gotID)
		}
	})

	t.Run("存在しない本はNotFound", func(t *testing.T) {
		t.Parallel()
		if _, err := (&query.GetUsecaseImpl{Books: &fakeQuery{err: common.ErrNotFound}}).Execute(context.Background(), 3); !errors.Is(err, common.ErrNotFound) {
			t.Fatalf("err = %v, want ErrNotFound", err)
		}
	})
}
