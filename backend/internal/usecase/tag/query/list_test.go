package query_test

import (
	"context"
	"testing"

	"github.com/mrstsgk/book-management-system/backend/internal/domain/tag"
	"github.com/mrstsgk/book-management-system/backend/internal/usecase/tag/query"
)

// fakeTagQuery は tag.Query の手書き Fake。
type fakeTagQuery struct {
	list *tag.TagList
}

func (f *fakeTagQuery) FindList(context.Context) (*tag.TagList, error) {
	return f.list, nil
}

func (f *fakeTagQuery) ExistsAll(context.Context, []tag.ID) (bool, error) {
	return true, nil
}

func TestListUsecase_Execute(t *testing.T) {
	t.Parallel()
	want := &tag.TagList{Items: []*tag.TagListItem{{ID: 1, Name: "データベース"}}}
	uc := &query.ListUsecaseImpl{Tags: &fakeTagQuery{list: want}}

	got, err := uc.Execute(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != want {
		t.Fatalf("got %+v, want the query's result", got)
	}
}
