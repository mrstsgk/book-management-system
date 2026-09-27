package command

import (
	"context"

	"github.com/mrstsgk/book-management-system/backend/internal/domain/author"
	"github.com/mrstsgk/book-management-system/backend/internal/domain/book"
)

// UpdateCommand is the input contract for UpdateUsecase (the boundary
// crossed from Presentation). It carries data only — no logic.
type UpdateCommand struct {
	ID        int64
	Title     string
	Price     int64
	AuthorIDs []int64
	Status    int
	AmazonURL *string
	Version   int
}

type UpdateUsecase interface {
	Execute(ctx context.Context, cmd UpdateCommand) (*book.BookDetail, error)
}

type UpdateUsecaseImpl struct {
	Books   book.Repository
	Authors author.Repository
	Details book.Query
	Images  book.ImageStorage
}

func (u *UpdateUsecaseImpl) Execute(ctx context.Context, cmd UpdateCommand) (*book.BookDetail, error) {
	b, err := u.Books.FindByID(ctx, book.ID(cmd.ID))
	if err != nil {
		return nil, err
	}
	c, err := newContents(cmd.Title, cmd.Price, cmd.AuthorIDs, cmd.Status, cmd.AmazonURL)
	if err != nil {
		return nil, err
	}
	if err := b.Change(c.title, c.price, c.authorIDs, c.status, cmd.Version); err != nil {
		return nil, err
	}
	// Omitting the URL on update clears it, matching how the other fields are replaced wholesale (PUT).
	b.ChangeAmazonURL(c.amazonURL)
	if err := ensureAuthorsExist(ctx, u.Authors, b.AuthorIDs); err != nil {
		return nil, err
	}
	if err := u.Books.Update(ctx, b); err != nil {
		return nil, err
	}
	return detailOf(ctx, u.Details, u.Images, b.ID)
}
