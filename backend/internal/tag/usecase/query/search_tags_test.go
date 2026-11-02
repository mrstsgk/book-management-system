package query_test

import (
	"context"
	"testing"

	tagdomain "github.com/mrstsgk/book-management-system/backend/internal/tag/domain"
	tagqry "github.com/mrstsgk/book-management-system/backend/internal/tag/usecase/query"
)

// fakeTagQuery は tagdomain.Query の手書き Fake。
type fakeTagQuery struct {
	list *tagdomain.TagList

	counts    *tagdomain.TagBookCounts
	countsErr error
}

func (f *fakeTagQuery) FindAll(context.Context) (*tagdomain.TagList, error) {
	return f.list, nil
}

func (f *fakeTagQuery) ExistsByName(context.Context, string, tagdomain.ID) (bool, error) {
	return false, nil
}

func (f *fakeTagQuery) ExistsAll(context.Context, []tagdomain.ID) (bool, error) {
	return true, nil
}

func (f *fakeTagQuery) CountBooks(context.Context) (*tagdomain.TagBookCounts, error) {
	return f.counts, f.countsErr
}

func TestSearchTagsUsecase_Execute(t *testing.T) {
	t.Parallel()
	want := &tagdomain.TagList{Items: []*tagdomain.TagListItem{{ID: 1, Name: "データベース"}}}
	uc := &tagqry.SearchTagsUsecaseImpl{Tags: &fakeTagQuery{list: want}}

	got, err := uc.Execute(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != want {
		t.Fatalf("got %+v, want the query's result", got)
	}
}
