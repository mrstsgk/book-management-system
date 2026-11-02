package query

import (
	"context"

	tagdomain "github.com/mrstsgk/book-management-system/backend/internal/tag/domain"
)

// CountBooksUsecase は、分野タグごとの冊数を数える。
type CountBooksUsecase interface {
	Execute(ctx context.Context) (*tagdomain.TagBookCounts, error)
}

type CountBooksUsecaseImpl struct {
	Tags tagdomain.Query
}

func (u *CountBooksUsecaseImpl) Execute(ctx context.Context) (*tagdomain.TagBookCounts, error) {
	return u.Tags.CountBooks(ctx)
}
