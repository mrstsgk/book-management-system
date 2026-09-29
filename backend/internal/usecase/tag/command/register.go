package command

import (
	"context"

	"github.com/mrstsgk/book-management-system/backend/internal/domain/tag"
)

// RegisterCommand は RegisterUsecase の入力。データだけを持ち、ロジックは持たない。
type RegisterCommand struct {
	Name string
}

// TagView はタグの登録・改名の結果を返す Read Model。
type TagView struct {
	ID      int64
	Name    string
	Version int
}

type RegisterUsecase interface {
	Execute(ctx context.Context, cmd RegisterCommand) (*TagView, error)
}

type RegisterUsecaseImpl struct {
	Tags tag.Repository
}

// Execute はタグ名を検証してから登録する。
func (u *RegisterUsecaseImpl) Execute(ctx context.Context, cmd RegisterCommand) (*TagView, error) {
	name, err := tag.NewName(cmd.Name)
	if err != nil {
		return nil, err
	}
	t := tag.New(name)
	if err := u.Tags.Create(ctx, t); err != nil {
		return nil, err
	}
	return &TagView{ID: int64(t.ID), Name: t.Name.String(), Version: t.Version}, nil
}
