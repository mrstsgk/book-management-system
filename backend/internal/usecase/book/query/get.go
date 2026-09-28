package query

import (
	"context"

	"github.com/mrstsgk/book-management-system/backend/internal/domain/book"
)

type GetUsecase interface {
	Execute(ctx context.Context, id int64) (*book.BookDetail, error)
}

type GetUsecaseImpl struct {
	Books book.Query
}

func (u *GetUsecaseImpl) Execute(ctx context.Context, id int64) (*book.BookDetail, error) {
	return u.Books.FindDetailByID(ctx, book.ID(id))
}
