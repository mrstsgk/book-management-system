package query_test

import (
	"context"
	"errors"
	"testing"

	bookdomain "github.com/mrstsgk/book-management-system/backend/internal/book/domain"
	bookqry "github.com/mrstsgk/book-management-system/backend/internal/book/usecase/query"
	"github.com/mrstsgk/book-management-system/backend/internal/domainerr"
)

// fakeQuery は bookdomain.Query の手書き Fake（docs/rules/testing.md）。
type fakeQuery struct {
	detail   *bookdomain.BookDetail
	list     *bookdomain.BookList
	err      error
	gotID    bookdomain.ID
	gotRange *bookdomain.ListRange
	gotCond  *bookdomain.ListCondition
}

func (f *fakeQuery) FindDetailByID(_ context.Context, id bookdomain.ID) (*bookdomain.BookDetail, error) {
	f.gotID = id
	return f.detail, f.err
}

func (f *fakeQuery) ExistsByISBN(context.Context, bookdomain.ISBN) (bool, error) { return false, nil }

func (f *fakeQuery) Search(_ context.Context, c bookdomain.ListCondition, r bookdomain.ListRange) (*bookdomain.BookList, error) {
	f.gotCond = &c
	f.gotRange = &r
	return f.list, f.err
}

func TestReadBookUsecase_Execute(t *testing.T) {
	t.Parallel()

	t.Run("指定したIDの詳細を返す", func(t *testing.T) {
		t.Parallel()
		want := &bookdomain.BookDetail{ID: 3}
		q := &fakeQuery{detail: want}
		got, err := (&bookqry.ReadBookUsecaseImpl{Books: q}).Execute(context.Background(), 3)
		if err != nil || got != want || q.gotID != 3 {
			t.Fatalf("got (%+v, %v) for id %d", got, err, q.gotID)
		}
	})

	t.Run("存在しない本はNotFound", func(t *testing.T) {
		t.Parallel()
		if _, err := (&bookqry.ReadBookUsecaseImpl{Books: &fakeQuery{err: domainerr.ErrNotFound}}).Execute(context.Background(), 3); !errors.Is(err, domainerr.ErrNotFound) {
			t.Fatalf("err = %v, want ErrNotFound", err)
		}
	})
}
