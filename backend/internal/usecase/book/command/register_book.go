package command

import (
	"context"
	"errors"

	"github.com/mrstsgk/book-management-system/backend/internal/domain/book"
	"github.com/mrstsgk/book-management-system/backend/internal/domain/common"
	"github.com/mrstsgk/book-management-system/backend/internal/domain/tag"
)

// RegisterBookCommand は RegisterBookUsecase の入力（Presentation から渡る境界）。データだけを持ち、ロジックは持たない。
type RegisterBookCommand struct {
	ISBN    string
	Summary string
	// TitleOverride は自分で上書きする書名。前後の空白を除いて空なら上書きしない。
	TitleOverride string
	TagIDs        []int64
	Comment       string
	Rating        int
}

// RegisterBookUsecase は、読んだ本を登録する。
type RegisterBookUsecase interface {
	Execute(ctx context.Context, cmd RegisterBookCommand) (*book.BookDetail, error)
}

type RegisterBookUsecaseImpl struct {
	Books   book.Repository
	Catalog book.BookCatalog
	Details book.Query
	Tags    tag.Query
}

// Execute は入力の違反（VO・相関ルール）をすべて集めてから、ISBN で外部カタログの書誌・書影を取得し、感想・評価と合わせて登録する。
func (u *RegisterBookUsecaseImpl) Execute(ctx context.Context, cmd RegisterBookCommand) (*book.BookDetail, error) {
	var vs common.Violations
	isbn, err := book.NewISBN(cmd.ISBN)
	isbnOK := err == nil
	if !isbnOK {
		vs.Add("isbn", "format")
	}
	review := parseReviewFields(&vs, cmd.Summary, cmd.TitleOverride, cmd.Comment, cmd.Rating)
	tags, err := parseTagSelection(ctx, &vs, u.Tags, cmd.TagIDs)
	if err != nil {
		return nil, err
	}
	// ISBN の形式が不正なら重複の検査はしない（前提が成立しないため）
	if isbnOK {
		dup, err := u.Details.ExistsByISBN(ctx, isbn)
		if err != nil {
			return nil, err
		}
		if dup {
			vs.Add("isbn", "isbnDuplicated")
		}
	}
	if err := vs.Err(); err != nil {
		return nil, err
	}
	// 外部カタログは他の違反が 1 つも無いときだけ問い合わせる（外部 API を無駄に呼ばないため。
	// その結果の isbnNotInCatalog だけは 2 回目の送信で分かることがある）
	entry, err := u.Catalog.Lookup(ctx, isbn)
	if errors.Is(err, common.ErrNotFound) {
		// 書誌が無いと書名すら分からないので、登録できない入力として扱う
		vs.Add("isbn", "isbnNotInCatalog")
		return nil, vs
	}
	if err != nil {
		return nil, err
	}
	b := book.New(isbn, entry.Bibliography, entry.Cover, review.summary, review.comment, review.rating, tags)
	b.OverrideTitle(review.override)
	// 同じ ISBN が同時に登録されたときは DB の一意制約が最後の砦（Repository が ErrConflict にする）
	if err := u.Books.Create(ctx, b); err != nil {
		return nil, err
	}
	return u.Details.FindDetailByID(ctx, b.ID)
}
