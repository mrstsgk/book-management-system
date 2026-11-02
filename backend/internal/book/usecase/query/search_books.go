package query

import (
	"context"

	bookdomain "github.com/mrstsgk/book-management-system/backend/internal/book/domain"
	tagdomain "github.com/mrstsgk/book-management-system/backend/internal/tag/domain"
)

// SearchBooksInput は SearchBooksUsecase の入力（Presentation から渡る境界）。データだけを持ち、判定は VO が行う。
// 引数を並べると同じ int の limit・offset・タグIDを取り違えやすいため、構造体で受け取る。
type SearchBooksInput struct {
	// 書名・著者の部分一致に使う語。
	Keyword string
	// 絞り込む分野タグの ID。nil なら絞り込まない。
	TagID *int64
	// 最大取得件数。0 なら既定値。
	Limit int
	// 取得開始位置。
	Offset int
}

// SearchBooksUsecase は、読んだ本を条件で探す。
type SearchBooksUsecase interface {
	Execute(ctx context.Context, in SearchBooksInput) (*bookdomain.BookList, error)
}

type SearchBooksUsecaseImpl struct {
	Books bookdomain.Query
}

func (u *SearchBooksUsecaseImpl) Execute(ctx context.Context, in SearchBooksInput) (*bookdomain.BookList, error) {
	r, err := bookdomain.NewListRange(in.Limit, in.Offset)
	if err != nil {
		return nil, err
	}
	var tagID *tagdomain.ID
	if in.TagID != nil {
		id := tagdomain.ID(*in.TagID)
		tagID = &id
	}
	c, err := bookdomain.NewListCondition(in.Keyword, tagID)
	if err != nil {
		return nil, err
	}
	return u.Books.Search(ctx, c, r)
}
