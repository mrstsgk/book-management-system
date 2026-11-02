package command

import (
	"context"
	"errors"
	"log/slog"

	"github.com/mrstsgk/book-management-system/backend/internal/domain/book"
	"github.com/mrstsgk/book-management-system/backend/internal/domain/common"
	"github.com/mrstsgk/book-management-system/backend/internal/domain/tag"
)

// UpdateCommand は UpdateUsecase の入力（Presentation から渡る境界）。データだけを持ち、ロジックは持たない。
type UpdateCommand struct {
	ID      int64
	Summary string
	// TitleOverride は自分で上書きする書名。全体の置き換えなので、前後の空白を除いて空なら上書きを外す。
	TitleOverride string
	TagIDs        []int64
	Comment       string
	Rating        int
	Version       int
}

// UpdateUsecase は、読んだ本の感想・評価・書名の上書き・分野タグを更新する。
type UpdateUsecase interface {
	Execute(ctx context.Context, cmd UpdateCommand) (*book.BookDetail, error)
}

type UpdateUsecaseImpl struct {
	Books   book.Repository
	Catalog book.BookCatalog
	Details book.Query
	Tags    tag.Query
}

// Execute は入力の違反（VO・相関ルール）をすべて集めてから、感想と評価を差し替え、あわせて書誌と書影を外部カタログから取り直して保存する。
func (u *UpdateUsecaseImpl) Execute(ctx context.Context, cmd UpdateCommand) (*book.BookDetail, error) {
	b, err := u.Books.FindByID(ctx, book.ID(cmd.ID))
	if err != nil {
		return nil, err
	}
	var vs common.Violations
	review := parseReviewFields(&vs, cmd.Summary, cmd.TitleOverride, cmd.Comment, cmd.Rating)
	tags, err := parseTagSelection(ctx, &vs, u.Tags, cmd.TagIDs)
	if err != nil {
		return nil, err
	}
	if err := vs.Err(); err != nil {
		return nil, err
	}
	b.ChangeReview(review.summary, review.comment, review.rating, tags, cmd.Version)
	b.OverrideTitle(review.override)
	u.refreshCatalog(ctx, b)
	// version の不一致は Repository が ErrConflict にする（409）
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
