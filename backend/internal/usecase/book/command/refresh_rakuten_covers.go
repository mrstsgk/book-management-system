package command

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/mrstsgk/book-management-system/backend/internal/domain/book"
	"github.com/mrstsgk/book-management-system/backend/internal/domain/common"
)

type RefreshRakutenCoversUsecase interface {
	Execute(ctx context.Context) error
}

type RefreshRakutenCoversUsecaseImpl struct {
	Books   book.Repository
	Catalog book.BookCatalog
	Now     func() time.Time
}

// Execute は取り直す時期に入った楽天の書影を取り直し、取り直せないまま保持期限を過ぎた書影は外す。
// 1冊の失敗で残りを止めない（楽天の規約上、期限を過ぎた本を放置しないことを優先するため）。
func (u *RefreshRakutenCoversUsecaseImpl) Execute(ctx context.Context) error {
	now := u.Now()
	targets, err := u.Books.FindRakutenRefreshTargets(ctx, now.Add(-book.RakutenRefreshAfter))
	if err != nil {
		return err
	}
	for _, b := range targets {
		u.refresh(ctx, b, now)
	}
	return nil
}

// refresh は1冊分の楽天の書影を取り直し、変わったときだけ保存する。
func (u *RefreshRakutenCoversUsecaseImpl) refresh(ctx context.Context, b *book.Book, now time.Time) {
	entry, err := u.Catalog.Lookup(ctx, b.ISBN)
	switch {
	case err == nil:
		b.RefreshCatalog(entry.Bibliography, entry.Cover)
	case b.DropExpiredCover(now):
		slog.WarnContext(ctx, "rakuten cover expired before it could be refreshed; dropping it", "isbn", b.ISBN.String(), "error", err)
	default:
		// 期限前なら保存せず次の契機を待つ（version を進めると、同時に編集している画面が無駄に競合するため）
		slog.WarnContext(ctx, "rakuten cover refresh failed; retrying later", "isbn", b.ISBN.String(), "error", err)
		return
	}
	if err := u.Books.Update(ctx, b); err != nil {
		if errors.Is(err, common.ErrConflict) {
			// その間に本が更新され、そちらで書影も取り直されている
			slog.InfoContext(ctx, "book was updated during the rakuten cover refresh; skipping", "isbn", b.ISBN.String())
			return
		}
		slog.WarnContext(ctx, "failed to save the refreshed rakuten cover", "isbn", b.ISBN.String(), "error", err)
	}
}
