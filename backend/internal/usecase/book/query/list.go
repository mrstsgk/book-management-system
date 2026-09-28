package query

import (
	"context"

	"github.com/mrstsgk/book-management-system/backend/internal/domain/book"
	"github.com/mrstsgk/book-management-system/backend/internal/domain/common"
	"github.com/mrstsgk/book-management-system/backend/internal/domain/tag"
)

// ListInput は ListUsecase の入力（Presentation から渡る境界）。データだけを持ち、判定は VO が行う。
// 引数を並べると同じ int の limit・offset・タグIDを取り違えやすいため、構造体で受け取る。
type ListInput struct {
	Keyword string
	TagID   *int64
	Limit   int
	Offset  int
}

type ListUsecase interface {
	Execute(ctx context.Context, in ListInput) (*book.BookList, error)
}

type ListUsecaseImpl struct {
	Books book.Query
}

func (u *ListUsecaseImpl) Execute(ctx context.Context, in ListInput) (*book.BookList, error) {
	r, err := common.NewListRange(in.Limit, in.Offset)
	if err != nil {
		return nil, err
	}
	var tagID *tag.ID
	if in.TagID != nil {
		id := tag.ID(*in.TagID)
		tagID = &id
	}
	c, err := book.NewListCondition(in.Keyword, tagID)
	if err != nil {
		return nil, err
	}
	return u.Books.FindList(ctx, c, r)
}
