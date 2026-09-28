package command

import (
	"context"
	"errors"
	"log/slog"

	"github.com/mrstsgk/book-management-system/backend/internal/domain/book"
	"github.com/mrstsgk/book-management-system/backend/internal/domain/common"
)

// UpdateCommand は UpdateUsecase の入力（Presentation から渡る境界）。データだけを持ち、ロジックは持たない。
type UpdateCommand struct {
	ID      int64
	Comment string
	Rating  int
	Version int
}

type UpdateUsecase interface {
	Execute(ctx context.Context, cmd UpdateCommand) (*book.BookDetail, error)
}

type UpdateUsecaseImpl struct {
	Books   book.Repository
	Catalog book.BookCatalog
	Details book.Query
}

// Execute は感想と評価を差し替え、あわせて書誌と書影を外部カタログから取り直して保存する。
func (u *UpdateUsecaseImpl) Execute(ctx context.Context, cmd UpdateCommand) (*book.BookDetail, error) {
	b, err := u.Books.FindByID(ctx, book.ID(cmd.ID))
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
	b.ChangeReview(comment, rating, cmd.Version)
	u.refreshCatalog(ctx, b)
	if err := u.Books.Update(ctx, b); err != nil {
		return nil, err
	}
	return u.Details.FindDetailByID(ctx, b.ID)
}

// refreshCatalog は保存のたびに書誌と書影を取り直す（openBD の規約上、提供元の変更をできるだけ早く反映するため）。
// カタログの障害や該当なしで感想の更新を止めないよう、取り直せなければ今の書誌と書影のままにする。
func (u *UpdateUsecaseImpl) refreshCatalog(ctx context.Context, b *book.Book) {
	entry, err := u.Catalog.Lookup(ctx, b.ISBN)
	if err != nil {
		if !errors.Is(err, common.ErrNotFound) {
			slog.WarnContext(ctx, "book catalog lookup failed; keeping the current bibliography", "isbn", b.ISBN.String(), "error", err)
		}
		return
	}
	b.RefreshCatalog(entry.Bibliography, entry.Cover)
}
