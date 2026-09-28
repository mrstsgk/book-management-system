package query

import (
	"context"

	"github.com/mrstsgk/book-management-system/backend/internal/domain/book"
)

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
