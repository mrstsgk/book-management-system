package command

import (
	"context"

	"github.com/mrstsgk/book-management-system/backend/internal/domainerr"
	tagdomain "github.com/mrstsgk/book-management-system/backend/internal/tag/domain"
)

// RegisterTagCommand は RegisterTagUsecase の入力。データだけを持ち、ロジックは持たない。
type RegisterTagCommand struct {
	// 追加するタグの名前。
	Name string
}

// TagView はタグの登録・改名の結果を返す Read Model。
type TagView struct {
	// タグを区別する番号。
	ID int64
	// 画面に出すタグの名前。
	Name string
	// 同時編集を見つけるための版。
	Version int
}

// RegisterTagUsecase は、分野タグを追加する。
type RegisterTagUsecase interface {
	Execute(ctx context.Context, cmd RegisterTagCommand) (*TagView, error)
}

type RegisterTagUsecaseImpl struct {
	Tags     tagdomain.Repository
	TagQuery tagdomain.Query
}

// Execute はタグ名を検証し、同名のタグが無いことを確かめてから登録する。
func (u *RegisterTagUsecaseImpl) Execute(ctx context.Context, cmd RegisterTagCommand) (*TagView, error) {
	name, err := parseUniqueName(ctx, u.TagQuery, cmd.Name, 0)
	if err != nil {
		return nil, err
	}
	t := tagdomain.New(name)
	// 同名のタグが同時に追加されたときは DB の一意制約が最後の砦（Repository が ErrConflict にする）
	if err := u.Tags.Create(ctx, t); err != nil {
		return nil, err
	}
	return &TagView{ID: int64(t.ID), Name: t.Name.String(), Version: t.Version}, nil
}

// parseUniqueName はタグ名を VO にし、同名のタグ（excludeID を除く）が無いことを確かめる。
// 違反は Violations（400）で返す。名前の形式が不正なら重複の検査はしない（前提が成立しないため）。
func parseUniqueName(ctx context.Context, q tagdomain.Query, raw string, excludeID tagdomain.ID) (tagdomain.Name, error) {
	var vs domainerr.Violations
	name, err := tagdomain.NewName(raw)
	if err != nil {
		vs.Add("name", "invalid")
		return tagdomain.Name{}, vs
	}
	dup, err := q.ExistsByName(ctx, name.String(), excludeID)
	if err != nil {
		return tagdomain.Name{}, err
	}
	if dup {
		vs.Add("name", "tagNameDuplicated")
		return tagdomain.Name{}, vs
	}
	return name, nil
}
