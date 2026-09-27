package command

import (
	"context"
	"fmt"

	"github.com/mrstsgk/book-management-system/backend/internal/domain/author"
	"github.com/mrstsgk/book-management-system/backend/internal/domain/book"
	"github.com/mrstsgk/book-management-system/backend/internal/domain/common"
)

// contents holds the validated values shared by the create and update commands.
type contents struct {
	title     book.Title
	price     book.Price
	authorIDs []author.ID
	status    book.PublishStatus
}

func newContents(title string, price int64, authorIDs []int64, status int) (contents, error) {
	t, err := book.NewTitle(title)
	if err != nil {
		return contents{}, err
	}
	p, err := book.NewPrice(price)
	if err != nil {
		return contents{}, err
	}
	s, err := book.NewPublishStatus(status)
	if err != nil {
		return contents{}, err
	}
	ids := make([]author.ID, 0, len(authorIDs))
	for _, id := range authorIDs {
		ids = append(ids, author.ID(id))
	}
	return contents{title: t, price: p, authorIDs: ids, status: s}, nil
}

// ensureAuthorsExist は指定された著者がすべて存在することを確認し、1人でも欠けていれば ErrInvalid を返す。
func ensureAuthorsExist(ctx context.Context, authors author.Repository, ids []author.ID) error {
	n, err := authors.CountByIDs(ctx, ids)
	if err != nil {
		return err
	}
	if n != len(ids) {
		return fmt.Errorf("%w: 存在しない著者が指定されています", common.ErrInvalid)
	}
	return nil
}
