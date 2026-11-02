package query

import (
	"context"

	bookdomain "github.com/mrstsgk/book-management-system/backend/internal/book/domain"
)

// LookupCatalogUsecase は登録前に、ISBN で外部カタログの書誌と書影を確かめるためのもの。
type LookupCatalogUsecase interface {
	Execute(ctx context.Context, isbn string) (*bookdomain.CatalogEntry, error)
}

type LookupCatalogUsecaseImpl struct {
	Catalog bookdomain.BookCatalog
}

func (u *LookupCatalogUsecaseImpl) Execute(ctx context.Context, isbn string) (*bookdomain.CatalogEntry, error) {
	v, err := bookdomain.NewISBN(isbn)
	if err != nil {
		return nil, err
	}
	return u.Catalog.Lookup(ctx, v)
}
