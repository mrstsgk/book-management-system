package query

import (
	"context"

	"github.com/mrstsgk/book-management-system/backend/internal/domain/book"
	"github.com/mrstsgk/book-management-system/backend/internal/domain/common"
	"github.com/mrstsgk/book-management-system/backend/internal/domain/tag"
)

// SearchBooksInput は SearchBooksUsecase の入力（Presentation から渡る境界）。データだけを持ち、判定は VO が行う。
// 引数を並べると同じ int の limit・offset・タグIDを取り違えやすいため、構造体で受け取る。
type SearchBooksInput struct {
	Keyword string
	TagID   *int64
	Limit   int
	Offset  int
}

type SearchBooksUsecase interface {
	Execute(ctx context.Context, in SearchBooksInput) (*book.BookList, error)
}

type SearchBooksUsecaseImpl struct {
	Books book.Query
}

func (u *SearchBooksUsecaseImpl) Execute(ctx context.Context, in SearchBooksInput) (*book.BookList, error) {
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
	return u.Books.Search(ctx, c, r)
}
