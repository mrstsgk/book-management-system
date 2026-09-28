package query_test

import (
	"context"
	"errors"
	"testing"

	"github.com/mrstsgk/book-management-system/backend/internal/domain/book"
	"github.com/mrstsgk/book-management-system/backend/internal/domain/common"
	"github.com/mrstsgk/book-management-system/backend/internal/usecase/book/query"
)

// fakeCatalog は book.BookCatalog の手書き Fake。
type fakeCatalog struct {
	entry  *book.CatalogEntry
	err    error
	called []book.ISBN
}

func (f *fakeCatalog) Lookup(_ context.Context, isbn book.ISBN) (*book.CatalogEntry, error) {
	f.called = append(f.called, isbn)
	return f.entry, f.err
}

func TestLookupCatalogUsecase_Execute(t *testing.T) {
	t.Parallel()

	t.Run("ISBNを正規化してカタログを引く", func(t *testing.T) {
		t.Parallel()
		want := &book.CatalogEntry{}
		c := &fakeCatalog{entry: want}
		got, err := (&query.LookupCatalogUsecaseImpl{Catalog: c}).Execute(context.Background(), "978-4-87311-870-3")
		if err != nil || got != want {
			t.Fatalf("got (%+v, %v)", got, err)
		}
		if len(c.called) != 1 || c.called[0].String() != "9784873118703" {
			t.Fatalf("catalog called with %v", c.called)
		}
	})

	t.Run("不正なISBNはカタログを引かずにエラー", func(t *testing.T) {
		t.Parallel()
		c := &fakeCatalog{}
		if _, err := (&query.LookupCatalogUsecaseImpl{Catalog: c}).Execute(context.Background(), "abc"); !errors.Is(err, common.ErrInvalid) {
			t.Fatalf("err = %v, want ErrInvalid", err)
		}
		if len(c.called) != 0 {
			t.Fatal("the catalog must not be called for an invalid ISBN")
		}
	})

	t.Run("該当なしはNotFoundのまま返す", func(t *testing.T) {
		t.Parallel()
		if _, err := (&query.LookupCatalogUsecaseImpl{Catalog: &fakeCatalog{err: common.ErrNotFound}}).Execute(context.Background(), "9784873118703"); !errors.Is(err, common.ErrNotFound) {
			t.Fatalf("err = %v, want ErrNotFound", err)
		}
	})
}
