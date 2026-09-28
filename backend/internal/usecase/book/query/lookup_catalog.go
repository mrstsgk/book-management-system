package query

import (
	"context"

	"github.com/mrstsgk/book-management-system/backend/internal/domain/book"
)

// LookupCatalogUsecase は登録前に、ISBN で外部カタログの書誌と書影を確かめるためのもの。
type LookupCatalogUsecase interface {
	Execute(ctx context.Context, isbn string) (*book.CatalogEntry, error)
}

type LookupCatalogUsecaseImpl struct {
	Catalog book.BookCatalog
}

func (u *LookupCatalogUsecaseImpl) Execute(ctx context.Context, isbn string) (*book.CatalogEntry, error) {
	v, err := book.NewISBN(isbn)
	if err != nil {
		return nil, err
	}
	return u.Catalog.Lookup(ctx, v)
}
