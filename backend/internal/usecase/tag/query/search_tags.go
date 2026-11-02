package query

import (
	"context"

	"github.com/mrstsgk/book-management-system/backend/internal/domain/tag"
)

type SearchTagsUsecase interface {
	Execute(ctx context.Context) (*tag.TagList, error)
}

type SearchTagsUsecaseImpl struct {
	Tags tag.Query
}

func (u *SearchTagsUsecaseImpl) Execute(ctx context.Context) (*tag.TagList, error) {
	return u.Tags.FindAll(ctx)
}
