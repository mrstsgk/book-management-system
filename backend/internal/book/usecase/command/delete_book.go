package command

import (
	"context"

	bookdomain "github.com/mrstsgk/book-management-system/backend/internal/book/domain"
)

// DeleteBookUsecase は、読んだ本を削除する。
type DeleteBookUsecase interface {
	Execute(ctx context.Context, id int64) error
}

type DeleteBookUsecaseImpl struct {
	Books bookdomain.Repository
}

func (u *DeleteBookUsecaseImpl) Execute(ctx context.Context, id int64) error {
	return u.Books.Delete(ctx, bookdomain.ID(id))
}
