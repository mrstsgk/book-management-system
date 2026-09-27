package command

import (
	"context"
	"time"

	"github.com/mrstsgk/book-management-system/backend/internal/domain/author"
)

// CreateCommand is the input contract for CreateUsecase (the boundary
// crossed from Presentation). It carries data only — no logic.
type CreateCommand struct {
	Name      string
	BirthDate *time.Time
}

type CreateUsecase interface {
	Execute(ctx context.Context, cmd CreateCommand) (*author.Author, error)
}

type CreateUsecaseImpl struct {
	Authors author.Repository
	Now     func() time.Time
}

func (u *CreateUsecaseImpl) Execute(ctx context.Context, cmd CreateCommand) (*author.Author, error) {
	name, err := author.NewName(cmd.Name)
	if err != nil {
		return nil, err
	}
	birthDate, err := newBirthDate(cmd.BirthDate, u.Now)
	if err != nil {
		return nil, err
	}
	a := author.New(name, birthDate)
	if err := u.Authors.Create(ctx, a); err != nil {
		return nil, err
	}
	return a, nil
}
