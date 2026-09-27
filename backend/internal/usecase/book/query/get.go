package query

import (
	"context"

	"github.com/mrstsgk/book-management-system/backend/internal/domain/book"
	"github.com/mrstsgk/book-management-system/backend/internal/usecase/book/imageurl"
)

type GetUsecase interface {
	Execute(ctx context.Context, id int64) (*book.BookDetail, error)
}

type GetUsecaseImpl struct {
	Books  book.Query
	Images book.ImageStorage
}

func (u *GetUsecaseImpl) Execute(ctx context.Context, id int64) (*book.BookDetail, error) {
	d, err := u.Books.FindDetailByID(ctx, book.ID(id))
	if err != nil {
		return nil, err
	}
	if err := imageurl.Fill(ctx, u.Images, d); err != nil {
		return nil, err
	}
	return d, nil
}
