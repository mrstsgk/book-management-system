package query

import (
	"context"

	bookdomain "github.com/mrstsgk/book-management-system/backend/internal/book/domain"
)

// ReadBookUsecase は、読んだ本を 1 件参照する。
type ReadBookUsecase interface {
	Execute(ctx context.Context, id int64) (*bookdomain.BookDetail, error)
}

type ReadBookUsecaseImpl struct {
	Books bookdomain.Query
}

func (u *ReadBookUsecaseImpl) Execute(ctx context.Context, id int64) (*bookdomain.BookDetail, error) {
	return u.Books.FindDetailByID(ctx, bookdomain.ID(id))
}
