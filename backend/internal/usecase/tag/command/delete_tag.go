package command

import (
	"context"

	"github.com/mrstsgk/book-management-system/backend/internal/domain/tag"
)

type DeleteTagUsecase interface {
	Execute(ctx context.Context, id int64) error
}

type DeleteTagUsecaseImpl struct {
	Tags tag.Repository
}

func (u *DeleteTagUsecaseImpl) Execute(ctx context.Context, id int64) error {
	return u.Tags.Delete(ctx, tag.ID(id))
}
