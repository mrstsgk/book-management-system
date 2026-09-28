package query

import (
	"context"

	"github.com/mrstsgk/book-management-system/backend/internal/domain/book"
	"github.com/mrstsgk/book-management-system/backend/internal/domain/common"
)

type ListUsecase interface {
	Execute(ctx context.Context, limit, offset int) (*book.BookList, error)
}

type ListUsecaseImpl struct {
	Books book.ListQuery
}

func (u *ListUsecaseImpl) Execute(ctx context.Context, limit, offset int) (*book.BookList, error) {
	r, err := common.NewListRange(limit, offset)
	if err != nil {
		return nil, err
	}
	return u.Books.FindList(ctx, r)
}
