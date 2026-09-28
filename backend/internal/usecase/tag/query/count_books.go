package query

import (
	"context"

	"github.com/mrstsgk/book-management-system/backend/internal/domain/tag"
)

type CountBooksUsecase interface {
	Execute(ctx context.Context) (*tag.TagBookCounts, error)
}

type CountBooksUsecaseImpl struct {
	Tags tag.Query
}

func (u *CountBooksUsecaseImpl) Execute(ctx context.Context) (*tag.TagBookCounts, error) {
	return u.Tags.CountBooks(ctx)
}
