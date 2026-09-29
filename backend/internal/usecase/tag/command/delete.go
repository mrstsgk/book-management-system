package command

import (
	"context"

	"github.com/mrstsgk/book-management-system/backend/internal/domain/tag"
)

type DeleteUsecase interface {
	Execute(ctx context.Context, id int64) error
}

type DeleteUsecaseImpl struct {
	Tags tag.Repository
}

func (u *DeleteUsecaseImpl) Execute(ctx context.Context, id int64) error {
	return u.Tags.Delete(ctx, tag.ID(id))
}
