package command

import (
	"context"
	"errors"

	"github.com/mrstsgk/book-management-system/backend/internal/domain/book"
	"github.com/mrstsgk/book-management-system/backend/internal/domain/common"
)

type DisableRakutenUsecase interface {
	Execute(ctx context.Context, id int64) error
}

type DisableRakutenUsecaseImpl struct {
	Books book.Repository
}

// Execute は楽天由来の情報を消して無効化する。読んだ直後の version で保存し、競合したら1回だけ読み直してやり直す。
// version を受け取らないのは、楽天からの削除の指示には画面で最後に読んだ内容に関係なく従う必要があるため。
func (u *DisableRakutenUsecaseImpl) Execute(ctx context.Context, id int64) error {
	for attempt := 0; ; attempt++ {
		b, err := u.Books.FindByID(ctx, book.ID(id))
		if err != nil {
			return err
		}
		if b.RakutenDisabled && (b.Cover == nil || b.Cover.Source() != book.CoverSourceRakuten) {
			return nil
		}
		b.DisableRakuten()
		err = u.Books.Update(ctx, b)
		if errors.Is(err, common.ErrConflict) && attempt == 0 {
			continue
		}
		return err
	}
}
