package query_test

import (
	"context"
	"errors"
	"testing"

	"github.com/mrstsgk/book-management-system/backend/internal/domain/book"
	"github.com/mrstsgk/book-management-system/backend/internal/domain/common"
	"github.com/mrstsgk/book-management-system/backend/internal/usecase/book/query"
)

// fakeCatalog is a hand-written Fake for book.BookCatalog.
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
		want := &book.CatalogEntry{Title: "データ指向アプリケーションデザイン"}
		catalog := &fakeCatalog{entry: want}
		uc := &query.LookupCatalogUsecaseImpl{Catalog: catalog}

		got, err := uc.Execute(context.Background(), "978-4-87311-870-3")
		if err != nil || got != want {
			t.Fatalf("got (%+v, %v), want (%+v, nil)", got, err, want)
		}
		if len(catalog.called) != 1 || catalog.called[0].String() != "9784873118703" {
			t.Fatalf("catalog called with %v, want the normalized ISBN-13", catalog.called)
		}
	})

	t.Run("不正なISBNはカタログを引かずにエラー", func(t *testing.T) {
		t.Parallel()
		catalog := &fakeCatalog{}
		uc := &query.LookupCatalogUsecaseImpl{Catalog: catalog}

		if _, err := uc.Execute(context.Background(), "abc"); !errors.Is(err, common.ErrInvalid) {
			t.Fatalf("err = %v, want ErrInvalid", err)
		}
		if len(catalog.called) != 0 {
			t.Fatal("the catalog must not be called for an invalid ISBN")
		}
	})

	t.Run("該当なしはNotFoundのまま返す", func(t *testing.T) {
		t.Parallel()
		uc := &query.LookupCatalogUsecaseImpl{Catalog: &fakeCatalog{err: common.ErrNotFound}}

		if _, err := uc.Execute(context.Background(), "9784873118703"); !errors.Is(err, common.ErrNotFound) {
			t.Fatalf("err = %v, want ErrNotFound", err)
		}
	})
}
