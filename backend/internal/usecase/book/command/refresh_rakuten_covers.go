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
	targets, err := u.Books.FindCoverRefreshTargets(ctx, now.Add(-book.RakutenRefreshAfter))
	if err != nil {
		return err
	}
	for _, b := range targets {
		if ctx.Err() != nil {
			// シャットダウンによる通常のキャンセルなので失敗として返さない（呼び出し側が「取り直しに
			// 失敗した」という誤解を招くwarnログを出さないため）。残りの対象を無駄に処理し続けない
			// （1冊ずつwarnログが増えるだけになるため）だけが目的。
			return nil //nolint:nilerr
		}
		u.refresh(ctx, b, now)
	}
	return nil
}

// refresh は1冊分の楽天の書影を取り直し、変わったときだけ保存する。
func (u *RefreshRakutenCoversUsecaseImpl) refresh(ctx context.Context, b *book.Book, now time.Time) {
	entry, err := u.Catalog.Lookup(ctx, b.ISBN)
	switch {
	case err == nil && entry.Cover != nil:
		b.RefreshCatalog(entry.Bibliography, entry.Cover)
	case err == nil:
		// 該当はあったが書影が無い。書影を取り直せなかったのと同じ扱いにし、期限切れのときだけ外す
		// （今の楽天の書影を無条件にnilへ差し替えると、期限前のものまで消してしまうため）。
		if !b.DropExpiredCover(now) {
			return
		}
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
