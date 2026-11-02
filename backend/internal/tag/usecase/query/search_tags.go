package query

import (
	"context"

	tagdomain "github.com/mrstsgk/book-management-system/backend/internal/tag/domain"
)

// SearchTagsUsecase は、分野タグの一覧を参照する。
type SearchTagsUsecase interface {
	Execute(ctx context.Context) (*tagdomain.TagList, error)
}

type SearchTagsUsecaseImpl struct {
	Tags tagdomain.Query
}

func (u *SearchTagsUsecaseImpl) Execute(ctx context.Context) (*tagdomain.TagList, error) {
	return u.Tags.FindAll(ctx)
}
