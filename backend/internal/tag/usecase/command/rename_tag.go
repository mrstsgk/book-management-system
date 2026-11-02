package command

import (
	"context"

	tagdomain "github.com/mrstsgk/book-management-system/backend/internal/tag/domain"
)

// RenameTagCommand は RenameTagUsecase の入力。データだけを持ち、ロジックは持たない。
type RenameTagCommand struct {
	// 改名するタグの番号。
	ID int64
	// 新しい名前。
	Name string
	// 改名を始めたときに読んだ版（楽観的ロックに使う）。
	Version int
}

// RenameTagUsecase は、分野タグの名前を変える。
type RenameTagUsecase interface {
	Execute(ctx context.Context, cmd RenameTagCommand) (*TagView, error)
}

type RenameTagUsecaseImpl struct {
	Tags     tagdomain.Repository
	TagQuery tagdomain.Query
}

// Execute はタグの名前を検証し、自分以外に同名のタグが無いことを確かめてから差し替える。
func (u *RenameTagUsecaseImpl) Execute(ctx context.Context, cmd RenameTagCommand) (*TagView, error) {
	t, err := u.Tags.FindByID(ctx, tagdomain.ID(cmd.ID))
	if err != nil {
		return nil, err
	}
	name, err := parseUniqueName(ctx, u.TagQuery, cmd.Name, t.ID)
	if err != nil {
		return nil, err
	}
	t.Rename(name, cmd.Version)
	// version の不一致は Repository が ErrConflict にする（409）
	if err := u.Tags.Update(ctx, t); err != nil {
		return nil, err
	}
	return &TagView{ID: int64(t.ID), Name: t.Name.String(), Version: t.Version}, nil
}
