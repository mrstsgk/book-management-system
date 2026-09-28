package command

import (
	"context"

	"github.com/mrstsgk/book-management-system/backend/internal/domain/book"
)

type DeleteUsecase interface {
	Execute(ctx context.Context, id int64) error
}

type DeleteUsecaseImpl struct {
	Books book.Repository
}

func (u *DeleteUsecaseImpl) Execute(ctx context.Context, id int64) error {
	return u.Books.Delete(ctx, book.ID(id))
}
