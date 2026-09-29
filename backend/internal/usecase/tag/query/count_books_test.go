package query_test

import (
	"context"
	"errors"
	"testing"

	"github.com/mrstsgk/book-management-system/backend/internal/domain/tag"
	"github.com/mrstsgk/book-management-system/backend/internal/usecase/tag/query"
)

func TestCountBooksUsecase_Execute(t *testing.T) {
	t.Parallel()

	t.Run("Queryの集計をそのまま返す", func(t *testing.T) {
		t.Parallel()
		want := &tag.TagBookCounts{Items: []*tag.TagBookCount{{ID: 1, Name: "設計", BookCount: 3}}}
		uc := &query.CountBooksUsecaseImpl{Tags: &fakeTagQuery{counts: want}}

		got, err := uc.Execute(context.Background())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != want {
			t.Fatalf("got %+v, want the query's result", got)
		}
	})

	t.Run("Queryの障害はそのまま返す", func(t *testing.T) {
		t.Parallel()
		wantErr := errors.New("db: timeout")
		uc := &query.CountBooksUsecaseImpl{Tags: &fakeTagQuery{countsErr: wantErr}}

		if _, err := uc.Execute(context.Background()); !errors.Is(err, wantErr) {
			t.Fatalf("err = %v, want %v", err, wantErr)
		}
	})
}
