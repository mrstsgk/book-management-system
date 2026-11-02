package query_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	bookdomain "github.com/mrstsgk/book-management-system/backend/internal/book/domain"
	bookqry "github.com/mrstsgk/book-management-system/backend/internal/book/usecase/query"
	"github.com/mrstsgk/book-management-system/backend/internal/domainerr"
)

func int64Ptr(v int64) *int64 { return &v }

func TestSearchBooksUsecase_Execute(t *testing.T) {
	t.Parallel()

	t.Run("指定した取得範囲で一覧を返す", func(t *testing.T) {
		t.Parallel()
		want := &bookdomain.BookList{Total: 30}
		q := &fakeQuery{list: want}
		got, err := (&bookqry.SearchBooksUsecaseImpl{Books: q}).Execute(context.Background(), bookqry.SearchBooksInput{Limit: 10, Offset: 20})
		if err != nil || got != want {
			t.Fatalf("got (%+v, %v)", got, err)
		}
		if q.gotRange == nil || q.gotRange.Limit() != 10 || q.gotRange.Offset() != 20 {
			t.Fatalf("query received %+v, want limit=10 offset=20", q.gotRange)
		}
		if _, ok := q.gotCond.TagID(); ok || q.gotCond.Keyword() != "" {
			t.Fatalf("condition = %+v, want no filter when neither keyword nor tag is given", q.gotCond)
		}
	})

	t.Run("キーワードと分野タグを条件として渡す", func(t *testing.T) {
		t.Parallel()
		q := &fakeQuery{list: &bookdomain.BookList{}}
		in := bookqry.SearchBooksInput{Keyword: " 設計 ", TagID: int64Ptr(3), Limit: 20}
		if _, err := (&bookqry.SearchBooksUsecaseImpl{Books: q}).Execute(context.Background(), in); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if q.gotCond.Keyword() != "設計" {
			t.Fatalf("Keyword = %q, want 設計 (trimmed)", q.gotCond.Keyword())
		}
		if id, ok := q.gotCond.TagID(); !ok || id != 3 {
			t.Fatalf("TagID = (%d, %v), want (3, true)", id, ok)
		}
	})

	for _, tt := range []struct {
		name string
		in   bookqry.SearchBooksInput
	}{
		{name: "limit上限超過", in: bookqry.SearchBooksInput{Limit: bookdomain.MaxListLimit + 1}},
		{name: "limit 0", in: bookqry.SearchBooksInput{Limit: 0}},
		{name: "offset負", in: bookqry.SearchBooksInput{Limit: 10, Offset: -1}},
		{name: "キーワード101文字", in: bookqry.SearchBooksInput{Keyword: strings.Repeat("あ", 101), Limit: 10}},
		{name: "タグID 0", in: bookqry.SearchBooksInput{TagID: int64Ptr(0), Limit: 10}},
	} {
		t.Run(tt.name+"はエラーで検索しない", func(t *testing.T) {
			t.Parallel()
			q := &fakeQuery{}
			if _, err := (&bookqry.SearchBooksUsecaseImpl{Books: q}).Execute(context.Background(), tt.in); !errors.Is(err, domainerr.ErrInvalid) {
				t.Fatalf("err = %v, want ErrInvalid", err)
			}
			if q.gotRange != nil {
				t.Fatal("the query must not run for invalid input")
			}
		})
	}

	t.Run("クエリのエラーはそのまま返す", func(t *testing.T) {
		t.Parallel()
		wantErr := errors.New("db: timeout")
		q := &fakeQuery{err: wantErr}
		if _, err := (&bookqry.SearchBooksUsecaseImpl{Books: q}).Execute(context.Background(), bookqry.SearchBooksInput{Limit: 10}); !errors.Is(err, wantErr) {
			t.Fatalf("err = %v, want %v", err, wantErr)
		}
	})
}
