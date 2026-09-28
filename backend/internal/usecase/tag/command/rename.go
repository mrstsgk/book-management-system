package command

import (
	"context"

	"github.com/mrstsgk/book-management-system/backend/internal/domain/tag"
)

// RenameCommand は RenameUsecase の入力。データだけを持ち、ロジックは持たない。
type RenameCommand struct {
	ID      int64
	Name    string
	Version int
}

type RenameUsecase interface {
	Execute(ctx context.Context, cmd RenameCommand) (*TagView, error)
}

type RenameUsecaseImpl struct {
	Tags tag.Repository
}

// Execute はタグの名前を検証してから差し替える。
func (u *RenameUsecaseImpl) Execute(ctx context.Context, cmd RenameCommand) (*TagView, error) {
	t, err := u.Tags.FindByID(ctx, tag.ID(cmd.ID))
	if err != nil {
		return nil, err
	}
	name, err := tag.NewName(cmd.Name)
	if err != nil {
		return nil, err
	}
	t.Rename(name, cmd.Version)
	if err := u.Tags.Update(ctx, t); err != nil {
		return nil, err
	}
	return &TagView{ID: int64(t.ID), Name: t.Name.String(), Version: t.Version}, nil
}
