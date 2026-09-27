package query

import (
	"context"
	"fmt"

	"github.com/mrstsgk/book-management-system/backend/internal/domain/author"
	"github.com/mrstsgk/book-management-system/backend/internal/domain/book"
	"github.com/mrstsgk/book-management-system/backend/internal/domain/common"
)

type ListByAuthorUsecase interface {
	Execute(ctx context.Context, authorID int64) ([]*book.BookSummary, error)
}

type ListByAuthorUsecaseImpl struct {
	Authors author.Query
	Books   book.Query
}

func (u *ListByAuthorUsecaseImpl) Execute(ctx context.Context, authorID int64) ([]*book.BookSummary, error) {
	exists, err := u.Authors.Exists(ctx, author.ID(authorID))
	if err != nil {
		return nil, err
	}
	// An unknown author is 404, while an author without books is an empty list.
	if !exists {
		return nil, fmt.Errorf("%w: 著者が見つかりません", common.ErrNotFound)
	}
	return u.Books.FindSummariesByAuthorID(ctx, author.ID(authorID))
}
