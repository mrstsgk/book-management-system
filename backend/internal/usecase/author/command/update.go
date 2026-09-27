package command

import (
	"context"
	"time"

	"github.com/mrstsgk/book-management-system/backend/internal/domain/author"
)

// UpdateCommand is the input contract for UpdateUsecase (the boundary
// crossed from Presentation). It carries data only — no logic.
type UpdateCommand struct {
	ID        int64
	Name      string
	BirthDate *time.Time
	Version   int
}

type UpdateUsecase interface {
	Execute(ctx context.Context, cmd UpdateCommand) (*author.Author, error)
}

type UpdateUsecaseImpl struct {
	Authors author.Repository
	Now     func() time.Time
}

func (u *UpdateUsecaseImpl) Execute(ctx context.Context, cmd UpdateCommand) (*author.Author, error) {
	a, err := u.Authors.FindByID(ctx, author.ID(cmd.ID))
	if err != nil {
		return nil, err
	}
	name, err := author.NewName(cmd.Name)
	if err != nil {
		return nil, err
	}
	birthDate, err := newBirthDate(cmd.BirthDate, u.Now)
	if err != nil {
		return nil, err
	}
	a.Name = name
	a.BirthDate = birthDate
	a.Version = cmd.Version
	if err := u.Authors.Update(ctx, a); err != nil {
		return nil, err
	}
	return a, nil
}
