package command

import (
	"context"
	"fmt"

	"github.com/mrstsgk/book-management-system/backend/internal/domain/author"
	"github.com/mrstsgk/book-management-system/backend/internal/domain/book"
	"github.com/mrstsgk/book-management-system/backend/internal/domain/common"
	"github.com/mrstsgk/book-management-system/backend/internal/usecase/book/imageurl"
)

// contents holds the validated values shared by the create and update commands.
type contents struct {
	title     book.Title
	price     book.Price
	authorIDs []author.ID
	status    book.PublishStatus
	amazonURL *book.AmazonURL
}

func newContents(title string, price int64, authorIDs []int64, status int, amazonURL *string) (contents, error) {
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
	var u *book.AmazonURL
	if amazonURL != nil {
		v, err := book.NewAmazonURL(*amazonURL)
		if err != nil {
			return contents{}, err
		}
		u = &v
	}
	ids := make([]author.ID, 0, len(authorIDs))
	for _, id := range authorIDs {
		ids = append(ids, author.ID(id))
	}
	return contents{title: t, price: p, authorIDs: ids, status: s, amazonURL: u}, nil
}

// detailOf は保存後の書籍の Read Model を取得し、画像があれば取得用 URL を詰める。
func detailOf(ctx context.Context, details book.Query, images book.ImageStorage, id book.ID) (*book.BookDetail, error) {
	d, err := details.FindDetailByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := imageurl.Fill(ctx, images, d); err != nil {
		return nil, err
	}
	return d, nil
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
