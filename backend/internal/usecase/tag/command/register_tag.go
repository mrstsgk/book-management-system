package command

import (
	"context"

	"github.com/mrstsgk/book-management-system/backend/internal/domain/common"
	"github.com/mrstsgk/book-management-system/backend/internal/domain/tag"
)

// RegisterTagCommand は RegisterTagUsecase の入力。データだけを持ち、ロジックは持たない。
type RegisterTagCommand struct {
	Name string
}

// TagView はタグの登録・改名の結果を返す Read Model。
type TagView struct {
	ID      int64
	Name    string
	Version int
}

// RegisterTagUsecase は、分野タグを追加する。
type RegisterTagUsecase interface {
	Execute(ctx context.Context, cmd RegisterTagCommand) (*TagView, error)
}

type RegisterTagUsecaseImpl struct {
	Tags     tag.Repository
	TagQuery tag.Query
}

// Execute はタグ名を検証し、同名のタグが無いことを確かめてから登録する。
func (u *RegisterTagUsecaseImpl) Execute(ctx context.Context, cmd RegisterTagCommand) (*TagView, error) {
	name, err := parseUniqueName(ctx, u.TagQuery, cmd.Name, 0)
	if err != nil {
		return nil, err
	}
	t := tag.New(name)
	// 同名のタグが同時に追加されたときは DB の一意制約が最後の砦（Repository が ErrConflict にする）
	if err := u.Tags.Create(ctx, t); err != nil {
		return nil, err
	}
	return &TagView{ID: int64(t.ID), Name: t.Name.String(), Version: t.Version}, nil
}

// parseUniqueName はタグ名を VO にし、同名のタグ（excludeID を除く）が無いことを確かめる。
// 違反は Violations（400）で返す。名前の形式が不正なら重複の検査はしない（前提が成立しないため）。
func parseUniqueName(ctx context.Context, q tag.Query, raw string, excludeID tag.ID) (tag.Name, error) {
	var vs common.Violations
	name, err := tag.NewName(raw)
	if err != nil {
		vs.Add("name", "invalid")
		return tag.Name{}, vs
	}
	dup, err := q.ExistsByName(ctx, name.String(), excludeID)
	if err != nil {
		return tag.Name{}, err
	}
	if dup {
		vs.Add("name", "tagNameDuplicated")
		return tag.Name{}, vs
	}
	return name, nil
}
