package command

import (
	"context"
	"errors"
	"log/slog"

	"github.com/mrstsgk/book-management-system/backend/internal/domain/book"
	"github.com/mrstsgk/book-management-system/backend/internal/domain/common"
)

type FillMissingCoversUsecase interface {
	Execute(ctx context.Context) error
}

type FillMissingCoversUsecaseImpl struct {
	Books   book.Repository
	Catalog book.BookCatalog
}

// Execute は書影の無い本を外部カタログに問い合わせ、書影が見つかった本だけ保存する（Google Books のキーを
// 後から設定しても、見本の本などに書影が付くようにするため）。1冊の失敗で残りを止めない。
func (u *FillMissingCoversUsecaseImpl) Execute(ctx context.Context) error {
	targets, err := u.Books.FindCoverless(ctx)
	if err != nil {
		return err
	}
	for _, b := range targets {
		if ctx.Err() != nil {
			// シャットダウンによる通常のキャンセルなので失敗にしない（誤解を招く warn ログを出さないため）
			return nil //nolint:nilerr
		}
		u.fill(ctx, b)
	}
	return nil
}

// fill は1冊分の書影を探し、見つかったときだけ保存する（版を無駄に進めると、同時に編集している画面が競合するため）。
func (u *FillMissingCoversUsecaseImpl) fill(ctx context.Context, b *book.Book) {
	entry, err := u.Catalog.Lookup(ctx, b.ISBN)
	if err != nil {
		if !errors.Is(err, common.ErrNotFound) {
			slog.WarnContext(ctx, "cover lookup failed; leaving the book without a cover", "isbn", b.ISBN.String(), "error", err)
		}
		return
	}
	if entry.Cover == nil {
		return
	}
	b.RefreshCatalog(entry.Bibliography, entry.Cover)
	if err := u.Books.Update(ctx, b); err != nil {
		if errors.Is(err, common.ErrConflict) {
			// その間に本が更新され、そちらで書影も取り直されている
			slog.InfoContext(ctx, "book was updated while filling its cover; skipping", "isbn", b.ISBN.String())
			return
		}
		slog.WarnContext(ctx, "failed to save the filled cover", "isbn", b.ISBN.String(), "error", err)
	}
}
