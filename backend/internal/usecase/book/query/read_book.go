package query

import (
	"context"

	"github.com/mrstsgk/book-management-system/backend/internal/domain/book"
)

type ReadBookUsecase interface {
	Execute(ctx context.Context, id int64) (*book.BookDetail, error)
}

type ReadBookUsecaseImpl struct {
	Books book.Query
}

func (u *ReadBookUsecaseImpl) Execute(ctx context.Context, id int64) (*book.BookDetail, error) {
	return u.Books.FindDetailByID(ctx, book.ID(id))
}
