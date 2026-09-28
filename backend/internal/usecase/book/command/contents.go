package command

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

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
	isbn      *book.ISBN
}

func newContents(title string, price int64, authorIDs []int64, status int, isbn *string) (contents, error) {
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
	var i *book.ISBN
	if isbn != nil {
		v, err := book.NewISBN(*isbn)
		if err != nil {
			return contents{}, err
		}
		i = &v
	}
	ids := make([]author.ID, 0, len(authorIDs))
	for _, id := range authorIDs {
		ids = append(ids, author.ID(id))
	}
	return contents{title: t, price: p, authorIDs: ids, status: s, isbn: i}, nil
}

// coverFor は ISBN で外部カタログを引いて書影を返す。ISBN が無い・該当なし・カタログの障害のときは書影なし（nil）とする。
func coverFor(ctx context.Context, catalog book.BookCatalog, isbn *book.ISBN) *book.Cover {
	if isbn == nil {
		return nil
	}
	entry, err := catalog.Lookup(ctx, *isbn)
	if err != nil {
		// A catalog outage must not block registering the book; saving again fetches the cover later.
		if !errors.Is(err, common.ErrNotFound) {
			slog.WarnContext(ctx, "book catalog lookup failed; saving without a cover", "isbn", isbn.String(), "error", err)
		}
		return nil
	}
	return entry.Cover
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
