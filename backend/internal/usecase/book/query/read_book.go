package query

import (
	"context"

	"github.com/mrstsgk/book-management-system/backend/internal/domain/book"
)

// ReadBookUsecase は、読んだ本を 1 件参照する。
type ReadBookUsecase interface {
	Execute(ctx context.Context, id int64) (*book.BookDetail, error)
}

type ReadBookUsecaseImpl struct {
	Books book.Query
}

func (u *ReadBookUsecaseImpl) Execute(ctx context.Context, id int64) (*book.BookDetail, error) {
	return u.Books.FindDetailByID(ctx, book.ID(id))
}
