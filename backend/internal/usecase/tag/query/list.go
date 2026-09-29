package query

import (
	"context"

	"github.com/mrstsgk/book-management-system/backend/internal/domain/tag"
)

type ListUsecase interface {
	Execute(ctx context.Context) (*tag.TagList, error)
}

type ListUsecaseImpl struct {
	Tags tag.Query
}

func (u *ListUsecaseImpl) Execute(ctx context.Context) (*tag.TagList, error) {
	return u.Tags.FindList(ctx)
}
