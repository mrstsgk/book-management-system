package command

import (
	"context"
	"errors"
	"fmt"

	"github.com/mrstsgk/book-management-system/backend/internal/domain/book"
	"github.com/mrstsgk/book-management-system/backend/internal/domain/common"
)

// RegisterCommand は RegisterUsecase の入力（Presentation から渡る境界）。データだけを持ち、ロジックは持たない。
type RegisterCommand struct {
	ISBN    string
	Comment string
	Rating  int
}

type RegisterUsecase interface {
	Execute(ctx context.Context, cmd RegisterCommand) (*book.BookDetail, error)
}

type RegisterUsecaseImpl struct {
	Books   book.Repository
	Catalog book.BookCatalog
	Details book.Query
}

// Execute は入力を検証してから ISBN で外部カタログの書誌・書影を取得し、感想・評価と合わせて登録する。
func (u *RegisterUsecaseImpl) Execute(ctx context.Context, cmd RegisterCommand) (*book.BookDetail, error) {
	isbn, err := book.NewISBN(cmd.ISBN)
	if err != nil {
		return nil, err
	}
	comment, err := book.NewComment(cmd.Comment)
	if err != nil {
		return nil, err
	}
	rating, err := book.NewRating(cmd.Rating)
	if err != nil {
		return nil, err
	}
	// 入力が不正なときに外部カタログを呼ばないよう、検証を先に済ませてから問い合わせる
	entry, err := u.Catalog.Lookup(ctx, isbn)
	if errors.Is(err, common.ErrNotFound) {
		// 書誌が無いと書名すら分からないので、登録できない入力として扱う
		return nil, fmt.Errorf("%w: ISBN %s の本が外部カタログに見つかりません", common.ErrInvalid, isbn)
	}
	if err != nil {
		return nil, err
	}
	b := book.New(isbn, entry.Bibliography, entry.Cover, comment, rating)
	if err := u.Books.Create(ctx, b); err != nil {
		return nil, err
	}
	return u.Details.FindDetailByID(ctx, b.ID)
}
