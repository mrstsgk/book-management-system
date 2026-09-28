package query_test

import (
	"context"
	"errors"
	"testing"

	"github.com/mrstsgk/book-management-system/backend/internal/domain/book"
	"github.com/mrstsgk/book-management-system/backend/internal/domain/common"
	"github.com/mrstsgk/book-management-system/backend/internal/usecase/book/query"
)

func TestListUsecase_Execute(t *testing.T) {
	t.Parallel()

	t.Run("指定した取得範囲で一覧を返す", func(t *testing.T) {
		t.Parallel()
		want := &book.BookList{Total: 30}
		q := &fakeQuery{list: want}
		got, err := (&query.ListUsecaseImpl{Books: q}).Execute(context.Background(), 10, 20)
		if err != nil || got != want {
			t.Fatalf("got (%+v, %v)", got, err)
		}
		if q.gotRange == nil || q.gotRange.Limit() != 10 || q.gotRange.Offset() != 20 {
			t.Fatalf("query received %+v, want limit=10 offset=20", q.gotRange)
		}
	})

	for _, tt := range []struct {
		name          string
		limit, offset int
	}{
		{name: "limit上限超過", limit: common.MaxListLimit + 1},
		{name: "limit 0", limit: 0},
		{name: "offset負", limit: 10, offset: -1},
	} {
		t.Run(tt.name+"はエラーで検索しない", func(t *testing.T) {
			t.Parallel()
			q := &fakeQuery{}
			if _, err := (&query.ListUsecaseImpl{Books: q}).Execute(context.Background(), tt.limit, tt.offset); !errors.Is(err, common.ErrInvalid) {
				t.Fatalf("err = %v, want ErrInvalid", err)
			}
			if q.gotRange != nil {
				t.Fatal("the query must not run for an invalid range")
			}
		})
	}
}
