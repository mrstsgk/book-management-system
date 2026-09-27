package command

import (
	"context"

	"github.com/mrstsgk/book-management-system/backend/internal/domain/author"
	"github.com/mrstsgk/book-management-system/backend/internal/domain/book"
)

// CreateCommand is the input contract for CreateUsecase (the boundary
// crossed from Presentation). It carries data only — no logic.
type CreateCommand struct {
	Title     string
	Price     int64
	AuthorIDs []int64
	Status    int
}

type CreateUsecase interface {
	Execute(ctx context.Context, cmd CreateCommand) (*book.BookDetail, error)
}

type CreateUsecaseImpl struct {
	Books   book.Repository
	Authors author.Repository
	Details book.Query
}

func (u *CreateUsecaseImpl) Execute(ctx context.Context, cmd CreateCommand) (*book.BookDetail, error) {
	c, err := newContents(cmd.Title, cmd.Price, cmd.AuthorIDs, cmd.Status)
	if err != nil {
		return nil, err
	}
	b, err := book.New(c.title, c.price, c.authorIDs, c.status)
	if err != nil {
		return nil, err
	}
	if err := ensureAuthorsExist(ctx, u.Authors, b.AuthorIDs); err != nil {
		return nil, err
	}
	if err := u.Books.Create(ctx, b); err != nil {
		return nil, err
	}
	return u.Details.FindDetailByID(ctx, b.ID)
}
