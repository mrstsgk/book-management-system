package command

import (
	"context"

	"github.com/mrstsgk/book-management-system/backend/internal/domain/book"
)

type DeleteBookUsecase interface {
	Execute(ctx context.Context, id int64) error
}

type DeleteBookUsecaseImpl struct {
	Books book.Repository
}

func (u *DeleteBookUsecaseImpl) Execute(ctx context.Context, id int64) error {
	return u.Books.Delete(ctx, book.ID(id))
}
