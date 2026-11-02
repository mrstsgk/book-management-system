package command

import (
	"context"

	tagdomain "github.com/mrstsgk/book-management-system/backend/internal/tag/domain"
)

// DeleteTagUsecase は、分野タグを削除する。
type DeleteTagUsecase interface {
	Execute(ctx context.Context, id int64) error
}

type DeleteTagUsecaseImpl struct {
	Tags tagdomain.Repository
}

func (u *DeleteTagUsecaseImpl) Execute(ctx context.Context, id int64) error {
	return u.Tags.Delete(ctx, tagdomain.ID(id))
}
